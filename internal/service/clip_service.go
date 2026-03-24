package service

import (
	"context"
	"fmt"
	"youclips/internal/entities"
	"youclips/internal/repository"
	"os"
)

type ClipService struct {
	repo      repository.ClipRepository
	processor *YTDLPProcessor // Changed from interface to concrete type
}

func NewClipService(repo repository.ClipRepository , processor *YTDLPProcessor) *ClipService {
	return &ClipService{
		repo:      repo,
		processor: processor,
	}
}

func (s *ClipService) CreateClip(ctx context.Context, url string, startTime, endTime int, format entities.ClipFormat, quality string) (*entities.Clip, error) {
	if startTime < 0 || endTime < 0 || endTime <= startTime {
		return nil, fmt.Errorf("invalid time range: start=%d, end=%d", startTime, endTime)
	}

	// Set default quality if not provided
	if quality == "" {
		if format == entities.FormatVideo {
			quality = "720p" // Default video quality
		} else {
			quality = "240p" // For audio, use 240p video source
		}
	}

	clip := &entities.Clip{
		OriginalURL:     url,
		StartTime:       startTime,
		EndTime:         endTime,
		DurationSeconds: endTime - startTime,
		Format:          format,
		Quality:         quality,
		Status:          entities.StatusProcessing,
	}

	if err := s.repo.Create(ctx, clip); err != nil {
		return nil, fmt.Errorf("failed to create clip: %w", err)
	}

	go func() {
		bgCtx := context.Background()
		if err := s.processor.ProcessClip(bgCtx, clip); err != nil {
			s.repo.UpdateStatus(bgCtx, clip.ID, entities.StatusFailed)
		}
	}()

	return clip, nil
}

func (s *ClipService) GetClip(ctx context.Context, id int) (*entities.Clip, error) {
	clip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get clip: %w", err)
	}
	if clip == nil {
		return nil, fmt.Errorf("clip not found")
	}
	return clip, nil
}

func (s *ClipService) ListClips(ctx context.Context, page, limit int) ([]*entities.Clip, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	clips, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list clips: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count clips: %w", err)
	}

	return clips, total, nil
}

func (s *ClipService) GetVideoMetadata(ctx context.Context, url string) (*entities.VideoMetadataResponse, error) {
	return s.processor.GetVideoMetadata(ctx, url)
}

func (s *ClipService) DeleteClip(ctx context.Context, id int) (bool, error) {

	clip, err := s.repo.GetByID(ctx, id)
	if err != nil || clip == nil {
		return false, fmt.Errorf("failed to get clip: %w", err)
	}

	// If clip is still processing, cancel the download and delete .part files immediately
	if clip.Status == entities.StatusProcessing {
		s.processor.CancelDownload(id)
		
		// Delete .part files immediately
		if clip.FilePath != "" {
			os.Remove(clip.FilePath + ".part")
			if clip.Format == entities.FormatAudio {
				// For audio, also remove .webm.part
				basePath := clip.FilePath[:len(clip.FilePath)-4]
				os.Remove(basePath + ".webm.part")
			}
		}
	}

	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to delete clip: %w", err)
	}
	 
	if deleted && clip.FilePath != "" {
		// Delete the final file if it exists
		os.Remove(clip.FilePath)
	}
	return deleted, nil
}

func (s *ClipService) GetClipSize(ctx context.Context, id int) (int64, error) {
	clip, err := s.repo.GetByID(ctx, id)
	if err != nil || clip == nil {
		return 0, fmt.Errorf("failed to get clip: %w", err)
	}
	fileInfo, err := os.Stat(clip.FilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get clip file size: %w", err)
	}
	return fileInfo.Size(), nil
}

