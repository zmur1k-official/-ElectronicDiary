package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// handleTeacherScheduleCreate добавляет структурную запись урока.
func (s *Server) handleTeacherScheduleCreate(w http.ResponseWriter, r *http.Request, teacher User) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	type request struct {
		ClassName string `json:"className"`
		Subject   string `json:"subject"`
		Weekday   string `json:"weekday"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
		Room      string `json:"room"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.ClassName == "" || req.Subject == "" || req.Weekday == "" || req.StartTime == "" || req.EndTime == "" {
		writeError(w, http.StatusBadRequest, "className, subject, weekday, startTime, endTime are required")
		return
	}
	entry := s.store.addSchedule(ScheduleEntry{
		ClassName: strings.TrimSpace(req.ClassName),
		Subject:   strings.TrimSpace(req.Subject),
		Weekday:   strings.TrimSpace(req.Weekday),
		StartTime: strings.TrimSpace(req.StartTime),
		EndTime:   strings.TrimSpace(req.EndTime),
		Room:      strings.TrimSpace(req.Room),
		TeacherID: teacher.ID,
	})
	writeJSON(w, http.StatusCreated, entry)
}

// handleTeacherStudents возвращает список учеников для выбора в интерфейсе учителя.
func (s *Server) handleTeacherStudents(w http.ResponseWriter, r *http.Request, _ User) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, s.store.listStudentsSortedByClass())
}

// handleTeacherGradeCreate выставляет оценку ученику.
func (s *Server) handleTeacherGradeCreate(w http.ResponseWriter, r *http.Request, teacher User) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	type request struct {
		StudentID int64  `json:"studentId"`
		Subject   string `json:"subject"`
		Value     int    `json:"value"`
		Comment   string `json:"comment"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	student, ok := s.store.getUser(req.StudentID)
	if !ok || student.Role != RoleStudent {
		writeError(w, http.StatusBadRequest, "student not found")
		return
	}
	if req.Subject == "" || req.Value < 1 || req.Value > 5 {
		writeError(w, http.StatusBadRequest, "subject and grade value 1..5 are required")
		return
	}
	g := s.store.addGrade(Grade{
		StudentID: req.StudentID,
		Subject:   strings.TrimSpace(req.Subject),
		Value:     req.Value,
		Comment:   strings.TrimSpace(req.Comment),
		TeacherID: teacher.ID,
		Date:      time.Now().Format("2006-01-02"),
	})
	writeJSON(w, http.StatusCreated, g)
}

// handleTeacherHomeworkCreate добавляет домашнее задание для класса.
func (s *Server) handleTeacherHomeworkCreate(w http.ResponseWriter, r *http.Request, teacher User) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	type request struct {
		ClassName   string `json:"className"`
		Subject     string `json:"subject"`
		Description string `json:"description"`
		DueDate     string `json:"dueDate"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.ClassName == "" || req.Subject == "" || req.Description == "" || req.DueDate == "" {
		writeError(w, http.StatusBadRequest, "className, subject, description, dueDate are required")
		return
	}
	hw := s.store.addHomework(Homework{
		ClassName:   strings.TrimSpace(req.ClassName),
		Subject:     strings.TrimSpace(req.Subject),
		Description: strings.TrimSpace(req.Description),
		DueDate:     strings.TrimSpace(req.DueDate),
		TeacherID:   teacher.ID,
	})
	writeJSON(w, http.StatusCreated, hw)
}
