package http

import (
	"database/sql"
	"encoding/csv"
	"errors"
	stdhttp "net/http"
	"strconv"
	"time"

	"salestracker/internal/domain"
	"salestracker/internal/dto"
	"salestracker/internal/service"

	"github.com/wb-go/wbf/ginext"
)

// Handler хранит HTTP-обработчики сервиса.
type Handler struct {
	service *service.ItemService
}

// NewHandler создаёт новый HTTP-handler.
func NewHandler(service *service.ItemService) *Handler {
	return &Handler{service: service}
}

// Index отдаёт простую HTML-страницу.
func (h *Handler) Index(c *ginext.Context) {
	c.File("./web/index.html")
}

// CreateItem создаёт запись.
func (h *Handler) CreateItem(c *ginext.Context) {
	var req dto.CreateItemRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{"error": "invalid json"})
		return
	}

	item, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(stdhttp.StatusCreated, ginext.H{"item": item})
}

// GetItem возвращает запись по ID.
func (h *Handler) GetItem(c *ginext.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{"error": "invalid id"})
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{"item": item})
}

// ListItems возвращает список записей.
func (h *Handler) ListItems(c *ginext.Context) {
	filter, err := parseListFilter(c)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	items, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{"items": items})
}

// UpdateItem обновляет запись.
func (h *Handler) UpdateItem(c *ginext.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{"error": "invalid id"})
		return
	}

	var req dto.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{"error": "invalid json"})
		return
	}

	item, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{"item": item})
}

// DeleteItem удаляет запись.
func (h *Handler) DeleteItem(c *ginext.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{"result": "deleted"})
}

// GetAnalytics возвращает агрегированную аналитику.
func (h *Handler) GetAnalytics(c *ginext.Context) {
	filter, err := parseAnalyticsFilter(c)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	analytics, err := h.service.Analytics(c.Request.Context(), filter)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{"analytics": analytics})
}

// ExportCSV выгружает список записей в CSV.
func (h *Handler) ExportCSV(c *ginext.Context) {
	filter, err := parseListFilter(c)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	filter.Limit = 100000
	filter.Offset = 0

	items, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="items.csv"`)

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{
		"id", "type", "amount", "category", "description", "occurred_at", "created_at", "updated_at",
	})

	for _, item := range items {
		_ = writer.Write([]string{
			strconv.FormatInt(item.ID, 10),
			item.Type,
			strconv.FormatFloat(item.Amount, 'f', 2, 64),
			item.Category,
			item.Description,
			item.OccurredAt.Format(time.RFC3339),
			item.CreatedAt.Format(time.RFC3339),
			item.UpdatedAt.Format(time.RFC3339),
		})
	}
}

func (h *Handler) writeError(c *ginext.Context, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(stdhttp.StatusNotFound, ginext.H{"error": "item not found"})
	case errors.Is(err, service.ErrInvalidType),
		errors.Is(err, service.ErrInvalidAmount),
		errors.Is(err, service.ErrInvalidCategory),
		errors.Is(err, service.ErrInvalidOccurredAt):
		c.JSON(stdhttp.StatusBadRequest, ginext.H{"error": err.Error()})
	default:
		c.JSON(stdhttp.StatusInternalServerError, ginext.H{"error": "internal server error"})
	}
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}

func parseListFilter(c *ginext.Context) (domain.ListFilter, error) {
	var filter domain.ListFilter
	var err error

	filter.From, err = parseTimePtr(c.Query("from"))
	if err != nil {
		return filter, err
	}

	filter.To, err = parseTimePtr(c.Query("to"))
	if err != nil {
		return filter, err
	}

	filter.Type = c.Query("type")
	filter.Category = c.Query("category")
	filter.SortBy = c.DefaultQuery("sort_by", "occurred_at")
	filter.Order = c.DefaultQuery("order", "desc")

	filter.Limit = 50
	if rawLimit := c.Query("limit"); rawLimit != "" {
		filter.Limit, err = strconv.Atoi(rawLimit)
		if err != nil {
			return filter, errors.New("invalid limit")
		}
	}

	filter.Offset = 0
	if rawOffset := c.Query("offset"); rawOffset != "" {
		filter.Offset, err = strconv.Atoi(rawOffset)
		if err != nil {
			return filter, errors.New("invalid offset")
		}
	}

	return filter, nil
}

func parseAnalyticsFilter(c *ginext.Context) (domain.AnalyticsFilter, error) {
	var filter domain.AnalyticsFilter
	var err error

	filter.From, err = parseTimePtr(c.Query("from"))
	if err != nil {
		return filter, err
	}

	filter.To, err = parseTimePtr(c.Query("to"))
	if err != nil {
		return filter, err
	}

	filter.Type = c.Query("type")
	filter.Category = c.Query("category")

	return filter, nil
}

func parseTimePtr(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err == nil {
		return &parsed, nil
	}

	parsed, err = time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, errors.New("invalid time, use RFC3339 or YYYY-MM-DD")
	}

	return &parsed, nil
}
