package repository

import (
	"context"
	"database/sql"
	"time"

	"Shortener/internal/dto"
	"Shortener/internal/model"

	// Регистрирует драйвер PostgreSQL для database/sql.
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type repo struct {
	db  *sql.DB
	log *zap.Logger
}

// NewRepository создаёт репозиторий для работы с PostgreSQL.
func NewRepository(masterDSN string, _ []string, log *zap.Logger) (Repository, error) {
	db, err := sql.Open("postgres", masterDSN)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &repo{db: db, log: log}, nil
}

func (r *repo) CreateLink(ctx context.Context, link *model.Link) error {
	query := `
		INSERT INTO links(alias, original_url, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query, link.Alias, link.OriginalURL, link.ExpiresAt).
		Scan(&link.ID, &link.CreatedAt)
}

func (r *repo) GetLinkByAlias(ctx context.Context, alias string) (*model.Link, error) {
	query := `
		SELECT id, alias, original_url, created_at, expires_at
		FROM links
		WHERE alias = $1
	`
	var l model.Link
	err := r.db.QueryRowContext(ctx, query, alias).
		Scan(&l.ID, &l.Alias, &l.OriginalURL, &l.CreatedAt, &l.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *repo) AliasExists(ctx context.Context, alias string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM links WHERE alias = $1)`, alias,
	).Scan(&exists)
	return exists, err
}

func (r *repo) SaveClick(ctx context.Context, linkID int64, userAgent, ip string, clickedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO link_clicks(link_id, clicked_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4)
	`, linkID, clickedAt, userAgent, ip)
	return err
}

func (r *repo) GetAnalyticsByDay(ctx context.Context, alias string) ([]dto.DayAnalyticsItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DATE(c.clicked_at)::text AS date, COUNT(*) AS clicks
		FROM link_clicks c
		JOIN links l ON l.id = c.link_id
		WHERE l.alias = $1
		GROUP BY DATE(c.clicked_at)
		ORDER BY DATE(c.clicked_at) DESC
	`, alias)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dto.DayAnalyticsItem
	for rows.Next() {
		var item dto.DayAnalyticsItem
		if err := rows.Scan(&item.Date, &item.Clicks); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *repo) GetAnalyticsByMonth(ctx context.Context, alias string) ([]dto.MonthAnalyticsItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT TO_CHAR(DATE_TRUNC('month', c.clicked_at), 'YYYY-MM') AS month,
		       COUNT(*) AS clicks
		FROM link_clicks c
		JOIN links l ON l.id = c.link_id
		WHERE l.alias = $1
		GROUP BY DATE_TRUNC('month', c.clicked_at)
		ORDER BY DATE_TRUNC('month', c.clicked_at) DESC
	`, alias)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dto.MonthAnalyticsItem
	for rows.Next() {
		var item dto.MonthAnalyticsItem
		if err := rows.Scan(&item.Month, &item.Clicks); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *repo) GetAnalyticsByUserAgent(ctx context.Context, alias string) ([]dto.UserAgentAnalyticsItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT COALESCE(NULLIF(c.user_agent, ''), 'unknown') AS user_agent,
		       COUNT(*) AS clicks
		FROM link_clicks c
		JOIN links l ON l.id = c.link_id
		WHERE l.alias = $1
		GROUP BY COALESCE(NULLIF(c.user_agent, ''), 'unknown')
		ORDER BY clicks DESC
	`, alias)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dto.UserAgentAnalyticsItem
	for rows.Next() {
		var item dto.UserAgentAnalyticsItem
		if err := rows.Scan(&item.UserAgent, &item.Clicks); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *repo) GetRawAnalytics(ctx context.Context, alias string) ([]dto.RawClickResponse, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.clicked_at, c.user_agent, c.ip_address
		FROM link_clicks c
		JOIN links l ON l.id = c.link_id
		WHERE l.alias = $1
		ORDER BY c.clicked_at DESC
	`, alias)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dto.RawClickResponse
	for rows.Next() {
		var item dto.RawClickResponse
		if err := rows.Scan(&item.ClickedAt, &item.UserAgent, &item.IPAddress); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
