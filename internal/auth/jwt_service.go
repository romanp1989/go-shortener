package auth

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"time"
)

const tokenExp = time.Hour * 3

type JWTService struct {
	secretKey string
	expire    time.Duration
	TokenName string
}

// Claims JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID *uuid.UUID
}

func NewJwtService(secretKey string) *JWTService {
	return &JWTService{
		secretKey: secretKey,
		expire:    tokenExp,
		TokenName: "auth",
	}
}

// NewCookie Function add new authorization cookie
func (j *JWTService) NewCookie(w http.ResponseWriter, userID *uuid.UUID, secretKey string) {

	token, err := j.CreateToken(userID, secretKey)
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
func (j *JWTService) CreateToken(userID *uuid.UUID, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// EnsureRandom Function generate random uuid
func (j *JWTService) EnsureRandom() (res uuid.UUID) {
	return uuid.Must(uuid.NewV4())
}

// GetUserID Function for get userID from auth token
func (j *JWTService) GetUserID(tokenString string) (*uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неизвестный алгоритм подписи: %v", t.Header["alg"])
			}
			return []byte(j.secretKey), nil
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
