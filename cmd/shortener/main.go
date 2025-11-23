package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

const (
	route            = "http://"
	methodNotAllowed = "Метод не поддерживается"
	incorrectRoute   = "Некорректный маршрут"
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

func main() {
	ResetGlobalMap()
	registrationRoutes()

	log.Println("Запуск сервера на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func registrationRoutes() {
	http.HandleFunc("/", createShortURLHandler)
	http.HandleFunc("/{id}", redirectHandler)
	http.HandleFunc("/.*", badRouteHandler)
}

func createShortURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	handleCreateShortURL(w, r)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	path := r.URL.Path
	// Разбираем путь и получаем ID
	parts := strings.Split(path, "/")
	if len(parts) != 2 || len(parts[1]) != 8 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if match, _ := regexp.MatchString("[0-9]", parts[1]); !match {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	handleRedirect(w, r)
}

func badRouteHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
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
