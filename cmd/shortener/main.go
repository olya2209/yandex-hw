package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Map для хранения оригиналов и коротких версий URL
var urlsMap map[string]string

const route = "http://"

func initArgs() {
	urlsMap = make(map[string]string)
	rand.Seed(time.Now().UnixNano())
}

//type ShortLink struct {
//	URL     string `json:"url"`
//	ShortID string `json:"short_id"`
//}

func generateRandomStr(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", errors.New("ошибка генерации случайного значения")
	}
	s := base64.StdEncoding.EncodeToString(b)
	return s[:n], nil
}

// POST-обработчик: создает новую запись с коротким URL
func handleCreateShortURL(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil || len(body) == 0 {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}

	originalURL := bytes.TrimSpace(body)

	// Генерация уникального короткого идентификатора
	shortID, err := generateRandomStr(8)
	if err != nil {
		http.Error(w, "Ошибка генерации сокращённого URL", http.StatusInternalServerError)
		return
	}

	// Сохраняем ссылку в карте
	urlsMap[shortID] = string(originalURL)

	// Возвращаем успешный ответ с новым URL
	responseURL := route + r.Host + "/" + shortID
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(responseURL))
}

// GET-обработчик: переадресация по короткому URL
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	log.Println("Получили Get request:", id)
	// Извлекаем оригинал из карты
	redirectURL, ok := urlsMap[id]
	if !ok {
		http.Error(w, "Ссылка не найдена", http.StatusNotFound)
		return
	}

	// Переадресация
	w.Header().Set("Location", redirectURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	initArgs()
	// Регистрация обработчиков маршрутов
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateShortURL(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusBadRequest)
		}
	})

	http.HandleFunc("/{id}", handleRedirect)

	// Запускаем сервер
	log.Println("Запуск сервера на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
