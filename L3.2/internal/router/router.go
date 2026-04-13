package router

import (
	"Shortener/internal/router/handlers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Router оборачивает HTTP-роутер приложения.
type Router struct {
	engine *gin.Engine
}

// NewRouter создаёт и настраивает маршруты HTTP-сервера.
func NewRouter(_ string, h *handlers.HandlerShortener, _ *zap.Logger) *Router {
	r := gin.Default()

	r.POST("/shorten", h.Shorten)
	r.GET("/s/:alias", h.Redirect)
	r.GET("/analytics/:alias", h.Analytics)

	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	return &Router{engine: r}
}

// GetEngine возвращает настроенный HTTP-движок.
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
