package auth

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"github.com/romanp1989/go-shortener/internal/config"
	"net/http"
	"time"
)

const tokenExp = time.Hour * 3

// Claims JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID *uuid.UUID
}

// NewCookie Function add new authorization cookie
func NewCookie(w http.ResponseWriter, userID *uuid.UUID) {

	token, err := CreateToken(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:  "auth",
		Value: token,
		Path:  "/",
	}

	http.SetCookie(w, cookie)
}

// CreateToken Function create auth token with userID
func CreateToken(userID *uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(config.Options.FlagSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Validation Function for validate auth token
func Validation(tokenString string) bool {

	token, err := jwt.Parse(tokenString,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(config.Options.FlagSecretKey), nil
		})

	if err != nil {
		return false
	}

	if !token.Valid {
		return false
	}
	return true
}

// EnsureRandom Function generate random uuid
func EnsureRandom() (res uuid.UUID) {
	return uuid.Must(uuid.NewV4())
}

// GetUserID Function for get userID from auth token
func GetUserID(tokenString string) (*uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неизвестный алгоритм подписи: %v", t.Header["alg"])
			}
			return []byte(config.Options.FlagSecretKey), nil
		})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	fmt.Println("токен валидный", claims.UserID)
	return claims.UserID, nil
}
