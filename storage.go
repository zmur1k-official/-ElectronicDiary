package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"sync"
)

// Storage — потокобезопасное in-memory хранилище приложения.
type Storage struct {
	mu       sync.RWMutex
	users    map[int64]User
	emailIdx map[string]int64
	tokens   map[string]int64

	schedule map[int64]ScheduleEntry
	photos   map[string]SchedulePhoto
	grades   map[int64]Grade
	homework map[int64]Homework

	nextUserID     int64
	nextScheduleID int64
	nextGradeID    int64
	nextHomeworkID int64
}

// NewStorage создает и инициализирует хранилище начальными структурами.
func NewStorage() *Storage {
	s := &Storage{
		users:    make(map[int64]User),
		emailIdx: make(map[string]int64),
		tokens:   make(map[string]int64),
		schedule: make(map[int64]ScheduleEntry),
		photos:   make(map[string]SchedulePhoto),
		grades:   make(map[int64]Grade),
		homework: make(map[int64]Homework),

		nextUserID:     1,
		nextScheduleID: 1,
		nextGradeID:    1,
		nextHomeworkID: 1,
	}
	s.seed()
	return s
}

// seed добавляет стартовые данные (дефолтного администратора).
func (s *Storage) seed() {
	admin := User{
		ID:           s.nextUserID,
		FullName:     "System Admin",
		Email:        "admin@school.local",
		PasswordHash: hashPassword("admin123"),
		Role:         RoleAdmin,
	}
	s.users[admin.ID] = admin
	s.emailIdx[strings.ToLower(admin.Email)] = admin.ID
	s.nextUserID++
}

// createUser создает нового пользователя и индексирует его email.
func (s *Storage) createUser(u User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	emailKey := strings.ToLower(strings.TrimSpace(u.Email))
	if emailKey == "" {
		return User{}, errors.New("email is required")
	}
	if _, exists := s.emailIdx[emailKey]; exists {
		return User{}, errors.New("email already exists")
	}

	u.ID = s.nextUserID
	s.nextUserID++
	s.users[u.ID] = u
	s.emailIdx[emailKey] = u.ID
	return u, nil
}

// findUserByEmail ищет пользователя по email.
func (s *Storage) findUserByEmail(email string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.emailIdx[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return User{}, false
	}
	u, ok := s.users[id]
	return u, ok
}

// getUser возвращает пользователя по ID.
func (s *Storage) getUser(id int64) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

// listUsers возвращает список всех пользователей.
func (s *Storage) listUsers() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]User, 0, len(s.users))
	for _, u := range s.users {
		res = append(res, u)
	}
	return res
}

// listStudentsSortedByClass возвращает учеников, отсортированных по классу и ФИО.
func (s *Storage) listStudentsSortedByClass() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]User, 0)
	for _, u := range s.users {
		if u.Role == RoleStudent {
			res = append(res, u)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		leftClass := strings.ToLower(normalizeClassName(res[i].ClassName))
		rightClass := strings.ToLower(normalizeClassName(res[j].ClassName))
		if leftClass == rightClass {
			return strings.ToLower(res[i].FullName) < strings.ToLower(res[j].FullName)
		}
		return leftClass < rightClass
	})
	return res
}

// deleteUser удаляет пользователя и все его активные токены.
func (s *Storage) deleteUser(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return false
	}
	delete(s.users, id)
	delete(s.emailIdx, strings.ToLower(u.Email))
	for token, userID := range s.tokens {
		if userID == id {
			delete(s.tokens, token)
		}
	}
	return true
}

// createToken создает и сохраняет токен сессии.
func (s *Storage) createToken(userID int64) (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	s.mu.Lock()
	s.tokens[token] = userID
	s.mu.Unlock()
	return token, nil
}

// userByToken возвращает пользователя по токену.
func (s *Storage) userByToken(token string) (User, bool) {
	s.mu.RLock()
	userID, ok := s.tokens[token]
	if !ok {
		s.mu.RUnlock()
		return User{}, false
	}
	u, ok := s.users[userID]
	s.mu.RUnlock()
	return u, ok
}

// addSchedule добавляет запись урока в расписание.
func (s *Storage) addSchedule(entry ScheduleEntry) ScheduleEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry.ClassName = normalizeClassName(entry.ClassName)
	entry.ID = s.nextScheduleID
	s.nextScheduleID++
	s.schedule[entry.ID] = entry
	return entry
}

// replaceSchedule полностью заменяет структурное расписание.
func (s *Storage) replaceSchedule(entries []ScheduleEntry) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedule = make(map[int64]ScheduleEntry)
	s.nextScheduleID = 1
	for i := range entries {
		entries[i].ClassName = normalizeClassName(entries[i].ClassName)
		entries[i].ID = s.nextScheduleID
		s.schedule[entries[i].ID] = entries[i]
		s.nextScheduleID++
	}
	return len(entries)
}

// clearSchedule очищает структурное расписание и фото расписаний.
func (s *Storage) clearSchedule() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedule = make(map[int64]ScheduleEntry)
	s.photos = make(map[string]SchedulePhoto)
	s.nextScheduleID = 1
}

// setSchedulePhoto сохраняет фото расписания для класса.
func (s *Storage) setSchedulePhoto(className, contentType string, raw []byte) SchedulePhoto {
	s.mu.Lock()
	defer s.mu.Unlock()

	className = normalizeClassName(className)
	if contentType == "" {
		contentType = "image/jpeg"
	}
	photo := SchedulePhoto{
		ClassName:   className,
		ContentType: contentType,
		ImageData:   "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(raw),
	}
	s.photos[className] = photo
	return photo
}

// getSchedulePhotoByClass возвращает фото расписания конкретного класса.
func (s *Storage) getSchedulePhotoByClass(className string) (SchedulePhoto, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	photo, ok := s.photos[normalizeClassName(className)]
	return photo, ok
}

// schedulePhotoStats считает количество фото расписаний по классам.
func (s *Storage) schedulePhotoStats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	classes := make(map[string]int, len(s.photos))
	for className := range s.photos {
		classes[className] = 1
	}
	return classes
}

// listScheduleByClass возвращает структурные записи расписания класса.
func (s *Storage) listScheduleByClass(className string) []ScheduleEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	className = normalizeClassName(className)
	res := []ScheduleEntry{}
	for _, entry := range s.schedule {
		if normalizeClassName(entry.ClassName) == className {
			res = append(res, entry)
		}
	}
	return res
}

// listAllSchedule возвращает все структурные записи расписания.
func (s *Storage) listAllSchedule() []ScheduleEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listAllScheduleLocked()
}

// listAllScheduleLocked возвращает расписание без повторного захвата mutex.
func (s *Storage) listAllScheduleLocked() []ScheduleEntry {
	res := make([]ScheduleEntry, 0, len(s.schedule))
	for _, entry := range s.schedule {
		res = append(res, entry)
	}
	return res
}

// addGrade добавляет оценку.
func (s *Storage) addGrade(g Grade) Grade {
	s.mu.Lock()
	defer s.mu.Unlock()
	g.ID = s.nextGradeID
	s.nextGradeID++
	s.grades[g.ID] = g
	return g
}

// listGradesByStudent возвращает оценки конкретного ученика.
func (s *Storage) listGradesByStudent(studentID int64) []Grade {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := []Grade{}
	for _, g := range s.grades {
		if g.StudentID == studentID {
			res = append(res, g)
		}
	}
	return res
}

// addHomework добавляет домашнее задание для класса.
func (s *Storage) addHomework(hw Homework) Homework {
	s.mu.Lock()
	defer s.mu.Unlock()
	hw.ClassName = normalizeClassName(hw.ClassName)
	hw.ID = s.nextHomeworkID
	s.nextHomeworkID++
	s.homework[hw.ID] = hw
	return hw
}

// listHomeworkByClass возвращает домашние задания указанного класса.
func (s *Storage) listHomeworkByClass(className string) []Homework {
	s.mu.RLock()
	defer s.mu.RUnlock()
	className = normalizeClassName(className)
	res := []Homework{}
	for _, hw := range s.homework {
		if normalizeClassName(hw.ClassName) == className {
			res = append(res, hw)
		}
	}
	return res
}
