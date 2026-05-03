package entities

type CreateClipRequest struct {
	URL       string `json:"url"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	Format    string `json:"format"`
	Quality   string `json:"quality"` // For video: 360p, 480p, 720p, 1080p. For audio: ignored
}

type CreateClipResponse struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

type ClipResponse struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	Title       string `json:"title"`
	Duration    int    `json:"duration"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"download_url,omitempty"`
}

type ListClipsResponse struct {
	Clips []*ClipResponse `json:"clips"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
}

type VideoMetadataRequest struct {
	URL string `json:"url"`
}

type VideoMetadataResponse struct {
	Title            string            `json:"title"`
	Duration         int               `json:"duration"`
	DurationString   string            `json:"duration_string"`
	Channel          string            `json:"channel"`
	ChannelURL       string            `json:"channel_url"`
	Uploader         string            `json:"uploader"`
	UploaderID       string            `json:"uploader_id"`
	UploadDate       string            `json:"upload_date"`
	Thumbnail        string            `json:"thumbnail"`
	Categories       []string          `json:"categories"`
	Ext              string            `json:"ext"`
	FilesizeApprox   int64             `json:"filesize_approx"`
	WebpageURL       string            `json:"webpage_url"`
	MaxQuality       string            `json:"max_quality"`
	Formats          []FormatInfo      `json:"formats"`
}

type DeleteClipResponse struct {
	Deleted 	bool `json:"deleted"`
	ID				int  `json:"id"`
}

type ClipProgressResponse struct {
	ID       int    `json:"id"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

