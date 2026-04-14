package dto

// UploadResponse возвращается после постановки изображения в очередь.
type UploadResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ImageStatusResponse возвращает состояние обработки изображения.
type ImageStatusResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	ImageURL  string `json:"image_url,omitempty"`
	ThumbURL  string `json:"thumb_url,omitempty"`
	ErrorText string `json:"error_text,omitempty"`
}

// DeleteResponse возвращается после удаления изображения.
type DeleteResponse struct {
	Result string `json:"result"`
}

// ProcessImageMessage описывает сообщение, которое отправляется в Kafka.
type ProcessImageMessage struct {
	ID           string `json:"id"`
	OriginalPath string `json:"original_path"`
	Resize       bool   `json:"resize"`
	Thumb        bool   `json:"thumb"`
	Watermark    bool   `json:"watermark"`
}
