package main

// Role описывает роль пользователя в системе.
type Role string

const (
	// RoleAdmin управляет пользователями и расписанием.
	RoleAdmin   Role = "admin"
	// RoleTeacher выставляет оценки, задает ДЗ и уроки.
	RoleTeacher Role = "teacher"
	// RoleStudent просматривает свои данные.
	RoleStudent Role = "student"
)

// User — учетная запись пользователя.
type User struct {
	ID           int64  `json:"id"`
	FullName     string `json:"fullName"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         Role   `json:"role"`
	ClassName    string `json:"className,omitempty"`
}

// ScheduleEntry — структурная запись урока.
type ScheduleEntry struct {
	ID        int64  `json:"id"`
	ClassName string `json:"className"`
	Subject   string `json:"subject"`
	Weekday   string `json:"weekday"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Room      string `json:"room"`
	TeacherID int64  `json:"teacherId"`
}

// SchedulePhoto — фото расписания, привязанное к классу.
type SchedulePhoto struct {
	ClassName   string `json:"className"`
	ContentType string `json:"contentType"`
	ImageData   string `json:"imageData"`
}

// Grade — оценка ученика.
type Grade struct {
	ID        int64  `json:"id"`
	StudentID int64  `json:"studentId"`
	Subject   string `json:"subject"`
	Value     int    `json:"value"`
	Comment   string `json:"comment"`
	TeacherID int64  `json:"teacherId"`
	Date      string `json:"date"`
}

// Homework — домашнее задание для класса.
type Homework struct {
	ID          int64  `json:"id"`
	ClassName   string `json:"className"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	DueDate     string `json:"dueDate"`
	TeacherID   int64  `json:"teacherId"`
}

// Server объединяет HTTP-слой и хранилище данных.
type Server struct {
	store *Storage
}
