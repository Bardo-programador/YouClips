package controller

type CreateClipRequest struct {
	URL       string `json:"url"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	Format    string `json:"format"`
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
