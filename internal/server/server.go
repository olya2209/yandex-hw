package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"

	chi "github.com/go-chi/chi/v5"

	srv "github.com/olya2209/yandex-hw/internal/service"
)

const (
	addr = "localhost:8080"
)

type Server struct {
	route *chi.Mux
	su    srv.CaseURL
}

func NewServer() *Server {
	server := &Server{
		route: chi.NewRouter(),
		su:    srv.NewService(),
	}
	server.router()
	return server
}

func (s *Server) router() {
	// Подключаем мидлварь
	s.route.Use(PathValidationMiddleware)

	s.route.HandleFunc("/", s.SetURL)
	s.route.HandleFunc("/{id}", s.GetURL)
}

// PathValidationMiddleware проверяет корректность пути и возвращает 400, если путь неверный
func PathValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		// Проверяем, является ли путь допустимым
		if path != "/" || !slices.Contains([]string{http.MethodGet, http.MethodPost}, method) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Если путь верный, продолжаем обработку
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Run() {
	fmt.Println("server started ...")
	if err := http.ListenAndServe(addr, s.route); err != nil {
		log.Fatalln(err)
	}
}

func (s *Server) SetURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "method must be POST", http.StatusBadRequest)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "text/plain" {
		http.Error(res, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "cannot read body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	hash, err := s.su.SetURL(string(body))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://localhost:8080/" + hash))
}

func (s *Server) GetURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "method must be GET", http.StatusBadRequest)
		return
	}

	hash := strings.TrimPrefix(req.URL.Path, "/")

	url, err := s.su.GetURL(hash)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
