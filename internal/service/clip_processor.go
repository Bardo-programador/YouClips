package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
	"youclips/internal/entities"
	"youclips/internal/repository"
)

type YTDLPProcessor struct {
	repo           repository.ClipRepository
	metadataRepo   repository.MetadataRepository
	storageDir     string
	clipTTL        time.Duration
	// Track active downloads for cancellation
	activeDownloads map[int]context.CancelFunc
	mu              sync.RWMutex
}

func NewYTDLPProcessor(repo repository.ClipRepository, metadataRepo repository.MetadataRepository, storageDir string, clipTTL time.Duration) *YTDLPProcessor {
	return &YTDLPProcessor{
		repo:            repo,
		storageDir:      storageDir,
		metadataRepo:    metadataRepo,
		clipTTL:         clipTTL,
		activeDownloads: make(map[int]context.CancelFunc),
	}
}

// CancelDownload cancels an active download by clip ID
func (p *YTDLPProcessor) CancelDownload(clipID int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if cancel, exists := p.activeDownloads[clipID]; exists {
		cancel() // Cancel the context
		delete(p.activeDownloads, clipID)
	}
}

func (p *YTDLPProcessor) ProcessClip(ctx context.Context, clip *entities.Clip) error {
	// Create a cancellable context for this download
	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	// Register the cancel function
	p.mu.Lock()
	p.activeDownloads[clip.ID] = cancel
	p.mu.Unlock()
	
	// Ensure cleanup
	defer func() {
		p.mu.Lock()
		delete(p.activeDownloads, clip.ID)
		p.mu.Unlock()
	}()
	
	if err := os.MkdirAll(p.storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	ytdlpMeta, err := p.getVideoMetadataJSON(downloadCtx, clip.OriginalURL)
	if err != nil {
		return fmt.Errorf("failed to get video metadata: %w", err)
	}
	
	clip.Title = ytdlpMeta.Title
	
	// Calculate expected clip size based on proportion of video duration
	// expectedClipSize = (videoSize / videoDuration) * clipDuration
	if ytdlpMeta.FilesizeApprox > 0 && ytdlpMeta.Duration > 0 {
		clip.ExpectedSize = (ytdlpMeta.FilesizeApprox / int64(ytdlpMeta.Duration)) * int64(clip.DurationSeconds)
	}
	
	// Set FilePath early so progress tracking can find the .part files
	var filename string
	if clip.Format == entities.FormatVideo {
		filename = fmt.Sprintf("clip_%d.mp4", clip.ID)
	} else {
		filename = fmt.Sprintf("clip_%d.mp3", clip.ID)
	}
	clip.FilePath = filepath.Join(p.storageDir, filename)
	
	p.repo.Update(downloadCtx, clip) // Update clip with title, expected size, and filepath before downloading
	if clip.EndTime > ytdlpMeta.Duration {
		clip.Status = entities.StatusFailed
		_ = p.repo.Update(downloadCtx, clip)
		return fmt.Errorf("requested end_time %d exceeds video duration %d", clip.EndTime, ytdlpMeta.Duration)
	}
	
	
	var outputPath string
	var err2 error

	if clip.Format == entities.FormatVideo {
		outputPath, err2 = p.downloadVideo(downloadCtx, clip)
	} else {
		outputPath, err2 = p.downloadAudio(downloadCtx, clip)
	}
	if err2 != nil {
		// Check if it was cancelled
		if downloadCtx.Err() == context.Canceled {
			return fmt.Errorf("download cancelled")
		}
		return fmt.Errorf("failed to download clip: %w", err2)
	}
	
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	
	expiresAt := time.Now().Add(p.clipTTL)
	
	clip.FilePath = outputPath
	clip.Size = fileInfo.Size()
	clip.Status = entities.StatusCompleted
	clip.ExpiresAt = &expiresAt

	if err := p.repo.Update(downloadCtx, clip); err != nil {
		return fmt.Errorf("failed to update clip: %w", err)
	}
	return nil
}
// getVideoMetadataJSON fetches video metadata using yt-dlp --dump-json
func (p *YTDLPProcessor) getVideoMetadataJSON(ctx context.Context, url string) (*entities.YTDLPMetadata, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "--dump-json", url)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute yt-dlp: %w", err)
	}

	var metadata entities.YTDLPMetadata
	if err := json.Unmarshal(output, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse JSON metadata: %w", err)
	}

	return &metadata, nil
}

func (p *YTDLPProcessor) downloadVideo(ctx context.Context, clip *entities.Clip) (string, error) {
	filename := fmt.Sprintf("clip_%d.mp4", clip.ID)
	outputPath := filepath.Join(p.storageDir, filename)
	
	// Quality-based format selection
	var formatSelector string
	switch clip.Quality {
	case "360p":
		formatSelector = "bestvideo[height<=360][ext=mp4]+bestaudio[ext=m4a]/best[height<=360]/best"
	case "480p":
		formatSelector = "bestvideo[height<=480][ext=mp4]+bestaudio[ext=m4a]/best[height<=480]/best"
	case "720p":
		formatSelector = "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720]/best"
	case "1080p":
		formatSelector = "bestvideo[height<=1080][ext=mp4]+bestaudio[ext=m4a]/best[height<=1080]/best"
	default:
		formatSelector = "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720]/best"
	}
	
	args := []string{
		"-f", formatSelector,
		"--merge-output-format", "mp4",
		"--download-sections", fmt.Sprintf("*%d-%d", clip.StartTime, clip.EndTime),
		"-o", outputPath,
		clip.OriginalURL,
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	
	if err := cmd.Run(); err != nil {
		return "", err
	}
	
	return outputPath, nil
}

func (p *YTDLPProcessor) downloadAudio(ctx context.Context, clip *entities.Clip) (string, error) {
	filename := fmt.Sprintf("clip_%d.mp3", clip.ID)
	outputPath := filepath.Join(p.storageDir, filename)
	
	// For audio, use 240p video source before extraction (smaller file, faster download)
	args := []string{
		"-f", "bestvideo[height<=240]+bestaudio/best[height<=240]/best",
		"-x", // Extract audio
		"--audio-format", "mp3",
		"--audio-quality", "0", // Best quality (VBR 220-260 kbps)
		"--download-sections", fmt.Sprintf("*%d-%d", clip.StartTime, clip.EndTime),
		"-o", outputPath,
		clip.OriginalURL,
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return outputPath, nil
}

func (p *YTDLPProcessor) GetVideoMetadata(ctx context.Context, url string) (*entities.VideoMetadataResponse, error) {
	// Check if metadata exists in database
	cached, err := p.metadataRepo.GetByURL(ctx, url)
	if err == nil && cached != nil {
		// Parse JSON strings back to arrays
		var categories []string
		var formats []entities.FormatInfo
		
		if cached.Categories != "" {
			json.Unmarshal([]byte(cached.Categories), &categories)
		}
		if cached.Formats != "" {
			json.Unmarshal([]byte(cached.Formats), &formats)
		}
		
		return &entities.VideoMetadataResponse{
			Title:          cached.Title,
			Duration:       cached.Duration,
			DurationString: cached.DurationString,
			Channel:        cached.Channel,
			ChannelURL:     cached.ChannelURL,
			Uploader:       cached.Uploader,
			UploaderID:     cached.UploaderID,
			UploadDate:     cached.UploadDate,
			Thumbnail:      cached.Thumbnail,
			Categories:     categories,
			Ext:            cached.Ext,
			FilesizeApprox: cached.FilesizeApprox,
			WebpageURL:     cached.WebpageURL,
			Formats:        formats,
		}, nil
	}

	// Metadata not found, fetch from YouTube using --dump-json
	ytdlpMeta, err := p.getVideoMetadataJSON(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get video metadata: %w", err)
	}

	// Convert arrays to JSON strings for database storage
	categoriesJSON, _ := json.Marshal(ytdlpMeta.Categories)
	formatsJSON, _ := json.Marshal(ytdlpMeta.Formats)

	// Save to database for future requests
	metadata := &entities.VideoMetadata{
		URL:            url,
		Title:          ytdlpMeta.Title,
		Duration:       ytdlpMeta.Duration,
		DurationString: ytdlpMeta.DurationString,
		Channel:        ytdlpMeta.Channel,
		ChannelURL:     ytdlpMeta.ChannelURL,
		Uploader:       ytdlpMeta.Uploader,
		UploaderID:     ytdlpMeta.UploaderID,
		UploadDate:     ytdlpMeta.UploadDate,
		Thumbnail:      ytdlpMeta.Thumbnail,
		Categories:     string(categoriesJSON),
		Ext:            ytdlpMeta.Ext,
		FilesizeApprox: ytdlpMeta.FilesizeApprox,
		Formats:        string(formatsJSON),
		WebpageURL:     ytdlpMeta.WebpageURL,
	}
	
	if err := p.metadataRepo.Create(ctx, metadata); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache metadata: %v\n", err)
	}

	return &entities.VideoMetadataResponse{
		Title:          ytdlpMeta.Title,
		Duration:       ytdlpMeta.Duration,
		DurationString: ytdlpMeta.DurationString,
		Channel:        ytdlpMeta.Channel,
		ChannelURL:     ytdlpMeta.ChannelURL,
		Uploader:       ytdlpMeta.Uploader,
		UploaderID:     ytdlpMeta.UploaderID,
		UploadDate:     ytdlpMeta.UploadDate,
		Thumbnail:      ytdlpMeta.Thumbnail,
		Categories:     ytdlpMeta.Categories,
		Ext:            ytdlpMeta.Ext,
		FilesizeApprox: ytdlpMeta.FilesizeApprox,
		WebpageURL:     ytdlpMeta.WebpageURL,
		Formats:        ytdlpMeta.Formats,
	}, nil
}

