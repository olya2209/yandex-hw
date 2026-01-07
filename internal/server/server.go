package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olya2209/yandex-hw/internal/config"
	"github.com/olya2209/yandex-hw/internal/logger"
	"github.com/olya2209/yandex-hw/internal/model"
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

func (s *Server) SetURLHandler() http.Handler {
	return http.HandlerFunc(s.SetURL)
}

func (s *Server) GetURLHandler() http.Handler {
	return http.HandlerFunc(s.GetURL)
}

func (s *Server) ShortURLHandler() http.Handler {
	return http.HandlerFunc(s.ShortURL)
}

func (s *Server) router() {
	s.route.Post("/", logger.WithLogging(s.SetURLHandler(), s.log))
	s.route.Post("/api/shorten", logger.WithLogging(s.ShortURLHandler(), s.log))
	//Сведения об ответах должны содержать код статуса и размер содержимого ответа
	s.route.Get("/{id}", logger.WithLogging(s.GetURLHandler(), s.log))
}

func (s *Server) Run() {
	fmt.Println("server started on port: " + s.cfg.Opts.Addr)
	//Все сообщения логгера должны быть на уровне Info
	//Сведения о запросах должны содержать URI, метод запроса и время, затраченное на его выполнение.
	if err := http.ListenAndServe(s.cfg.Opts.Addr, s.route); err != nil {
		s.log.Fatalw(err.Error(), "event", "start server")
	}
}

func (s *Server) SetURL(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if contentType != "text/plain" {
		errM := "Content-Type must be text/plain"
		s.log.Errorln(errM)
		http.Error(res, errM, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		errM := "cannot read body"
		s.log.Errorln(errM)
		http.Error(res, errM, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	hash, err := s.su.SetURL(string(body))
	if err != nil {
		s.log.Errorln(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(s.cfg.Opts.BaseURL + "/" + hash))
}

func (s *Server) ShortURL(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		errM := "Content-Type must be application/json"
		s.log.Errorln()
		http.Error(res, errM, http.StatusBadRequest)
		return
	}

	var r models.Request
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&r); err != nil {
		errM := "cannot decode request JSON body"
		s.log.Errorln(errM)
		http.Error(res, errM, http.StatusBadRequest)
		return
	}

	hash, err := s.su.SetURL(r.Url)
	if err != nil {
		s.log.Errorln(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// заполняем модель ответа
	resp := models.Response{
		ResultURL: hash,
	}

	res.Header().Set("Content-Type", "application/json")

	if res.Header().Get("Content-Type") != "application/json" {
		errM := "Content-Type must be application/json"
		s.log.Errorln(errM)
		http.Error(res, errM, http.StatusBadRequest)
		return
	}

	// сериализуем ответ сервера
	enc := json.NewEncoder(res)
	if err = enc.Encode(resp); err != nil {
		errM := "error encoding response"
		s.log.Errorln(errM)
		http.Error(res, errM, http.StatusBadRequest)
		return
	}
	s.log.Infoln("sending HTTP 200 response")
}
func (s *Server) GetURL(res http.ResponseWriter, req *http.Request) {
	pathURL := chi.URLParam(req, "id")
	if pathURL == "" {
		pathURL = req.URL.Path
	}
	hash := strings.TrimPrefix(pathURL, "/")

	url, err := s.su.GetURL(hash)
	if err != nil {
		errM := err.Error()
		s.log.Errorln(errM)
		http.Error(res, errM, http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
