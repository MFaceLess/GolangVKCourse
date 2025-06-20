package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"

	"redditclone/pkg/session"
)

var (
	ErrMissingToken       = errors.New("передан пустой токен")
	ErrReadSecretJWTToken = errors.New("ошибка при чтении secretJWTToken")
	ErrInvalidJWTToken    = errors.New("невалидный токен")
	ErrDatabaseRead       = errors.New("ошибка при чтении из бд")
)

type ContextKey string

const claimsCtxKey ContextKey = "claims"

func JWTMiddleWare(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

			var storedToken string
			err = db.QueryRow(
				"SELECT jwt FROM sessions WHERE login = ? AND expires_at > NOW()",
				claims.User.Username,
			).Scan(&storedToken)

			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, ErrInvalidJWTToken.Error(), http.StatusUnauthorized)
				} else {
					http.Error(w, ErrDatabaseRead.Error(), http.StatusInternalServerError)
				}
				return
			}

			if storedToken != tokenString {
				http.Error(w, ErrInvalidJWTToken.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
