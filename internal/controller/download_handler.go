package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"videodownloader/internal/entities"
	"videodownloader/internal/service"
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

	if len(parts) == 2 && parts[1] == "download" {
		h.downloadClip(w, r, id)
		return
	}

	if r.Method == http.MethodGet {
		h.getClip(w, r, id)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ClipHandler) createClip(w http.ResponseWriter, r *http.Request) {
	var req CreateClipRequest
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

	clip, err := h.service.CreateClip(r.Context(), req.URL, req.StartTime, req.EndTime, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := CreateClipResponse{
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

	response := ClipResponse{
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

	clipResponses := make([]*ClipResponse, len(clips))
	for i, clip := range clips {
		clipResponses[i] = &ClipResponse{
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

	response := ListClipsResponse{
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

	if clip.Status != entities.StatusCompleted {
		http.Error(w, "Clip not ready for download", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", clip.FilePath))
	if clip.Format == entities.FormatVideo {
		w.Header().Set("Content-Type", "video/mp4")
	} else {
		w.Header().Set("Content-Type", "audio/mpeg")
	}

	http.ServeFile(w, r, clip.FilePath)
}
