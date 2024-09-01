package storage

import (
	"errors"
	"fmt"
)

var (
	ErrConflict = errors.New("данные уже существуют")
)

type URLConflictError struct {
	URL string
	Err error
}

func (ue *URLConflictError) Error() string {
	return fmt.Sprintf("ошибка добавления URL %v: %v", ue.URL, ue.Err)
}

func NewURLConflictError(url string, err error) error {
	return &URLConflictError{
		URL: url,
		Err: err,
	}
}

type AlreadyDeleted struct {
	URLs []string
}

func (ad *AlreadyDeleted) Error() string {
	return fmt.Sprintf("URLs %v уже удален", ad.URLs)
}

func NewAlreadyDeletedError(urls []string, err error) error {
	return &AlreadyDeleted{
		URLs: urls,
	}
}
