package main

import "net/http"

// handleStudentSchedule возвращает фото расписания для класса текущего ученика.
func (s *Server) handleStudentSchedule(w http.ResponseWriter, r *http.Request, student User) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	photo, ok := s.store.getSchedulePhotoByClass(student.ClassName)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{})
		return
	}
	writeJSON(w, http.StatusOK, photo)
}

// handleStudentGrades возвращает оценки текущего ученика.
func (s *Server) handleStudentGrades(w http.ResponseWriter, r *http.Request, student User) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, s.store.listGradesByStudent(student.ID))
}

// handleStudentHomework возвращает домашние задания класса текущего ученика.
func (s *Server) handleStudentHomework(w http.ResponseWriter, r *http.Request, student User) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, s.store.listHomeworkByClass(student.ClassName))
}
