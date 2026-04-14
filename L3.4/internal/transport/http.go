package transport

import (
	"imageProcessor/internal/handler"
	"net/http"
	"strings"
)

// NewRouter создает HTTP-маршруты.
func NewRouter(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./static")))
	mux.Handle("/files/processed/", http.StripPrefix("/files/processed/", http.FileServer(http.Dir("./uploads/processed"))))
	mux.Handle("/files/thumbs/", http.StripPrefix("/files/thumbs/", http.FileServer(http.Dir("./uploads/thumbs"))))

	mux.HandleFunc("/upload", h.UploadImage)

	mux.HandleFunc("/image/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/image/")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.GetImage(w, r, id)
		case http.MethodDelete:
			h.DeleteImage(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
