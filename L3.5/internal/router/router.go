package router

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

// Handler описывает набор HTTP-обработчиков, необходимых для инициализации роутера.
type Handler interface {
	CreateEvent(c *ginext.Context)
	GetEvent(c *ginext.Context)
	ListEvents(c *ginext.Context)
	BookEvent(c *ginext.Context)
	ConfirmBooking(c *ginext.Context)
	CreateUser(c *ginext.Context)
	ListUsers(c *ginext.Context)
	GetUserBookings(c *ginext.Context)
}

// InitRouter настраивает HTTP-роутер приложения, API-маршруты и веб-страницы.
func InitRouter(mode string, h Handler, mw ...ginext.HandlerFunc) *ginext.Engine {
	router := ginext.New(mode)
	router.Use(mw...)

	api := router.Group("/api")
	{
		api.POST("/events", h.CreateEvent)
		api.GET("/events", h.ListEvents)
		api.GET("/events/:id", h.GetEvent)

		api.POST("/events/:id/book", h.BookEvent)
		api.POST("/events/:id/confirm", h.ConfirmBooking)

		api.POST("/users", h.CreateUser)
		api.GET("/users", h.ListUsers)
		api.GET("/users/:id/bookings", h.GetUserBookings)
	}

	router.GET("/health", func(c *ginext.Context) {
		c.JSON(http.StatusOK, ginext.H{"status": "ok"})
	})

	router.LoadHTMLGlob("web/templates/*")
	router.Static("/static", "web/static")

	router.GET("/", func(c *ginext.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/admin", func(c *ginext.Context) {
		c.HTML(http.StatusOK, "admin.html", nil)
	})

	return router
}
