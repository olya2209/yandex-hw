package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		ResponseData        *ResponseData
	}
	// берём структуру для хранения сведений об ответе
	ResponseData struct {
		status int
		size   int
	}
)

func NewLogger(addr string) (zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return zap.SugaredLogger{}, err
	}

	defer logger.Sync()

	// делаем регистратор SugaredLogger
	sugar := *logger.Sugar()
	sugar.Infow(
		"Starting server",
		"addr", addr,
	)

	return sugar, nil
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.status = statusCode // захватываем код статуса
}

func WithLogging(h http.Handler, sugar zap.SugaredLogger) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &ResponseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			ResponseData:   responseData,
		}
		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(start)

		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)
	}
	return logFn
}
