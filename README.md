# School Diary API (RU/EN)

## Русская версия

### 1. О проекте
`School Diary API` — учебный backend на Go для школьного дневника с простым web-интерфейсом.

Поддерживаются роли:
- `admin`
- `teacher`
- `student`

Ключевые возможности:
- регистрация и логин пользователей;
- просмотр профиля текущего пользователя;
- управление пользователями (админ);
- загрузка фото расписания по классу (админ);
- получение расписания (фото) для ученика его класса;
- добавление уроков/оценок/домашки (учитель);
- просмотр оценок и домашки (ученик).

### 2. Технологии
- Go (стандартная библиотека, без внешних зависимостей)
- HTTP сервер на `net/http`
- In-memory storage (данные хранятся в памяти процесса)
- Статический frontend в папке `static/`

### 3. Архитектура и структура проекта
Проект разделен по слоям и ролям:

- `server.go`  
  Точка входа, конфигурация порта, регистрация маршрутов.

- `types.go`  
  Доменные типы (`User`, `SchedulePhoto`, `Grade`, `Homework`, роли и т.д.).

- `storage.go`  
  In-memory хранилище и CRUD-операции.

- `auth.go`  
  Хеширование пароля и middleware авторизации по Bearer токену.

- `handlers_auth.go`  
  Публичные и общие auth-endpoints (`register`, `login`, `me`).

- `handlers_admin.go`  
  Админские endpoints.

- `handlers_teacher.go`  
  Endpoints для учителя.

- `handlers_student.go`  
  Endpoints для ученика.

- `utils.go`  
  Общие helper-функции (`writeJSON`, `writeError`, `normalizeClassName`).

- `static/`  
  Встроенный фронтенд (HTML/CSS/JS).

### 4. Запуск
Из корня проекта:

```powershell
go run .
```

По умолчанию сервер запускается на `http://localhost:8080`.

С другим портом:

```powershell
$env:PORT="18080"; go run .
```

### 5. Дефолтный пользователь
При старте автоматически создается админ:

- email: `admin@school.local`
- password: `admin123`

### 6. Авторизация
Защищенные endpoints ожидают заголовок:

```http
Authorization: Bearer <token>
```

Токен выдается в ответе `POST /api/login`.

### 7. Модель данных

#### 7.1 User
- `id`
- `fullName`
- `email`
- `role` (`admin|teacher|student`)
- `className` (только для ученика)

#### 7.2 SchedulePhoto
- `className`
- `contentType` (например, `image/jpeg`)
- `imageData` (`data:<mime>;base64,...`)

#### 7.3 Grade
- `id`, `studentId`, `subject`, `value (1..5)`, `comment`, `teacherId`, `date`

#### 7.4 Homework
- `id`, `className`, `subject`, `description`, `dueDate`, `teacherId`

### 8. API

Базовый URL: `http://localhost:8080`

#### 8.1 Public/Auth

1. `POST /api/register`  
Создать пользователя.

Пример body:

```json
{
  "fullName": "Иван Петров",
  "email": "ivan@student.local",
  "password": "123456",
  "role": "student",
  "className": "7A"
}
```

2. `POST /api/login`  
Логин и получение токена.

```json
{
  "email": "admin@school.local",
  "password": "admin123"
}
```

3. `GET /api/me`  
Информация о текущем пользователе (нужен токен).

#### 8.2 Admin

1. `GET /api/admin/users`  
Список пользователей.

2. `POST /api/admin/users`  
Создание пользователя (аналогично `/api/register`).

3. `DELETE /api/admin/users/{id}`  
Удаление пользователя.

4. `POST /api/admin/schedule/import`  
Загрузка фото расписания для класса (`multipart/form-data`):
- `className` (string, обязателен)
- `file` (image/*, обязателен, до ~20MB)

Ответ:

```json
{
  "status": "imported",
  "className": "7A"
}
```

5. `DELETE /api/admin/schedule`  
Очистка всех записей расписания и фото расписаний.

6. `GET /api/admin/schedule/stats`  
Статистика загруженных фото расписания по классам.

#### 8.3 Teacher

1. `POST /api/teacher/schedule`  
Добавить урок в расписание (структурные записи).

2. `POST /api/teacher/grades`  
Поставить оценку ученику.

3. `POST /api/teacher/homework`  
Добавить домашнее задание для класса.

4. `GET /api/teacher/students`  
Список учеников, отсортированный по классу.

#### 8.4 Student

1. `GET /api/student/schedule`  
Вернет фото расписания для класса ученика:
- `{}` если фото не найдено;
- объект `SchedulePhoto`, если найдено.

2. `GET /api/student/grades`  
Список оценок текущего ученика.

3. `GET /api/student/homework`  
Список домашнего задания для класса ученика.

### 9. Примеры запросов (curl)

#### 9.1 Логин
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"admin@school.local\",\"password\":\"admin123\"}"
```

#### 9.2 Импорт фото расписания для класса
```bash
curl -X POST http://localhost:8080/api/admin/schedule/import \
  -H "Authorization: Bearer <TOKEN>" \
  -F "className=7A" \
  -F "file=@schedule_7a.jpg"
```

#### 9.3 Получить расписание ученика
```bash
curl http://localhost:8080/api/student/schedule \
  -H "Authorization: Bearer <TOKEN>"
```

### 10. Нормализация className
Сервер нормализует имя класса:
- удаляет пробелы;
- приводит к верхнему регистру;
- заменяет похожие кириллические символы на латиницу (`А -> A`, `В -> B` и т.д.);
- оставляет только `[0-9A-Z]`.

Пример: `7 а` -> `7A`.

### 11. Ограничения
- Все данные хранятся в памяти и теряются при рестарте.
- Нет БД и миграций.
- Нет refresh-токенов/выхода из системы по API.
- Нет ролевой иерархии кроме явной проверки в middleware.

### 12. Идеи для развития
- Подключить PostgreSQL/SQLite.
- Добавить JWT с истечением срока и refresh.
- Добавить unit/integration тесты.
- Вынести сервисный слой между handlers и storage.
- Добавить Swagger/OpenAPI.

---

## English Version

### 1. About The Project
`School Diary API` is a Go learning backend for a school diary with a simple web UI.

Supported roles:
- `admin`
- `teacher`
- `student`

Key features:
- user registration and login;
- current user profile endpoint;
- user management (admin);
- schedule photo upload per class (admin);
- schedule photo retrieval for a student’s class;
- adding lessons/grades/homework (teacher);
- viewing grades and homework (student).

### 2. Tech Stack
- Go (standard library only, no external dependencies)
- HTTP server with `net/http`
- In-memory storage (data lives in process memory)
- Static frontend in `static/`

### 3. Architecture and Project Structure
The project is split by layers and roles:

- `server.go`  
  Entry point, port config, route registration.

- `types.go`  
  Domain types (`User`, `SchedulePhoto`, `Grade`, `Homework`, roles, etc.).

- `storage.go`  
  In-memory storage and CRUD operations.

- `auth.go`  
  Password hashing and Bearer-token auth middleware.

- `handlers_auth.go`  
  Public/shared auth endpoints (`register`, `login`, `me`).

- `handlers_admin.go`  
  Admin endpoints.

- `handlers_teacher.go`  
  Teacher endpoints.

- `handlers_student.go`  
  Student endpoints.

- `utils.go`  
  Shared helpers (`writeJSON`, `writeError`, `normalizeClassName`).

- `static/`  
  Built-in frontend (HTML/CSS/JS).

### 4. Run
From the project root:

```powershell
go run .
```

Default address: `http://localhost:8080`.

Custom port:

```powershell
$env:PORT="18080"; go run .
```

### 5. Default User
At startup, a default admin user is created:

- email: `admin@school.local`
- password: `admin123`

### 6. Authorization
Protected endpoints require:

```http
Authorization: Bearer <token>
```

The token is returned by `POST /api/login`.

### 7. Data Model

#### 7.1 User
- `id`
- `fullName`
- `email`
- `role` (`admin|teacher|student`)
- `className` (student only)

#### 7.2 SchedulePhoto
- `className`
- `contentType` (e.g. `image/jpeg`)
- `imageData` (`data:<mime>;base64,...`)

#### 7.3 Grade
- `id`, `studentId`, `subject`, `value (1..5)`, `comment`, `teacherId`, `date`

#### 7.4 Homework
- `id`, `className`, `subject`, `description`, `dueDate`, `teacherId`

### 8. API

Base URL: `http://localhost:8080`

#### 8.1 Public/Auth

1. `POST /api/register`  
Create a user.

Example body:

```json
{
  "fullName": "John Smith",
  "email": "john@student.local",
  "password": "123456",
  "role": "student",
  "className": "7A"
}
```

2. `POST /api/login`  
Login and receive token.

```json
{
  "email": "admin@school.local",
  "password": "admin123"
}
```

3. `GET /api/me`  
Current user profile (token required).

#### 8.2 Admin

1. `GET /api/admin/users`  
List users.

2. `POST /api/admin/users`  
Create a user (same as `/api/register`).

3. `DELETE /api/admin/users/{id}`  
Delete user.

4. `POST /api/admin/schedule/import`  
Upload schedule photo for a class (`multipart/form-data`):
- `className` (string, required)
- `file` (image/*, required, up to ~20MB)

Response:

```json
{
  "status": "imported",
  "className": "7A"
}
```

5. `DELETE /api/admin/schedule`  
Clear all schedule entries and schedule photos.

6. `GET /api/admin/schedule/stats`  
Stats for uploaded schedule photos grouped by class.

#### 8.3 Teacher

1. `POST /api/teacher/schedule`  
Add a lesson to structured schedule entries.

2. `POST /api/teacher/grades`  
Add grade for a student.

3. `POST /api/teacher/homework`  
Add homework for a class.

4. `GET /api/teacher/students`  
List students sorted by class.

#### 8.4 Student

1. `GET /api/student/schedule`  
Returns schedule photo for the student’s class:
- `{}` if not found;
- `SchedulePhoto` object if found.

2. `GET /api/student/grades`  
Current student grades.

3. `GET /api/student/homework`  
Homework list for the student’s class.

### 9. Request Examples (curl)

#### 9.1 Login
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"admin@school.local\",\"password\":\"admin123\"}"
```

#### 9.2 Upload schedule photo for class
```bash
curl -X POST http://localhost:8080/api/admin/schedule/import \
  -H "Authorization: Bearer <TOKEN>" \
  -F "className=7A" \
  -F "file=@schedule_7a.jpg"
```

#### 9.3 Get student schedule
```bash
curl http://localhost:8080/api/student/schedule \
  -H "Authorization: Bearer <TOKEN>"
```

### 10. `className` Normalization
Server normalizes class name by:
- removing spaces;
- uppercasing;
- replacing similar Cyrillic letters with Latin (`А -> A`, `В -> B`, etc.);
- keeping only `[0-9A-Z]`.

Example: `7 а` -> `7A`.

### 11. Limitations
- Data is in-memory and lost on restart.
- No database or migrations.
- No refresh tokens/logout API.
- No role hierarchy beyond explicit middleware checks.

### 12. Next Improvements
- Add PostgreSQL/SQLite.
- Add expiring JWT + refresh flow.
- Add unit/integration tests.
- Add service layer between handlers and storage.
- Add Swagger/OpenAPI.
