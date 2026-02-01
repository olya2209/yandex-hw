package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/olya2209/yandex-hw/internal/config"
)

type Repository interface {
	Get(hash string) (string, error)
	Set(url, hash string) error
	Ping(ctx context.Context) error
}

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	cfg         *config.Config
	s           []URLRecord
	mu          sync.RWMutex
	db          *sql.DB
	memoryCache map[string]string
	hasFile     bool
}

func NewStorage(cfg *config.Config, db *sql.DB, loggger zap.SugaredLogger) (Repository, error) {
	fs := &FileStorage{
		cfg:         cfg,
		db:          db,
		memoryCache: make(map[string]string),
		hasFile:     cfg.Opts.StorageFile != "",
	}

	err := fs.loadFromFile(loggger)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func (fs *FileStorage) Get(hash string) (string, error) {
	if fs.db != nil {
		return fs.getPsql(hash)
	}
	if fs.hasFile {
		return fs.getFromFile(hash)
	}
	return fs.getMemory(hash)
}

func (fs *FileStorage) Set(url, hash string) error {
	if fs.db != nil {
		return fs.setPsql(url, hash)
	}
	if fs.hasFile {
		return fs.setInFile(url, hash)
	}
	return fs.setMemory(url, hash)
}

func (fs *FileStorage) saveToFile() error {
	dir := filepath.Dir(fs.cfg.Opts.StorageFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(fs.s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(fs.cfg.Opts.StorageFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (fs *FileStorage) loadFromFile(loggger zap.SugaredLogger) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.hasFile {
		return nil
	}

	if _, err := os.Stat(fs.cfg.Opts.StorageFile); os.IsNotExist(err) {
		loggger.Warn("File does not exist")
		return nil
	}

	data, err := os.ReadFile(fs.cfg.Opts.StorageFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &fs.s); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	if fs.db == nil {
		return fmt.Errorf("db is not init")
	}
	return fs.db.PingContext(ctx)
}

func (fs *FileStorage) setPsql(url, hash string) error {
	_, err := fs.db.Exec(querySetURL, url, hash)
	return err
}

func (fs *FileStorage) getPsql(hash string) (string, error) {
	var originalURL string
	err := fs.db.QueryRow(queryGetURL, hash).Scan(&originalURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("URL not found")
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	return originalURL, nil
}

func (fs *FileStorage) setInFile(url, hash string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.s = append(fs.s, URLRecord{
		UUID:        strconv.Itoa(len(fs.s)),
		ShortURL:    hash,
		OriginalURL: url,
	})
	return fs.saveToFile()
}

func (fs *FileStorage) getFromFile(hash string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	item, exist := lo.Find(fs.s, func(item URLRecord) bool {
		return item.ShortURL == hash
	})
	if !exist {
		return "", fmt.Errorf("%s not found", hash)
	}

	return item.OriginalURL, nil
}

func (fs *FileStorage) getMemory(hash string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	url, exist := fs.memoryCache[hash]
	if !exist {
		return "", fmt.Errorf("%s not found", hash)
	}
	return url, nil
}

func (fs *FileStorage) setMemory(url, hash string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.memoryCache[hash] = url
	return nil
}
