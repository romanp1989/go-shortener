package storage

import (
	"bufio"
	"encoding/json"
	"github.com/romanp1989/go-shortener/internal/models"
	"io"
	"log"
	"os"
	"path/filepath"
)

type FileStorage struct {
	FileStoragePath string
}

func NewFileStorage(path string) *FileStorage {
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			log.Fatalf("Ошибка: %s", err)
			return nil
		}
	}

	return &FileStorage{
		FileStoragePath: path,
	}
}

func (s *FileStorage) Save(originalURL string, shortURL string) error {
	var urlStorage models.StorageURL

	file, err := os.OpenFile(s.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Ошибка при открытии: %s", err)
		return nil
	}

	defer file.Close()

	urlStorage.OriginalURL, urlStorage.ShortURL = originalURL, shortURL
	encoder := json.NewEncoder(file)

	return encoder.Encode(urlStorage)
}

func (s *FileStorage) Get(inputURL string) string {
	var (
		read       [][]byte
		storageURL []models.StorageURL
	)
	file, err := os.OpenFile(s.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Ошибка при открытии: %s", err)
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

	for _, ur := range storageURL {
		if ur.ShortURL == inputURL {
			url := ur.OriginalURL
			return url
		}
	}

	for _, ur := range storageURL {
		if ur.OriginalURL == inputURL {
			url := ur.ShortURL
			return url
		}
	}

	return ""
}
