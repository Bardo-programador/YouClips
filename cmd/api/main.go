package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"youclips/internal/controller"
	"youclips/internal/repository"
	"youclips/internal/service"
)

const (
	defaultDBPath     = "./youclips.db"
	defaultStorageDir = "./storage/clips"
	SERVER_ADDR    		= ":8080"
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

	clipRepo := repository.NewSQLiteClipRepository(db)
	processor := service.NewYTDLPProcessor(clipRepo, storageDir)
	clipService := service.NewClipService(clipRepo, processor)
	clipHandler := controller.NewClipHandler(clipService)

	http.HandleFunc("/clips", clipHandler.Clips)
	http.HandleFunc("/clips/", clipHandler.ClipByID)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(SERVER_ADDR, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initDatabase(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS clips (
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
		status TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_clips_status ON clips(status);
	CREATE INDEX IF NOT EXISTS idx_clips_created_at ON clips(created_at DESC);
	`
	_, err := db.Exec(schema)
	return err
}
