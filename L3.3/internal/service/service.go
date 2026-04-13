package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"commentTree/internal/dto"
	"commentTree/internal/model"
	"commentTree/internal/repository"
)

const maxChildrenLimit = 1000

var (
	ErrInvalidAuthor      = errors.New("author is required")
	ErrInvalidBody        = errors.New("body is required")
	ErrParentNotFound     = errors.New("parent comment not found")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrInvalidSearchQuery = errors.New("search query is required")
)

// CommentService описывает контракт бизнес-логики для комментариев.
type CommentService interface {
	CreateComment(ctx context.Context, req dto.CreateCommentRequest) (*model.Comment, error)
	GetCommentTree(ctx context.Context, parentID *int64, page, limit int, sortBy, order string) ([]model.Comment, int, error)
	GetAllCommentTrees(ctx context.Context, page, limit int, sortBy, order string) ([]model.Comment, int, error)
	DeleteComment(ctx context.Context, id int64) error
	SearchComments(ctx context.Context, query string, page, limit int, sortBy, order string) ([]model.Comment, int, error)
}

// Service реализует CommentService.
type Service struct {
	repo repository.CommentRepository
}

// New создает новый сервис комментариев.
func New(repo repository.CommentRepository) *Service {
	return &Service{repo: repo}
}

// Проверка, что Service реализует интерфейс CommentService.
var _ CommentService = (*Service)(nil)

// CreateComment создает новый комментарий.
func (s *Service) CreateComment(ctx context.Context, req dto.CreateCommentRequest) (*model.Comment, error) {
	req.Author = strings.TrimSpace(req.Author)
	req.Body = strings.TrimSpace(req.Body)

	if req.Author == "" {
		return nil, ErrInvalidAuthor
	}

	if req.Body == "" {
		return nil, ErrInvalidBody
	}

	if req.ParentID != nil {
		_, err := s.repo.GetByID(ctx, *req.ParentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrParentNotFound
			}
			return nil, err
		}
	}

	comment := &model.Comment{
		ParentID: req.ParentID,
		Author:   req.Author,
		Body:     req.Body,
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

// GetCommentTree возвращает дерево комментариев от parentID.
func (s *Service) GetCommentTree(ctx context.Context, parentID *int64, page, limit int, sortBy, order string) ([]model.Comment, int, error) {
	page, limit = normalizePagination(page, limit)
	offset := (page - 1) * limit

	comments, total, err := s.repo.GetChildren(ctx, parentID, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}

	for i := range comments {
		if err := s.fillChildrenRecursive(ctx, &comments[i], sortBy, order); err != nil {
			return nil, 0, err
		}
	}

	return comments, total, nil
}

// GetAllCommentTrees возвращает все корневые комментарии с деревом.
func (s *Service) GetAllCommentTrees(ctx context.Context, page, limit int, sortBy, order string) ([]model.Comment, int, error) {
	page, limit = normalizePagination(page, limit)
	offset := (page - 1) * limit

	comments, total, err := s.repo.GetAllRoots(ctx, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}

	for i := range comments {
		if err := s.fillChildrenRecursive(ctx, &comments[i], sortBy, order); err != nil {
			return nil, 0, err
		}
	}

	return comments, total, nil
}

// DeleteComment удаляет комментарий вместе с дочерними.
func (s *Service) DeleteComment(ctx context.Context, id int64) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCommentNotFound
		}
		return err
	}

	if err := s.repo.DeleteSubtree(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCommentNotFound
		}
		return err
	}

	return nil
}

// SearchComments выполняет полнотекстовый поиск комментариев.
func (s *Service) SearchComments(ctx context.Context, query string, page, limit int, sortBy, order string) ([]model.Comment, int, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, 0, ErrInvalidSearchQuery
	}

	page, limit = normalizePagination(page, limit)
	offset := (page - 1) * limit

	comments, total, err := s.repo.Search(ctx, query, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (s *Service) fillChildrenRecursive(ctx context.Context, comment *model.Comment, sortBy, order string) error {
	children, _, err := s.repo.GetChildren(ctx, &comment.ID, maxChildrenLimit, 0, sortBy, order)
	if err != nil {
		return err
	}

	for i := range children {
		if err := s.fillChildrenRecursive(ctx, &children[i], sortBy, order); err != nil {
			return err
		}
	}

	comment.Children = children
	return nil
}

func normalizePagination(page, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	return page, limit
}
