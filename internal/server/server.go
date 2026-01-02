package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olya2209/yandex-hw/internal/config"
	"github.com/olya2209/yandex-hw/internal/logger"
	srv "github.com/olya2209/yandex-hw/internal/service"
	"go.uber.org/zap"
)

type Server struct {
	cfg   *config.Config
	route *chi.Mux
	su    srv.CaseURL
	log   zap.SugaredLogger
}

func NewServer(cfg *config.Config, sugar zap.SugaredLogger) *Server {
	server := &Server{
		cfg:   cfg,
		route: chi.NewRouter(),
		su:    srv.NewService(),
		log:   sugar,
	}
	server.router()
	return server
}

// хендлер для /ping
func (s *Server) SetURLHandler() http.Handler {
	return http.HandlerFunc(s.SetURL)
}

func (s *Server) GetURLHandler() http.Handler {
	return http.HandlerFunc(s.GetURL)
}

func (s *Server) router() {
	s.route.Post("/", logger.WithLogging(s.SetURLHandler(), s.log))
	//Сведения об ответах должны содержать код статуса и размер содержимого ответа
	s.route.Get("/{id}", logger.WithLogging(s.GetURLHandler(), s.log))
}

func (s *Server) Run() {
	fmt.Println("server started ...")
	//Все сообщения логгера должны быть на уровне Info
	//Сведения о запросах должны содержать URI, метод запроса и время, затраченное на его выполнение.
	if err := http.ListenAndServe(s.cfg.Opts.Addr, s.route); err != nil {
		s.log.Fatalw(err.Error(), "event", "start server")
	}
}

func (s *Server) SetURL(res http.ResponseWriter, req *http.Request) {
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
	res.Write([]byte(s.cfg.Opts.BaseURL + "/" + hash))
}

func (s *Server) GetURL(res http.ResponseWriter, req *http.Request) {
	pathURL := chi.URLParam(req, "id")
	if pathURL == "" {
		pathURL = req.URL.Path
	}
	hash := strings.TrimPrefix(pathURL, "/")

	url, err := s.su.GetURL(hash)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
