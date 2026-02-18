package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

// hashPassword вычисляет SHA-256 хеш пароля.
func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

// withAuth проверяет Bearer-токен и роль пользователя перед вызовом обработчика.
func (s *Server) withAuth(next func(http.ResponseWriter, *http.Request, User), roles ...Role) http.HandlerFunc {
	allowed := map[Role]bool{}
	for _, role := range roles {
		allowed[role] = true
	}

	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}
		user, ok := s.store.userByToken(token)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		if len(allowed) > 0 && !allowed[user.Role] {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		next(w, r, user)
	}
}
