package repository

import (
	"context"
	"time"

	"Shortener/internal/dto"
	"Shortener/internal/model"
)

// Repository описывает набор методов для работы с ссылками и аналитикой в хранилище.
type Repository interface {
	CreateLink(ctx context.Context, link *model.Link) error
	GetLinkByAlias(ctx context.Context, alias string) (*model.Link, error)
	AliasExists(ctx context.Context, alias string) (bool, error)

	SaveClick(ctx context.Context, linkID int64, userAgent, ip string, clickedAt time.Time) error

	GetRawAnalytics(ctx context.Context, alias string) ([]dto.RawClickResponse, error)
	GetAnalyticsByDay(ctx context.Context, alias string) ([]dto.DayAnalyticsItem, error)
	GetAnalyticsByMonth(ctx context.Context, alias string) ([]dto.MonthAnalyticsItem, error)
	GetAnalyticsByUserAgent(ctx context.Context, alias string) ([]dto.UserAgentAnalyticsItem, error)
}
