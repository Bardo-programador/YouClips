package entities

import "time"

type ClipStatus string

const (
	StatusProcessing ClipStatus = "processing"
	StatusCompleted  ClipStatus = "completed"
	StatusFailed     ClipStatus = "failed"
)

type ClipFormat string

const (
	FormatVideo ClipFormat = "video"
	FormatAudio ClipFormat = "audio"
)

type Clip struct {
	ID              int        `json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	Title           string     `json:"title"`
	StartTime       int        `json:"start_time"`
	EndTime         int        `json:"end_time"`
	DurationSeconds int        `json:"duration_seconds"`
	Format          ClipFormat `json:"format"`
	Size            int64      `json:"size"`
	FilePath        string     `json:"file_path"`
	OriginalURL     string     `json:"original_url"`
	Status          ClipStatus `json:"status"`
}
