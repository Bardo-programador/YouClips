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
	defaultStorageDir       = "./storage/clips"
	defaultClipTTL          = "5" // minutes
	defaultServerAddr       = ":8080"
	defaultServerURL        = "http://localhost"
	defaultDatabaseURL      = "postgres://user:password@localhost/youclips?sslmode=disable"
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

	// Load environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = defaultDatabaseURL
	}

	log.Printf("Using DATABASE_URL: %s", dbURL)

	STORAGE_DIR := os.Getenv("STORAGE_DIR")
	if STORAGE_DIR == "" {
		STORAGE_DIR = defaultStorageDir
	}

	CLIP_TTL := os.Getenv("CLIP_TTL_MINUTES")
	if CLIP_TTL == "" {
		CLIP_TTL = defaultClipTTL
	}

	// Get PORT from environment (Render sets this), fallback to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := ":" + port

	ctx := context.Background()

	pool, err := configPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}
	defer pool.Close()

	if err := repository.InitPostgresDatabase(ctx, pool); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if STORAGE_DIR == "" {
		STORAGE_DIR = defaultStorageDir
	}

	if err := os.MkdirAll(STORAGE_DIR, os.ModePerm); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}
	
	if CLIP_TTL == "" {
		CLIP_TTL = defaultClipTTL
	}
	clipTTL_int , err := strconv.Atoi(CLIP_TTL)
	clipTTLDuration := time.Duration(clipTTL_int) * time.Minute

	clipRepo := repository.NewPostgresClipRepository(pool)
	metadataRepo := repository.NewPostgresMetadataRepository(pool)
	processor := service.NewYTDLPProcessor(clipRepo, metadataRepo, STORAGE_DIR, clipTTLDuration)
	clipService := service.NewClipService(clipRepo, processor)
	clipHandler := controller.NewClipHandler(clipService)

	// Start cleanup worker to remove expired clips
	cleanupWorker := service.NewCleanupWorker(clipRepo, STORAGE_DIR, 1*time.Minute)
	go cleanupWorker.Start(ctx)

	// Wrap handlers with CORS middleware
	// http.HandleFunc("/clips", corsMiddleware(clipHandler.Clips))
	http.HandleFunc("/clips/", corsMiddleware(clipHandler.ClipByID))
	http.HandleFunc("/metadata", corsMiddleware(clipHandler.ClipMetaData))
	
	log.Printf("Starting server on %s with storage at %s, clip TTL %s minutes", serverAddr, STORAGE_DIR, CLIP_TTL)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
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
