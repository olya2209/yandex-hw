package cmdOld

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
)

const (
	route = "http://"
)

var urlsMap = make(map[string]string)

func SetValue(key, value string) {
	urlsMap[key] = value
}

func GetValue(key string) (string, bool) {
	val, ok := urlsMap[key]
	return val, ok
}

func ResetGlobalMap() {
	for k := range urlsMap {
		delete(urlsMap, k)
	}
}

func mainOld() {
	ResetGlobalMap()

	log.Println("Запуск сервера на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", registrationRoutes()))
}

func registrationRoutes() chi.Router {
	r := chi.NewRouter()

	// Подключаем мидлварь
	r.Use(PathValidationMiddleware)

	r.Post("/", createShortURLHandler)
	r.Get("/{id}", handleRedirect)

	return r
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

func createShortURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	handleCreateShortURL(w, r)
}

func handleCreateShortURL(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil || len(body) == 0 {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}

	originalURL := bytes.TrimSpace(body)

	shortID, err := generateRandomStr(8)
	if err != nil {
		http.Error(w, "Ошибка генерации сокращённого URL", http.StatusInternalServerError)
		return
	}

	SetValue(shortID, string(originalURL))

	responseURL := route + r.Host + "/" + shortID
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(responseURL))
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	//Ненужная валидация, почему-то решила проверять получившийся хэш на соотв формату

	//path := r.URL.Path
	//parts := strings.Split(path, "/")
	//if len(parts) != 2 || len(parts[1]) != 8 {
	//	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	//	return
	//}
	//if match, _ := regexp.MatchString("[0-9]", parts[1]); !match {
	//	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	//	return
	//}

	id := r.URL.Path[1:]
	redirectURL, ok := GetValue(id)
	if !ok {
		http.Error(w, "Ссылка не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", redirectURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func generateRandomStr(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", errors.New("ошибка генерации случайного значения")
	}
	s := base64.StdEncoding.EncodeToString(b)
	return s[:n], nil
}
