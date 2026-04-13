package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"commentTree/internal/dto"
	"commentTree/internal/model"
	"commentTree/internal/service"
)

// Handler обрабатывает HTTP-запросы комментариев.
type Handler struct {
	service service.CommentService
}

// New создает новый HTTP-handler комментариев.
func New(service service.CommentService) *Handler {
	return &Handler{service: service}
}

// CreateComment обрабатывает создание комментария.
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req dto.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	comment, err := h.service.CreateComment(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAuthor),
			errors.Is(err, service.ErrInvalidBody),
			errors.Is(err, service.ErrParentNotFound):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto.CommentResponse{Comment: *comment})
}

// GetComments обрабатывает получение дерева комментариев по parent.
func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query()

	parentID, err := parseOptionalInt64(query.Get("parent"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid parent")
		return
	}

	page, err := parseIntWithDefault(query.Get("page"), 1)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid page")
		return
	}

	limit, err := parseIntWithDefault(query.Get("limit"), 10)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid limit")
		return
	}

	sortBy := query.Get("sort_by")
	order := query.Get("order")

	comments, total, err := h.service.GetCommentTree(r.Context(), parentID, page, limit, sortBy, order)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if comments == nil {
		comments = []model.Comment{}
	}
	writeJSON(w, http.StatusOK, dto.CommentsResponse{
		Comments: comments,
		Page:     page,
		Limit:    limit,
		Total:    total,
	})
}

// GetAllComments обрабатывает получение всех корневых комментариев с деревом.
func (h *Handler) GetAllComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query()

	page, err := parseIntWithDefault(query.Get("page"), 1)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid page")
		return
	}

	limit, err := parseIntWithDefault(query.Get("limit"), 10)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid limit")
		return
	}

	sortBy := query.Get("sort_by")
	order := query.Get("order")

	comments, total, err := h.service.GetAllCommentTrees(r.Context(), page, limit, sortBy, order)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if comments == nil {
		comments = []model.Comment{}
	}
	writeJSON(w, http.StatusOK, dto.CommentsResponse{
		Comments: comments,
		Page:     page,
		Limit:    limit,
		Total:    total,
	})
}

// DeleteComment обрабатывает удаление комментария и всех его потомков.
func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := extractIDFromPath(r.URL.Path, "/comments/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	err = h.service.DeleteComment(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SearchComments обрабатывает полнотекстовый поиск комментариев.
func (h *Handler) SearchComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req dto.SearchCommentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	queryParams := r.URL.Query()

	page, err := parseIntWithDefault(queryParams.Get("page"), 1)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid page")
		return
	}

	limit, err := parseIntWithDefault(queryParams.Get("limit"), 10)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid limit")
		return
	}

	sortBy := queryParams.Get("sort_by")
	order := queryParams.Get("order")

	comments, total, err := h.service.SearchComments(r.Context(), req.Query, page, limit, sortBy, order)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidSearchQuery):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if comments == nil {
		comments = []model.Comment{}
	}
	writeJSON(w, http.StatusOK, dto.CommentsResponse{
		Comments: comments,
		Page:     page,
		Limit:    limit,
		Total:    total,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, dto.ErrorResponse{Error: message})
}

func parseIntWithDefault(value string, defaultValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}

func parseOptionalInt64(value string) (*int64, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func extractIDFromPath(path string, prefix string) (int64, error) {
	idPart := strings.TrimPrefix(path, prefix)
	idPart = strings.Trim(idPart, "/")

	return strconv.ParseInt(idPart, 10, 64)
}
