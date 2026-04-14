package model

import "time"

// Image описывает изображение и результат его обработки.
type Image struct {
	ID                 string
	OriginalName       string
	ContentType        string
	Status             string
	OperationResize    bool
	OperationThumb     bool
	OperationWatermark bool
	OriginalPath       string
	ProcessedPath      string
	ThumbPath          string
	ErrorText          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
