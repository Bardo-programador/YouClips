package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	_ "github.com/mattn/go-sqlite3"
	"youclips/internal/controller"
	"youclips/internal/repository"
	"youclips/internal/service"
)

const (
	defaultDBPath       = "./youclips.db"
	defaultStorageDir   = "./storage/clips"
	defaultClipTTL      = 5 // minutes
	SERVER_ADDR         = ":8080"
	SERVER_URL					= "http://localhost"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := initDatabase(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	storageDir := os.Getenv("STORAGE_DIR")
	if storageDir == "" {
		storageDir = defaultStorageDir
	}

	clipTTL := defaultClipTTL
	if ttlEnv := os.Getenv("CLIP_TTL_MINUTES"); ttlEnv != "" {
		if ttl, err := strconv.Atoi(ttlEnv); err == nil && ttl > 0 {
			clipTTL = ttl
		}
	}
	clipTTLDuration := time.Duration(clipTTL) * time.Minute

	clipRepo := repository.NewSQLiteClipRepository(db)
	metadataRepo := repository.NewSQLiteMetadataRepository(db)
	processor := service.NewYTDLPProcessor(clipRepo, metadataRepo, storageDir, clipTTLDuration)
	clipService := service.NewClipService(clipRepo, processor)
	clipHandler := controller.NewClipHandler(clipService)

	// Start cleanup worker to remove expired clips
	cleanupWorker := service.NewCleanupWorker(clipRepo, storageDir, 1*time.Minute)
	go cleanupWorker.Start(context.Background())

	http.HandleFunc("/clips", clipHandler.Clips)
	http.HandleFunc("/clips/", clipHandler.ClipByID)
	http.HandleFunc("/metadata", clipHandler.ClipMetaData)

	log.Printf("Starting server on %s%s with DB at %s, storage at %s, clip TTL %d minutes", SERVER_URL, SERVER_ADDR, dbPath, storageDir, clipTTL)
	if err := http.ListenAndServe(SERVER_ADDR, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initDatabase(db *sql.DB) error {
	// Check if clips table exists
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='clips'").Scan(&tableName)
	tableExists := err == nil

	if !tableExists {
		// Create table with all columns including expires_at
		schema := `
		CREATE TABLE clips (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			start_time INTEGER NOT NULL,
			end_time INTEGER NOT NULL,
			duration_seconds INTEGER NOT NULL,
			format TEXT NOT NULL,
			size INTEGER NOT NULL DEFAULT 0,
			file_path TEXT NOT NULL DEFAULT '',
			original_url TEXT NOT NULL,
			status TEXT NOT NULL,
			expires_at DATETIME
		);
		CREATE INDEX idx_clips_status ON clips(status);
		CREATE INDEX idx_clips_created_at ON clips(created_at DESC);
		CREATE INDEX idx_clips_expires_at ON clips(expires_at);
		`
		if _, err := db.Exec(schema); err != nil {
			return err
		}
	} else {
		// Table exists, check if expires_at column exists
		var columnExists bool
		err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('clips') WHERE name='expires_at'").Scan(&columnExists)
		if err == nil && !columnExists {
			// Add expires_at column
			if _, err := db.Exec("ALTER TABLE clips ADD COLUMN expires_at DATETIME"); err != nil {
				return err
			}
			log.Println("Added expires_at column to clips table")
		}
		
		// Ensure indexes exist
		db.Exec("CREATE INDEX IF NOT EXISTS idx_clips_status ON clips(status)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_clips_created_at ON clips(created_at DESC)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_clips_expires_at ON clips(expires_at)")
	}

	// Create video_metadata table
	metadataSchema := `
	CREATE TABLE IF NOT EXISTS video_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		title TEXT NOT NULL,
		duration INTEGER NOT NULL,
		created_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_video_metadata_url ON video_metadata(url);
	`
	if _, err := db.Exec(metadataSchema); err != nil {
		return err
	}

	return nil
}
