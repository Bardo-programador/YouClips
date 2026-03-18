package service

import (
	"context"
	"log"
	"os"
	"time"
	"youclips/internal/entities"
	"youclips/internal/repository"
)

type CleanupWorker struct {
	repo       repository.ClipRepository
	storageDir string
	interval   time.Duration
}

func NewCleanupWorker(repo repository.ClipRepository, storageDir string, interval time.Duration) *CleanupWorker {
	return &CleanupWorker{
		repo:       repo,
		storageDir: storageDir,
		interval:   interval,
	}
}

func (w *CleanupWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("Cleanup worker started (checking every %v)", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Cleanup worker stopped")
			return
		case <-ticker.C:
			w.cleanupExpiredClips()
		}
	}
}

func (w *CleanupWorker) cleanupExpiredClips() {
	ctx := context.Background()
	
	expiredClips, err := w.repo.FindExpiredClips(ctx)
	if err != nil {
		log.Printf("Error finding expired clips: %v", err)
		return
	}

	if len(expiredClips) == 0 {
		return
	}

	log.Printf("Found %d expired clips to cleanup", len(expiredClips))

	for _, clip := range expiredClips {
		if err := w.cleanupClip(ctx, clip); err != nil {
			log.Printf("Error cleaning up clip %d: %v", clip.ID, err)
		} else {
			log.Printf("Successfully cleaned up clip %d (%s)", clip.ID, clip.Title)
		}
	}
}

func (w *CleanupWorker) cleanupClip(ctx context.Context, clip *entities.Clip) error {
	if clip.FilePath != "" {
		if err := os.Remove(clip.FilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to delete file %s: %v", clip.FilePath, err)
		}
	}

	return w.repo.UpdateStatus(ctx, clip.ID, entities.StatusExpired)
}
