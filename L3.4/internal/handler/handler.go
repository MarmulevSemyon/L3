package handler

import (
	"encoding/json"
	"errors"
	"imageProcessor/internal/repository"
	"imageProcessor/internal/service"
	"net/http"
)

// Handler обрабатывает HTTP-запросы.
type Handler struct {
	svc *service.Service
}

// New создает handler.
func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// UploadImage обрабатывает POST /upload.
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	resize := r.FormValue("resize") == "true"
	thumb := r.FormValue("thumb") == "true"
	watermark := r.FormValue("watermark") == "true"

	resp, err := h.svc.UploadImage(r.Context(), file, header, resize, thumb, watermark)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, resp)
}

// GetImage обрабатывает GET /image/{id}.
func (h *Handler) GetImage(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	resp, err := h.svc.GetImageStatus(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrImageNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "image not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeleteImage обрабатывает DELETE /image/{id}.
func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	err := h.svc.DeleteImage(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrImageNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "image not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": "deleted"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
