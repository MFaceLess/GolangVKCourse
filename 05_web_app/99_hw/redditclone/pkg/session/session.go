package session

import (
	"errors"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	ErrJWTSecretNotSet = errors.New("JWT_SECRET is not set in environment")
)

type User struct {
	Username string `json:"username"`
	UserID   string `json:"id"`
}

type Claims struct {
	User User `json:"user"`
	jwt.StandardClaims
}

func GenerateJWT(username, userID string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		User: User{
			Username: username,
			UserID:   userID,
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	jwtKey, err := ReadJWTSecretKey()
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}

func ReadJWTSecretKey() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, ErrJWTSecretNotSet
	}

	return []byte(secret), nil
}
