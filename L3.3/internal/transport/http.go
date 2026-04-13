package transport

import (
	"net/http"

	"commentTree/internal/handler"
)

// NewHTTPHandler регистрирует HTTP-маршруты приложения.
func NewHTTPHandler(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/comments", commentsEntry(h))
	mux.HandleFunc("/comments/all", h.GetAllComments)
	mux.HandleFunc("/comments/search", h.SearchComments)
	mux.HandleFunc("/comments/", h.DeleteComment)

	return mux
}

func commentsEntry(h *handler.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateComment(w, r)
		case http.MethodGet:
			h.GetComments(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "static/index.html")
}
