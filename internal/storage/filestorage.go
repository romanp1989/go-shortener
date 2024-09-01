package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/romanp1989/go-shortener/internal/models"
	"io"
	"log"
	"os"
	"path/filepath"
)

type FileStorage struct {
	FileStoragePath string
}

func NewFileStorage(path string) (*FileStorage, error) {
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return &FileStorage{}, err
		}
	}

	return &FileStorage{FileStoragePath: path}, nil
}

func (s *FileStorage) Save(ctx context.Context, originalURL string, shortURL string, userID *uuid.UUID) (string, error) {
	var urlStorage models.StorageURL

	file, err := os.OpenFile(s.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Ошибка при открытии: %s", err)
		return "", err
	}

	defer file.Close()

	urlStorage.OriginalURL, urlStorage.ShortURL, urlStorage.UserID = originalURL, shortURL, userID
	encoder := json.NewEncoder(file)

	if err = encoder.Encode(urlStorage); err != nil {
		return "", err
	}

	return shortURL, nil
}

func (s *FileStorage) Get(inputURL string) (string, error) {
	var (
		read       [][]byte
		storageURL []models.StorageURL
	)
	file, err := os.OpenFile(s.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		data, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		read = append(read, data)
	}

	for _, line := range read {
		urls := models.StorageURL{}
		err := json.Unmarshal(line, &urls)
		if err == nil {
			storageURL = append(storageURL, urls)
		}
	}

	// Поиск соотвествия полученного url сокращенному или полному url в хранилище, в зависимости от типа запроса.
	// Для POST запросов ищем по OriginalURL, для GET - ShortURL
	for _, ur := range storageURL {
		if ur.ShortURL == inputURL {
			return ur.OriginalURL, nil
		} else if ur.OriginalURL == inputURL {
			return ur.ShortURL, nil
		}
	}

	return "", nil
}

func (s *FileStorage) SaveBatch(ctx context.Context, urls []models.StorageURL, userID *uuid.UUID) ([]string, error) {
	var urlStorage models.StorageURL

	file, err := os.OpenFile(s.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Ошибка при открытии: %s", err)
		return nil, err
	}

	defer file.Close()

	for _, url := range urls {
		urlStorage.OriginalURL, urlStorage.ShortURL, urlStorage.UserID = url.OriginalURL, url.ShortURL, userID
		encoder := json.NewEncoder(file)

		if err = encoder.Encode(urlStorage); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (s *FileStorage) DeleteUrlsBatch(ctx context.Context, userID *uuid.UUID, urls []string) error {
	return nil
}

func (s *FileStorage) GetAllUrlsByUser(ctx context.Context, userID *uuid.UUID) ([]models.StorageURL, error) {
	return nil, nil
}

func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}
