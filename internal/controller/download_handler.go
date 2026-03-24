package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"log"
	"strconv"
	"strings"
	"youclips/internal/entities"
	"youclips/internal/service"
	)

	type ClipHandler struct {
		service *service.ClipService
	}

	func NewClipHandler(service *service.ClipService) *ClipHandler {
		return &ClipHandler{service: service}
	}

	func (h *ClipHandler) Clips(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.createClip(w, r)
		case http.MethodGet:
			h.listClips(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}

	func (h *ClipHandler) ClipByID(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/clips/")
		parts := strings.Split(path, "/")
		
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "Invalid clip ID", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			http.Error(w, "Invalid clip ID", http.StatusBadRequest)
			return
		}

		if len(parts) == 2 && parts[1] == "download" && r.Method == http.MethodGet {
			h.downloadClip(w, r, id)
			return
		}

		if len(parts) == 2 && parts[1] == "progress" && r.Method == http.MethodGet {
			h.getProgress(w, r, id)
			return
		}
		
		switch r.Method {
		case http.MethodGet:
			h.getClip(w, r, id)
		case http.MethodDelete:
			h.deleteClip(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}

	func (h *ClipHandler) deleteClip(w http.ResponseWriter, r *http.Request, id int) {

		deleted, err := h.service.DeleteClip(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		response := entities.DeleteClipResponse{
			Deleted: deleted,
			ID:      id,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
	func (h *ClipHandler) createClip(w http.ResponseWriter, r *http.Request) {
		var req entities.CreateClipRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		format := entities.FormatVideo
		if req.Format == "audio" {
			format = entities.FormatAudio
		}

		clip, err := h.service.CreateClip(r.Context(), req.URL, req.StartTime, req.EndTime, format, req.Quality)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := entities.CreateClipResponse{
			ID:     clip.ID,
			Status: string(clip.Status),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}

	func (h *ClipHandler) getClip(w http.ResponseWriter, r *http.Request, id int) {
		clip, err := h.service.GetClip(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		response := entities.ClipResponse{
			ID:       clip.ID,
			Status:   string(clip.Status),
			Title:    clip.Title,
			Duration: clip.DurationSeconds,
			Size:     clip.Size,
		}

		if clip.Status == entities.StatusCompleted {
			response.DownloadURL = fmt.Sprintf("/clips/%d/download", clip.ID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	func (h *ClipHandler) listClips(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

		clips, total, err := h.service.ListClips(r.Context(), page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		clipResponses := make([]*entities.ClipResponse, len(clips))
		for i, clip := range clips {
			clipResponses[i] = &entities.ClipResponse{
				ID:       clip.ID,
				Status:   string(clip.Status),
				Title:    clip.Title,
				Duration: clip.DurationSeconds,
				Size:     clip.Size,
			}
			if clip.Status == entities.StatusCompleted {
				clipResponses[i].DownloadURL = fmt.Sprintf("/clips/%d/download", clip.ID)
			}
		}

		if page < 1 {
			page = 1
		}

	response := entities.ListClipsResponse{
		Clips: clipResponses,
		Total: total,
		Page:  page,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ClipHandler) downloadClip(w http.ResponseWriter, r *http.Request, id int) {
	clip, err := h.service.GetClip(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if clip.Status == entities.StatusExpired {
		http.Error(w, "Clip has expired. Please create a new clip.", http.StatusGone)
		return
	}

	if clip.Status != entities.StatusCompleted {
		http.Error(w, "Clip not ready for download", http.StatusBadRequest)
		return
	}

	// Check if file still exists
	if _, err := os.Stat(clip.FilePath); os.IsNotExist(err) {
		http.Error(w, "Clip file not found. It may have expired.", http.StatusGone)
		return
	}

	filename := "untitled"
	if clip.Title != "" {
		filename = clip.Title
	}

	var extension string
	switch clip.Format{
	case entities.FormatVideo:
		extension = ".mp4"
		w.Header().Set("Content-Type", "video/mp4")
	case entities.FormatAudio:
		extension = ".mp3"
		w.Header().Set("Content-Type", "audio/mpeg")
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s%s\"", filename, extension))
	http.ServeFile(w, r, clip.FilePath)
}


func (h *ClipHandler) ClipMetaData(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req entities.VideoMetadataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		url := req.URL

		metadata, err := h.service.GetVideoMetadata(r.Context(), url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ClipHandler) getProgress(w http.ResponseWriter, r *http.Request, id int) {
	clip, err := h.service.GetClip(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var progress int
	
	if clip.Status == entities.StatusCompleted {
		progress = 100
	} else if clip.Status == entities.StatusProcessing {
		// Calculate progress by comparing current file size with expected size
		// yt-dlp downloads to .part files, so we need to check those during download
		if clip.ExpectedSize > 0 {
			var partFilePath string
			
			// Determine the .part file path based on format
			if clip.Format == entities.FormatVideo {
				// Video files: clip_1.mp4 -> clip_1.mp4.part
				partFilePath = clip.FilePath + ".part"
			} else {
				// Audio files: clip_1.mp3 -> clip_1.webm.part (yt-dlp downloads webm first, then converts to mp3)
				partFilePath = strings.TrimSuffix(clip.FilePath, ".mp3") + ".webm.part"
			}
			
			log.Printf("Checking progress for clip %d - partFilePath: %s, expectedSize: %d", 
				clip.ID, partFilePath, clip.ExpectedSize)
			
			// Try to stat the .part file
			fileInfo, err := os.Stat(partFilePath)
			if err == nil {
				currentSize := fileInfo.Size()
				log.Printf("Clip %d (%s): current size %d bytes, expected size %d bytes", 
					clip.ID, clip.Format, currentSize, clip.ExpectedSize)
				
				progress = int((float64(currentSize) / float64(clip.ExpectedSize)) * 100)
				if progress > 100 {
					progress = 100
				}
			} else {
				log.Printf("Part file not found (%s), checking final file: %s", err.Error(), clip.FilePath)
				// If .part file doesn't exist, check if final file exists (might have just completed)
				fileInfo, err := os.Stat(clip.FilePath)
				if err == nil {
					currentSize := fileInfo.Size()
					log.Printf("Final file exists: %d bytes", currentSize)
					progress = int((float64(currentSize) / float64(clip.ExpectedSize)) * 100)
					if progress > 100 {
						progress = 100
					}
				} else {
					log.Printf("Final file also not found: %s - progress stays at 0", err.Error())
				}
				// If neither exists, progress stays at 0 (download hasn't started writing yet)
			}
		} else {
			log.Printf("Clip %d has no expected size set, cannot calculate progress", clip.ID)
		}
	}

	response := entities.ClipProgressResponse{
		ID:       clip.ID,
		Status:   string(clip.Status),
		Progress: progress,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
