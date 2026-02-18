# School Diary API (RU/EN)

## Русская версия

### 1. О проекте
`School Diary API` — учебный backend на Go для электронного дневника.

Роли:
- `admin`
- `teacher`
- `student`

Основные возможности:
- регистрация и вход пользователей;
- авторизация по Bearer-токену;
- управление пользователями (админ);
- загрузка фото расписания по классу (админ);
- просмотр фото расписания своего класса (ученик);
- постановка оценок (учитель);
- табличный просмотр оценок для ученика;
- табличный журнал учителя по предмету.

### 2. Технологии
- Go (стандартная библиотека);
- `net/http`;
- in-memory хранилище (данные в памяти процесса);
- статический frontend (`static/`).

### 3. Архитектура
Структура проекта сохранена модульной:

- `server.go` — запуск сервера и маршруты;
- `types.go` — модели данных;
- `storage.go` — потокобезопасное in-memory хранилище;
- `auth.go` — auth middleware и хеширование пароля;
- `handlers_auth.go` — публичные/auth endpoints;
- `handlers_admin.go` — endpoints администратора;
- `handlers_teacher.go` — endpoints учителя;
- `handlers_student.go` — endpoints ученика;
- `utils.go` — общие helper-функции;
- `static/` — клиентская часть.

### 4. Запуск
Из корня проекта:

```powershell
go run .
```

По умолчанию сервер: `http://localhost:8080`.

Другой порт:

```powershell
$env:PORT="18080"; go run .
```

### 5. Дефолтный админ
- email: `admin@school.local`
- password: `admin123`

### 6. Авторизация
Для защищенных endpoints:

```http
Authorization: Bearer <token>
```

Токен выдается через `POST /api/login`.

### 7. Ключевые модели

#### User
- `id`, `fullName`, `email`, `role`, `className`

#### Grade
- `id`, `studentId`, `subject`, `value`, `comment`, `teacherId`, `date`

#### SchedulePhoto
- `className`, `contentType`, `imageData`

### 8. API

Базовый URL: `http://localhost:8080`

#### 8.1 Auth/Public
1. `POST /api/register`
2. `POST /api/login`
3. `GET /api/me`

#### 8.2 Admin
1. `GET /api/admin/users`
2. `POST /api/admin/users`
3. `DELETE /api/admin/users/{id}`
4. `POST /api/admin/schedule/import` (`multipart/form-data`: `className`, `file:image/*`)
5. `DELETE /api/admin/schedule`
6. `GET /api/admin/schedule/stats`

#### 8.3 Teacher
1. `POST /api/teacher/schedule`
2. `GET /api/teacher/students`
3. `GET /api/teacher/subject` — текущий закрепленный предмет учителя
4. `POST /api/teacher/subject` — закрепить предмет:
```json
{ "subject": "Математика" }
```
5. `GET /api/teacher/grades/journal?from=YYYY-MM-DD&to=YYYY-MM-DD` — оценки учителя по его предмету за период
6. `GET /api/teacher/grades?studentId=<id>` — оценки конкретного ученика
7. `POST /api/teacher/grades` — поставить оценку:
```json
{
  "studentId": 12,
  "value": 5,
  "comment": "Отлично",
  "date": "2026-02-18"
}
```
Важно: `subject` в этом запросе не передается, берется из закрепленного предмета учителя.
8. `POST /api/teacher/homework`

#### 8.4 Student
1. `GET /api/student/schedule`
2. `GET /api/student/grades`
3. `GET /api/student/homework`

### 9. Таблицы оценок в UI

#### Для ученика
- таблица: строки — предметы;
- столбцы — даты;
- в ячейке — оценка;
- при наведении на оценку отображается комментарий.

#### Для учителя
- сначала учитель задает свой предмет;
- таблица: первый столбец — ученики (уже отсортированы бэкендом по классам);
- остальные столбцы — даты от `-7` до `+7` дней относительно текущей даты;
- клик по ячейке открывает ввод оценки и комментария;
- оценка сохраняется на выбранную дату.

### 10. Ограничения
- данные хранятся в памяти и пропадают при рестарте;
- нет БД и миграций;
- нет refresh-токенов.

---

## English Version

### 1. About
`School Diary API` is a Go learning backend for an electronic school diary.

Roles:
- `admin`
- `teacher`
- `student`

Features:
- registration/login;
- Bearer-token auth;
- user management (admin);
- class schedule photo upload (admin);
- student schedule photo view;
- teacher grading;
- student grade table view;
- teacher subject-based grade journal.

### 2. Stack
- Go (standard library);
- `net/http`;
- in-memory storage;
- static frontend in `static/`.

### 3. Architecture
- `server.go` — bootstrap + routes
- `types.go` — domain models
- `storage.go` — thread-safe storage
- `auth.go` — auth middleware/password hashing
- `handlers_auth.go` — auth/public endpoints
- `handlers_admin.go` — admin endpoints
- `handlers_teacher.go` — teacher endpoints
- `handlers_student.go` — student endpoints
- `utils.go` — helpers
- `static/` — frontend

### 4. Run
```powershell
go run .
```

Default: `http://localhost:8080`

Custom port:
```powershell
$env:PORT="18080"; go run .
```

### 5. Default admin
- email: `admin@school.local`
- password: `admin123`

### 6. Auth
Protected endpoints require:
```http
Authorization: Bearer <token>
```

### 7. Main models
- `User`
- `Grade`
- `SchedulePhoto`

### 8. API
Base URL: `http://localhost:8080`

#### 8.1 Auth/Public
1. `POST /api/register`
2. `POST /api/login`
3. `GET /api/me`

#### 8.2 Admin
1. `GET /api/admin/users`
2. `POST /api/admin/users`
3. `DELETE /api/admin/users/{id}`
4. `POST /api/admin/schedule/import` (`className` + `file:image/*`)
5. `DELETE /api/admin/schedule`
6. `GET /api/admin/schedule/stats`

#### 8.3 Teacher
1. `POST /api/teacher/schedule`
2. `GET /api/teacher/students`
3. `GET /api/teacher/subject`
4. `POST /api/teacher/subject`
5. `GET /api/teacher/grades/journal?from=YYYY-MM-DD&to=YYYY-MM-DD`
6. `GET /api/teacher/grades?studentId=<id>`
7. `POST /api/teacher/grades` (subject is taken from teacher profile)
8. `POST /api/teacher/homework`

#### 8.4 Student
1. `GET /api/student/schedule`
2. `GET /api/student/grades`
3. `GET /api/student/homework`

### 9. Grade tables in UI

#### Student
- rows: subjects
- columns: dates
- cell: grade
- hover on grade: comment tooltip

#### Teacher
- teacher sets a fixed subject first
- first column: students (already sorted by class from backend)
- date columns: from `-7` to `+7` days around today
- click a cell to enter grade + comment for that date

### 10. Limitations
- in-memory data (reset on restart)
- no DB/migrations
- no refresh tokens
