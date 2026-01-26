package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"strconv"
	"youclips/internal/entities"
	"youclips/internal/repository"
)

type YTDLPProcessor struct {
	repo       repository.ClipRepository
	storageDir string
}

func NewYTDLPProcessor(repo repository.ClipRepository, storageDir string) *YTDLPProcessor {
	return &YTDLPProcessor{
		repo:       repo,
		storageDir: storageDir,
	}
}

func (p *YTDLPProcessor) ProcessClip(ctx context.Context, clip *entities.Clip) error {
	if err := os.MkdirAll(p.storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	title, err := p.getVideoTitle(ctx, clip.OriginalURL)
	if err != nil {
		return fmt.Errorf("failed to get video title: %w", err)
	}
	clip.Title = title

	// Validate requested end time against actual video duration
	duration, err := p.getVideoDuration(ctx, clip.OriginalURL)
	if err != nil {
		return fmt.Errorf("failed to get video duration: %w", err)
	}
	if clip.EndTime > duration {
		clip.Status = entities.StatusFailed
		_ = p.repo.Update(ctx, clip) // best effort update
		return fmt.Errorf("requested end_time %d exceeds video duration %d", clip.EndTime, duration)
	}
	var outputPath string
	var err2 error

	if clip.Format == entities.FormatVideo {
		outputPath, err2 = p.downloadVideo(ctx, clip)
	} else {
		outputPath, err2 = p.downloadAudio(ctx, clip)
	}
	if err2 != nil {
		return fmt.Errorf("failed to download clip: %w", err2)
	}
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	clip.FilePath = outputPath
	clip.Size = fileInfo.Size()
	clip.Status = entities.StatusCompleted

	if err := p.repo.Update(ctx, clip); err != nil {
		return fmt.Errorf("failed to update clip: %w", err)
	}
	return nil
}
func (*YTDLPProcessor) getVideoDuration(ctx context.Context, url string) (int, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "--get-duration", url)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	time := strings.Split(strings.TrimSpace(string(output)), ":")
	
	switch len(time) {
	case 3:
		hours, _ := strconv.Atoi(time[0])
		minutes, _ := strconv.Atoi(time[1])
		seconds, _ := strconv.Atoi(time[2])
		return hours*3600 + minutes*60 + seconds, nil
	case 2:
		minutes, _ := strconv.Atoi(time[0])
		seconds, _ := strconv.Atoi(time[1])
		return minutes*60 + seconds, nil
	case 1:
		seconds, _ := strconv.Atoi(time[0])
		return seconds, nil
	default:
		return 0, fmt.Errorf("invalid duration format")
	}

	
}
func (p *YTDLPProcessor) getVideoTitle(ctx context.Context, url string) (string, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "--get-title", url)
	output, err := cmd.Output()
	if err != nil {
		return "", err 
	}
	return strings.TrimSpace(string(output)), nil
	}

func (p *YTDLPProcessor) downloadVideo(ctx context.Context, clip *entities.Clip) (string, error) {
	filename := fmt.Sprintf("clip_%d.mp4", clip.ID)
	outputPath := filepath.Join(p.storageDir, filename)
	args := []string{
		"-f", "bv*+ba*[ext=m4a]/b[ext=mp4]",
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
	tempVideoFile := fmt.Sprintf("clip_%d_temp.mp4", clip.ID)
	tempVideoPath := filepath.Join(p.storageDir, tempVideoFile)
	
	filename := fmt.Sprintf("clip_%d.mp3", clip.ID)
	outputPath := filepath.Join(p.storageDir, filename)

	// Step 1: Download video with precise cut
	args := []string{
		"-f", "bv*+ba*[ext=m4a]/b[ext=mp4]",
		"--merge-output-format", "mp4",
		"--download-sections", fmt.Sprintf("*%d-%d", clip.StartTime, clip.EndTime),
		"-o", tempVideoPath,
		clip.OriginalURL,
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Step 2: Extract audio from the cut video using ffmpeg
	ffmpegArgs := []string{
		"-i", tempVideoPath,
		"-vn",
		"-acodec", "libmp3lame",
		"-q:a", "2",
		outputPath,
	}

	ffmpegCmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...)
	if err := ffmpegCmd.Run(); err != nil {
		os.Remove(tempVideoPath)
		return "", err
	}

	// Step 3: Clean up temporary video file
	os.Remove(tempVideoPath)

	return outputPath, nil
}

func (p *YTDLPProcessor) GetVideoMetadata(ctx context.Context, url string) (*entities.VideoMetadataResponse, error) {
	title, err := p.getVideoTitle(ctx, url)
	if err != nil {
	return nil, fmt.Errorf("failed to get video title: %w", err)
	}

	duration, err := p.getVideoDuration(ctx, url)
	if err != nil {
	return nil, fmt.Errorf("failed to get video duration: %w", err)
	}

	return &entities.VideoMetadataResponse{
	Title:    title,
	Duration: duration,
	}, nil
}
