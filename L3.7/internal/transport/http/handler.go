package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strconv"

	"warehousecontrol/internal/domain"
	"warehousecontrol/internal/service"

	"github.com/wb-go/wbf/ginext"
)

// AuthService описывает методы сервиса авторизации.
type AuthService interface {
	Login(ctx context.Context, input service.LoginRequest) (service.LoginResponse, error)
}

// ItemService описывает методы сервиса товаров.
type ItemService interface {
	Create(ctx context.Context, actor domain.Actor, input domain.CreateItemInput) (domain.Item, error)
	List(ctx context.Context, actor domain.Actor) ([]domain.Item, error)
	GetByID(ctx context.Context, actor domain.Actor, id int64) (domain.Item, error)
	Update(ctx context.Context, actor domain.Actor, id int64, input domain.UpdateItemInput) (domain.Item, error)
	Delete(ctx context.Context, actor domain.Actor, id int64) error
}

// HistoryService описывает методы сервиса истории.
type HistoryService interface {
	ListByItemID(ctx context.Context, actor domain.Actor, itemID int64) ([]domain.ItemHistory, error)
}

// Handler хранит HTTP-обработчики сервиса.
type Handler struct {
	authService    AuthService
	itemService    ItemService
	historyService HistoryService
}

// NewHandler создаёт новый HTTP-handler.
func NewHandler(
	authService AuthService,
	itemService ItemService,
	historyService HistoryService,
) *Handler {
	return &Handler{
		authService:    authService,
		itemService:    itemService,
		historyService: historyService,
	}
}

// Index отдаёт простую HTML-страницу.
func (h *Handler) Index(c *ginext.Context) {
	c.File("./web/index.html")
}

// Health возвращает состояние сервиса.
func (h *Handler) Health(c *ginext.Context) {
	c.JSON(stdhttp.StatusOK, ginext.H{
		"status":  "ok",
		"service": "warehousecontrol",
	})
}

// Login выдаёт JWT для выбранного пользователя.
func (h *Handler) Login(c *ginext.Context) {
	var input service.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{
			"error": "invalid json body",
		})
		return
	}

	response, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			c.JSON(stdhttp.StatusBadRequest, ginext.H{
				"error": err.Error(),
			})
		default:
			c.JSON(stdhttp.StatusUnauthorized, ginext.H{
				"error": "invalid username",
			})
		}
		return
	}

	c.JSON(stdhttp.StatusOK, response)
}

// Me возвращает пользователя из JWT.
func (h *Handler) Me(c *ginext.Context) {
	actor, ok := ActorFromContext(c)
	if !ok {
		c.JSON(stdhttp.StatusUnauthorized, ginext.H{
			"error": "unauthorized",
		})
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{
		"user": actor,
	})
}

// CreateItem создаёт товар.
func (h *Handler) CreateItem(c *ginext.Context) {
	actor, ok := ActorFromContext(c)
	if !ok {
		c.JSON(stdhttp.StatusUnauthorized, ginext.H{
			"error": "unauthorized",
		})
		return
	}

	var input domain.CreateItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{
			"error": "invalid json body",
		})
		return
	}

	item, err := h.itemService.Create(c.Request.Context(), actor, input)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(stdhttp.StatusCreated, ginext.H{
		"result": item,
	})
}

// ListItems возвращает список товаров.
func (h *Handler) ListItems(c *ginext.Context) {
	actor, ok := ActorFromContext(c)
	if !ok {
		c.JSON(stdhttp.StatusUnauthorized, ginext.H{
			"error": "unauthorized",
		})
		return
	}

	items, err := h.itemService.List(c.Request.Context(), actor)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{
		"result": items,
	})
}

// GetItem возвращает товар по ID.
func (h *Handler) GetItem(c *ginext.Context) {
	actor, ok := ActorFromContext(c)
	if !ok {
		c.JSON(stdhttp.StatusUnauthorized, ginext.H{
			"error": "unauthorized",
		})
		return
	}

	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{
			"error": err.Error(),
		})
		return
	}

	item, err := h.itemService.GetByID(c.Request.Context(), actor, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{
		"result": item,
	})
}

// UpdateItem обновляет товар.
func (h *Handler) UpdateItem(c *ginext.Context) {
	actor, ok := ActorFromContext(c)
	if !ok {
		c.JSON(stdhttp.StatusUnauthorized, ginext.H{
			"error": "unauthorized",
		})
		return
	}

	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{
			"error": err.Error(),
		})
		return
	}

	var input domain.UpdateItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{
			"error": "invalid json body",
		})
		return
	}

	item, err := h.itemService.Update(c.Request.Context(), actor, id, input)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{
		"result": item,
	})
}

// DeleteItem удаляет товар.
func (h *Handler) DeleteItem(c *ginext.Context) {
	actor, ok := ActorFromContext(c)
	if !ok {
		c.JSON(stdhttp.StatusUnauthorized, ginext.H{
			"error": "unauthorized",
		})
		return
	}

	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.itemService.Delete(c.Request.Context(), actor, id); err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{
		"result": "deleted",
	})
}

func parseIDParam(c *ginext.Context) (int64, error) {
	idRaw := c.Param("id")

	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}

	return id, nil
}

func writeServiceError(c *ginext.Context, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		c.JSON(stdhttp.StatusForbidden, ginext.H{
			"error": "forbidden",
		})
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(stdhttp.StatusBadRequest, ginext.H{
			"error": err.Error(),
		})
	case errors.Is(err, service.ErrNotFound):
		c.JSON(stdhttp.StatusNotFound, ginext.H{
			"error": "not found",
		})
	default:
		c.JSON(stdhttp.StatusInternalServerError, ginext.H{
			"error": "internal server error",
		})
	}
}

// GetItemHistory возвращает историю изменений товара.
func (h *Handler) GetItemHistory(c *ginext.Context) {
	actor, ok := ActorFromContext(c)
	if !ok {
		c.JSON(stdhttp.StatusUnauthorized, ginext.H{
			"error": "unauthorized",
		})
		return
	}

	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ginext.H{
			"error": err.Error(),
		})
		return
	}

	history, err := h.historyService.ListByItemID(c.Request.Context(), actor, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(stdhttp.StatusOK, ginext.H{
		"result": history,
	})
}
