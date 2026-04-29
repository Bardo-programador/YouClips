package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"youclips/internal/controller"
	"youclips/internal/repository"
	"youclips/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultStorageDir   = "./storage/clips"
	defaultClipTTL      = 5 // minutes
	SERVER_ADDR         = ":8080"
	SERVER_URL					= "http://localhost"
	USER								= "user"
	PASSWORD						= "password"
	DATABASSE_NAME 			= "youclips"
	DATABASE_URL				= "postgres://user:password@localhost/youclips?sslmode=disable"
)

func configPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 10 * time.Second
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = DATABASE_URL
	}

	pool, err := configPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}
	defer pool.Close()

	if err := repository.InitPostgresDatabase(ctx, pool); err != nil {
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

	clipRepo := repository.NewPostgresClipRepository(pool)
	metadataRepo := repository.NewPostgresMetadataRepository(pool)
	processor := service.NewYTDLPProcessor(clipRepo, metadataRepo, storageDir, clipTTLDuration)
	clipService := service.NewClipService(clipRepo, processor)
	clipHandler := controller.NewClipHandler(clipService)

	// Start cleanup worker to remove expired clips
	cleanupWorker := service.NewCleanupWorker(clipRepo, storageDir, 1*time.Minute)
	go cleanupWorker.Start(ctx)

	// Wrap handlers with CORS middleware
	// http.HandleFunc("/clips", corsMiddleware(clipHandler.Clips))
	http.HandleFunc("/clips/", corsMiddleware(clipHandler.ClipByID))
	http.HandleFunc("/metadata", corsMiddleware(clipHandler.ClipMetaData))

	log.Printf("Starting server on %s%s with storage at %s, clip TTL %d minutes", SERVER_URL, SERVER_ADDR, storageDir, clipTTL)
	if err := http.ListenAndServe(SERVER_ADDR, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
