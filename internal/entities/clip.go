package entities

import "time"

type ClipStatus string

const (
	StatusProcessing ClipStatus = "processing"
	StatusCompleted  ClipStatus = "completed"
	StatusFailed     ClipStatus = "failed"
	StatusExpired    ClipStatus = "expired"
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
	Quality         string     `json:"quality"`
	Size            int64      `json:"size"`
	ExpectedSize    int64      `json:"expected_size"`
	FilePath        string     `json:"file_path"`
	OriginalURL     string     `json:"original_url"`
	Status          ClipStatus `json:"status"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}

type VideoMetadata struct {
	ID                 int       `json:"id"`
	URL                string    `json:"url"`
	Title              string    `json:"title"`
	Duration           int       `json:"duration"`
	DurationString     string    `json:"duration_string"`
	Channel            string    `json:"channel"`
	ChannelURL         string    `json:"channel_url"`
	Uploader           string    `json:"uploader"`
	UploaderID         string    `json:"uploader_id"`
	UploadDate         string    `json:"upload_date"`
	Thumbnail          string    `json:"thumbnail"`
	Categories         string    `json:"categories"`       // JSON array stored as string
	Ext                string    `json:"ext"`
	FilesizeApprox     int64     `json:"filesize_approx"`
	Formats            string    `json:"formats"`          // JSON array stored as string
	WebpageURL         string    `json:"webpage_url"`
	MaxQuality         string    `json:"max_quality"`      // Maximum available quality (e.g., "1080p")
	CreatedAt          time.Time `json:"created_at"`
}

// YTDLPMetadata represents the raw JSON response from yt-dlp --dump-json
type YTDLPMetadata struct {
	Title              string        `json:"title"`
	Duration           int           `json:"duration"`
	DurationString     string        `json:"duration_string"`
	Channel            string        `json:"channel"`
	ChannelURL         string        `json:"channel_url"`
	Uploader           string        `json:"uploader"`
	UploaderID         string        `json:"uploader_id"`
	UploadDate         string        `json:"upload_date"`
	Thumbnail          string        `json:"thumbnail"`
	Categories         []string      `json:"categories"`
	Ext                string        `json:"ext"`
	FilesizeApprox     int64         `json:"filesize_approx"`
	WebpageURL         string        `json:"webpage_url"`
	Formats            []FormatInfo  `json:"formats"`
}

// FormatInfo represents format information from yt-dlp
type FormatInfo struct {
	FormatID       string  `json:"format_id"`
	FormatNote     string  `json:"format_note"`
	Ext            string  `json:"ext"`
	Filesize       int64   `json:"filesize"`
	FilesizeApprox int64   `json:"filesize_approx"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	FPS            float64 `json:"fps"`
	VideoCodec     string  `json:"vcodec"`
	AudioCodec     string  `json:"acodec"`
	Resolution     string  `json:"resolution"`
	TBR            float64 `json:"tbr"`
	VBR            float64 `json:"vbr"`
	ABR            float64 `json:"abr"`
}
