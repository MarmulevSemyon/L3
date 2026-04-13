package handlers

import (
	"errors"
	"net/http"

	"Shortener/internal/dto"
	"Shortener/internal/service"

	"github.com/gin-gonic/gin"
)

// HandlerShortener обрабатывает HTTP-запросы сервиса сокращения ссылок.
type HandlerShortener struct {
	service *service.Service
}

// NewHandlerShortener создаёт обработчик запросов сервиса сокращения ссылок.
func NewHandlerShortener(s *service.Service) *HandlerShortener {
	return &HandlerShortener{service: s}
}

// Shorten создаёт новую короткую ссылку.
func (h *HandlerShortener) Shorten(c *gin.Context) {
	var req dto.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	baseURL := "http://" + c.Request.Host
	resp, err := h.service.CreateShortLink(c.Request.Context(), req, baseURL)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrAliasTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Redirect выполняет переход по короткой ссылке и сохраняет информацию о клике.
func (h *HandlerShortener) Redirect(c *gin.Context) {
	alias := c.Param("alias")
	target, err := h.service.ResolveAndTrack(
		c.Request.Context(),
		alias,
		c.GetHeader("User-Agent"),
		c.ClientIP(),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrExpiredLink):
			c.JSON(http.StatusGone, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.Redirect(http.StatusMovedPermanently, target)
}

// Analytics возвращает статистику переходов по короткой ссылке.
func (h *HandlerShortener) Analytics(c *gin.Context) {
	alias := c.Param("alias")
	groupBy := c.Query("group_by")

	data, err := h.service.GetAnalytics(c.Request.Context(), alias, groupBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analytics": data})
}
