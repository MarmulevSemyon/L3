package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"time"

	"Shortener/internal/dto"
	"Shortener/internal/model"
	"Shortener/internal/repository"

	"go.uber.org/zap"
)

var (
	// ErrInvalidURL возвращается при невалидном исходном URL.
	ErrInvalidURL = errors.New("invalid url")

	// ErrAliasTaken возвращается, если указанный alias уже занят.
	ErrAliasTaken = errors.New("alias already exists")

	// ErrLinkNotFound возвращается, если короткая ссылка не найдена.
	ErrLinkNotFound = errors.New("link not found")

	// ErrExpiredLink возвращается, если срок действия ссылки истёк.
	ErrExpiredLink = errors.New("link expired")
)

// Service реализует бизнес-логику сервиса сокращения ссылок.
type Service struct {
	repo repository.Repository
	log  *zap.Logger
}

// NewService создаёт сервис бизнес-логики.
func NewService(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// CreateShortLink создаёт короткую ссылку на основе входного запроса.
func (s *Service) CreateShortLink(ctx context.Context, req dto.ShortenRequest, baseURL string) (*dto.ShortenResponse, error) {
	if !isValidURL(req.URL) {
		return nil, ErrInvalidURL
	}

	alias := req.Alias
	if alias == "" {
		var err error
		alias, err = s.generateUniqueAlias(ctx, 6)
		if err != nil {
			return nil, err
		}
	} else {
		exists, err := s.repo.AliasExists(ctx, alias)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrAliasTaken
		}
	}

	link := &model.Link{
		Alias:       alias,
		OriginalURL: req.URL,
		ExpiresAt:   req.ExpiresAt,
	}
	if err := s.repo.CreateLink(ctx, link); err != nil {
		return nil, err
	}

	return &dto.ShortenResponse{
		Alias: alias,
		Short: baseURL + "/s/" + alias,
	}, nil
}

// ResolveAndTrack находит исходный URL по alias и сохраняет информацию о переходе.
func (s *Service) ResolveAndTrack(ctx context.Context, alias, userAgent, ip string) (string, error) {
	link, err := s.repo.GetLinkByAlias(ctx, alias)
	if err != nil {
		return "", ErrLinkNotFound
	}

	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return "", ErrExpiredLink
	}

	if err := s.repo.SaveClick(ctx, link.ID, userAgent, ip, time.Now()); err != nil {
		return "", err
	}

	return link.OriginalURL, nil
}

// GetAnalytics возвращает аналитику по короткой ссылке с учётом типа группировки.
func (s *Service) GetAnalytics(ctx context.Context, alias, groupBy string) (any, error) {
	switch groupBy {
	case "":
		return s.repo.GetRawAnalytics(ctx, alias)
	case "day":
		return s.repo.GetAnalyticsByDay(ctx, alias)
	case "month":
		return s.repo.GetAnalyticsByMonth(ctx, alias)
	case "user_agent":
		return s.repo.GetAnalyticsByUserAgent(ctx, alias)
	default:
		return nil, errors.New("invalid group_by")
	}
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func (s *Service) generateUniqueAlias(ctx context.Context, n int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for i := 0; i < 10; i++ {
		buf := make([]byte, n)
		for j := range buf {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
			if err != nil {
				return "", err
			}
			buf[j] = alphabet[num.Int64()]
		}

		alias := string(buf)
		exists, err := s.repo.AliasExists(ctx, alias)
		if err != nil {
			return "", err
		}
		if !exists {
			return alias, nil
		}
	}

	return "", errors.New("failed to generate unique alias")
}
