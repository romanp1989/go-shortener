package storage

import (
	"errors"
	"fmt"
)

// ErrConflict data already exists
var ErrConflict = errors.New("данные уже существуют")

// URLConflictError structure for url conflict error, if URL already exists in DB
type URLConflictError struct {
	URL string
	Err error
}

// Error function for create error URL save
func (ue *URLConflictError) Error() string {
	return fmt.Sprintf("ошибка добавления URL %v: %v", ue.URL, ue.Err)
}

// NewURLConflictError factory for create DB errors
func NewURLConflictError(url string, err error) error {
	return &URLConflictError{
		URL: url,
		Err: err,
	}
}

// AlreadyDeleted structure for errors, if URL already deleted
type AlreadyDeleted struct {
	URL string
}

// Error function for errors, if URL already deleted
func (ad *AlreadyDeleted) Error() string {
	return fmt.Sprintf("URLs %v уже удален", ad.URL)
}

// NewURLConflictError factory for create DB errors, if URL already deleted
func NewAlreadyDeletedError(url string) error {
	return &AlreadyDeleted{
		URL: url,
	}
}
