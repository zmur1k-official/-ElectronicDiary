package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// handleRegister регистрирует нового пользователя.
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	type request struct {
		FullName  string `json:"fullName"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		Role      Role   `json:"role"`
		ClassName string `json:"className"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.FullName == "" || req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "fullName, email and password are required")
		return
	}
	if req.Role != RoleAdmin && req.Role != RoleTeacher && req.Role != RoleStudent {
		writeError(w, http.StatusBadRequest, "role must be admin|teacher|student")
		return
	}
	if req.Role == RoleStudent && strings.TrimSpace(req.ClassName) == "" {
		writeError(w, http.StatusBadRequest, "className is required for student")
		return
	}

	user, err := s.store.createUser(User{
		FullName:     strings.TrimSpace(req.FullName),
		Email:        strings.TrimSpace(req.Email),
		PasswordHash: hashPassword(req.Password),
		Role:         req.Role,
		ClassName:    normalizeClassName(req.ClassName),
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

// handleLogin выполняет аутентификацию и выдает токен.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	u, ok := s.store.findUserByEmail(req.Email)
	if !ok || u.PasswordHash != hashPassword(req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := s.store.createToken(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  u,
	})
}

// handleMe возвращает профиль текущего авторизованного пользователя.
func (s *Server) handleMe(w http.ResponseWriter, _ *http.Request, user User) {
	writeJSON(w, http.StatusOK, user)
}
