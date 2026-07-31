package handlers

import (
	"context"
	"net/http"
	"news-restapi/utils"
	"strings"

	"github.com/golang-jwt/jwt"
)

type contextKey string

const AuthorIDKey contextKey = "author_id"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.SendError(w, http.StatusUnauthorized, "Missing Authorization header", nil)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.SendError(w, http.StatusUnauthorized, "Invalid Authorization format", nil)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})
		if err != nil || !token.Valid {
			utils.SendError(w, http.StatusUnauthorized, "Invalid or expired token", nil)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.SendError(w, http.StatusUnauthorized, "Invalid token claims", nil)
			return
		}

		authorID := int(claims["author_id"].(float64))

		ctx := context.WithValue(r.Context(), AuthorIDKey, authorID)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
