package main

import (
	"encoding/json"
	"net/http"
	"strconv"
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

// handleTeacherStudents возвращает список учеников, отсортированный по классу.
func (s *Server) handleTeacherStudents(w http.ResponseWriter, r *http.Request, _ User) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, s.store.listStudentsSortedByClass())
}

// handleTeacherSubject позволяет учителю закрепить за собой предмет.
func (s *Server) handleTeacherSubject(w http.ResponseWriter, r *http.Request, teacher User) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]string{
			"subject": s.store.getTeacherSubject(teacher.ID),
		})
		return
	case http.MethodPost:
		type request struct {
			Subject string `json:"subject"`
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		subject := strings.TrimSpace(req.Subject)
		if subject == "" {
			writeError(w, http.StatusBadRequest, "subject is required")
			return
		}
		subject = s.store.setTeacherSubject(teacher.ID, subject)
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "saved",
			"subject": subject,
		})
		return
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
}

// handleTeacherGradesJournal возвращает оценки учителя по его предмету за диапазон дат.
func (s *Server) handleTeacherGradesJournal(w http.ResponseWriter, r *http.Request, teacher User) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	subject := strings.TrimSpace(s.store.getTeacherSubject(teacher.ID))
	if subject == "" {
		writeError(w, http.StatusBadRequest, "teacher subject is not set")
		return
	}
	from := strings.TrimSpace(r.URL.Query().Get("from"))
	to := strings.TrimSpace(r.URL.Query().Get("to"))
	if from != "" {
		if _, err := time.Parse("2006-01-02", from); err != nil {
			writeError(w, http.StatusBadRequest, "from must be YYYY-MM-DD")
			return
		}
	}
	if to != "" {
		if _, err := time.Parse("2006-01-02", to); err != nil {
			writeError(w, http.StatusBadRequest, "to must be YYYY-MM-DD")
			return
		}
	}
	rows := s.store.listGradesByTeacherSubjectDateRange(teacher.ID, subject, from, to)
	writeJSON(w, http.StatusOK, map[string]any{
		"subject": subject,
		"grades":  rows,
	})
}

// handleTeacherGradeCreate работает с оценками: получить для ученика или добавить новую.
func (s *Server) handleTeacherGradeCreate(w http.ResponseWriter, r *http.Request, teacher User) {
	switch r.Method {
	case http.MethodGet:
		studentIDStr := strings.TrimSpace(r.URL.Query().Get("studentId"))
		studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
		if err != nil || studentID <= 0 {
			writeError(w, http.StatusBadRequest, "valid studentId is required")
			return
		}
		student, ok := s.store.getUser(studentID)
		if !ok || student.Role != RoleStudent {
			writeError(w, http.StatusBadRequest, "student not found")
			return
		}
		writeJSON(w, http.StatusOK, s.store.listGradesByStudent(studentID))
		return
	case http.MethodPost:
		// handled below
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	subject := strings.TrimSpace(s.store.getTeacherSubject(teacher.ID))
	if subject == "" {
		writeError(w, http.StatusBadRequest, "teacher subject is not set")
		return
	}

	type request struct {
		StudentID int64  `json:"studentId"`
		Value     int    `json:"value"`
		Comment   string `json:"comment"`
		Date      string `json:"date"`
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
	if req.Value < 1 || req.Value > 5 {
		writeError(w, http.StatusBadRequest, "grade value 1..5 is required")
		return
	}

	date := strings.TrimSpace(req.Date)
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		writeError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}

	g := s.store.addGrade(Grade{
		StudentID: req.StudentID,
		Subject:   subject,
		Value:     req.Value,
		Comment:   strings.TrimSpace(req.Comment),
		TeacherID: teacher.ID,
		Date:      date,
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
