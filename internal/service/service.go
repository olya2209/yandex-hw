package service

import (
	"crypto/sha256"
	"fmt"
	"strings"

	repo "github.com/olya2209/yandex-hw/internal/repository"
)

const sizeHash = 16

type CaseURL interface {
	SetURL(url string) (string, error)
	GetURL(hash string) (string, error)
}

type Service struct {
	repo repo.Repository
}

func NewService() CaseURL {
	return &Service{
		repo: repo.NewStorage(),
	}
}

func (s *Service) SetURL(originalURL string) (string, error) {
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		return "", fmt.Errorf("incorrect url")
	}

	hash := sha256.Sum256([]byte(originalURL))
	shortHash := fmt.Sprintf("%x", hash[:sizeHash])
	err := s.repo.Set(originalURL, shortHash)

	return shortHash, err
}

func (s *Service) GetURL(hash string) (string, error) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return "", fmt.Errorf("incorrect id")
	}
	return s.repo.Get(hash)
}
