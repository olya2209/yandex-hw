package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"

	"github.com/olya2209/yandex-hw/internal/config"
	repo "github.com/olya2209/yandex-hw/internal/repository"
	"go.uber.org/zap"
)

const sizeHash = 16

type CaseURL interface {
	SetURL(url string) (string, error)
	GetURL(hash string) (string, error)
	Ping(ctx context.Context) error
}

type Service struct {
	repo repo.Repository
}

func NewService(cfg *config.Config, db *sql.DB, logger zap.SugaredLogger) (CaseURL, error) {
	repo, err := repo.NewStorage(cfg, db, logger)
	if err != nil {
		return nil, err
	}
	return &Service{
		repo: repo,
	}, nil
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

func (s *Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
