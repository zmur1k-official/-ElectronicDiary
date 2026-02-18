package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// routes регистрирует все HTTP-маршруты приложения.
func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/register", s.handleRegister)
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/me", s.withAuth(s.handleMe, RoleAdmin, RoleTeacher, RoleStudent))

	mux.HandleFunc("/api/admin/users", s.withAuth(s.handleAdminUsers, RoleAdmin))
	mux.HandleFunc("/api/admin/users/", s.withAuth(s.handleAdminUserByID, RoleAdmin))
	mux.HandleFunc("/api/admin/schedule/import", s.withAuth(s.handleAdminScheduleImport, RoleAdmin))
	mux.HandleFunc("/api/admin/schedule", s.withAuth(s.handleAdminScheduleClear, RoleAdmin))
	mux.HandleFunc("/api/admin/schedule/stats", s.withAuth(s.handleAdminScheduleStats, RoleAdmin))

	mux.HandleFunc("/api/teacher/schedule", s.withAuth(s.handleTeacherScheduleCreate, RoleTeacher))
	mux.HandleFunc("/api/teacher/grades", s.withAuth(s.handleTeacherGradeCreate, RoleTeacher))
	mux.HandleFunc("/api/teacher/homework", s.withAuth(s.handleTeacherHomeworkCreate, RoleTeacher))
	mux.HandleFunc("/api/teacher/students", s.withAuth(s.handleTeacherStudents, RoleTeacher))

	mux.HandleFunc("/api/student/schedule", s.withAuth(s.handleStudentSchedule, RoleStudent))
	mux.HandleFunc("/api/student/grades", s.withAuth(s.handleStudentGrades, RoleStudent))
	mux.HandleFunc("/api/student/homework", s.withAuth(s.handleStudentHomework, RoleStudent))

	staticDir := "static"
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	return mux
}

// main запускает HTTP-сервер и логирует параметры старта.
func main() {
	store := NewStorage()
	srv := &Server{store: store}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	log.Printf("server started on %s", addr)
	log.Printf("default admin: admin@school.local / admin123")
	if err := http.ListenAndServe(addr, srv.routes()); err != nil {
		log.Fatal(err)
	}
}
