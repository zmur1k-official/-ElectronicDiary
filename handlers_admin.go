package main

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

// handleAdminUsers обрабатывает список пользователей и создание пользователя админом.
func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request, _ User) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.store.listUsers())
	case http.MethodPost:
		s.handleRegister(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleAdminUserByID удаляет пользователя по ID.
func (s *Server) handleAdminUserByID(w http.ResponseWriter, r *http.Request, _ User) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if !s.store.deleteUser(id) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// handleAdminScheduleImport загружает фото расписания для выбранного класса.
func (s *Server) handleAdminScheduleImport(w http.ResponseWriter, r *http.Request, _ User) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}
	className := normalizeClassName(r.FormValue("className"))
	if className == "" {
		writeError(w, http.StatusBadRequest, "className is required")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(io.LimitReader(file, 20<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read uploaded file")
		return
	}
	contentType := http.DetectContentType(raw)
	if !strings.HasPrefix(contentType, "image/") {
		writeError(w, http.StatusBadRequest, "uploaded file must be an image")
		return
	}
	photo := s.store.setSchedulePhoto(className, contentType, raw)
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "imported",
		"className": photo.ClassName,
	})
}

// handleAdminScheduleClear очищает все данные расписания.
func (s *Server) handleAdminScheduleClear(w http.ResponseWriter, r *http.Request, _ User) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.store.clearSchedule()
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "schedule cleared",
	})
}

// handleAdminScheduleStats возвращает статистику фото расписаний по классам.
func (s *Server) handleAdminScheduleStats(w http.ResponseWriter, r *http.Request, _ User) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	classes := s.store.schedulePhotoStats()
	writeJSON(w, http.StatusOK, map[string]any{
		"total":   len(classes),
		"classes": classes,
	})
}
