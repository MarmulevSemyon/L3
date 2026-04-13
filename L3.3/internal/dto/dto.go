package dto

import "commentTree/internal/model"

// CreateCommentRequest запрос на создание комментария.
type CreateCommentRequest struct {
	ParentID *int64 `json:"parent_id"`
	Author   string `json:"author"`
	Body     string `json:"body"`
}

// SearchCommentsRequest запрос на поиск комментариев.
type SearchCommentsRequest struct {
	Query string `json:"query"`
}

// GetCommentsQuery параметры получения комментариев.
type GetCommentsQuery struct {
	ParentID *int64
	Page     int
	Limit    int
	SortBy   string
	Order    string
}

// CommentResponse ответ с одним комментарием.
type CommentResponse struct {
	Comment model.Comment `json:"comment"`
}

// CommentsResponse ответ со списком комментариев.
type CommentsResponse struct {
	Comments []model.Comment `json:"comments"`
	Page     int             `json:"page"`
	Limit    int             `json:"limit"`
	Total    int             `json:"total"`
}

// ErrorResponse ответ с ошибкой.
type ErrorResponse struct {
	Error string `json:"error"`
}
