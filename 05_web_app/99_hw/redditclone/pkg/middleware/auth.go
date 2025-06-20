package middleware

import (
	"context"
	"errors"
	"net/http"
	"redditclone/pkg/session"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var (
	ErrMissingToken       = errors.New("передан пустой токен")
	ErrReadSecretJWTToken = errors.New("ошибка при чтении secretJWTToken")
	ErrInvalidJWTToken    = errors.New("невалидный токен")
)

type ContextKey string

const claimsCtxKey ContextKey = "claims"

func JWTMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStringWithMethod := r.Header.Get("Authorization")
		if tokenStringWithMethod == "" {
			http.Error(w, ErrMissingToken.Error(), http.StatusUnauthorized)
			return
		}

		pair := strings.Split(tokenStringWithMethod, " ")

		if len(pair) <= 1 {
			http.Error(w, ErrInvalidJWTToken.Error(), http.StatusUnauthorized)
			return
		}

		tokenString := pair[1]

		jwtKey, err := session.ReadJWTSecretKey()
		if err != nil {
			http.Error(w, ErrReadSecretJWTToken.Error(), http.StatusInternalServerError)
			return
		}

		claims := &session.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			method, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok || method.Alg() != "HS256" {
				return nil, ErrInvalidJWTToken
			}
			return jwtKey, nil
		})

		if token == nil || err != nil || !token.Valid {
			http.Error(w, ErrInvalidJWTToken.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
