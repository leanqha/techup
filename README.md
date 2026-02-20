

# API Documentation

Документация по публичным и административным API сервера TechUp.

---

# Account API

## POST /api/v1/account/register
Регистрирует нового пользователя.
### Request
```json
{
  "email": "string",
  "password": "string",
  "first_name": "string",
  "last_name": "string"
}
```
### Response
```json
{
  "id": 1,
  "email": "user@example.com"
}
```

---

## POST /api/v1/account/login
Авторизация пользователя. Устанавливает access_token и refresh_token в HttpOnly куки.
### Request
```json
{
  "email": "string",
  "password": "string"
}
```
### Response
```json
{
  "message": "login successful"
}
```

---

## GET /api/v1/account/secure/profile
Получение профиля текущего пользователя.
### Response
```json
{
  "id": 1,
  "uid": "uuid",
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "role": "student"
}
```

---

# Schedule API

## GET /api/v1/schedule/faculties
Список факультетов.

## POST /api/v1/admin/faculty
Добавление факультета (только админ).
### Request
```json
{"name": "ФКТИ"}
```

## PUT /api/v1/admin/faculty/:id
Обновление факультета.

## DELETE /api/v1/admin/faculty/:id
Удаление факультета.

---

## GET /api/v1/schedule/groups
Список групп.

## POST /api/v1/admin/group
Добавление группы.

---

## GET /api/v1/schedule/lessons
Список занятий.

## POST /api/v1/admin/lesson
Добавление занятия.

---

# Map API

## GET /api/v1/map/search
Поиск комнат по building_id и/или floor.

Пример:
```
/api/v1/map/search?building_id=1&floor=2
```

## GET /api/v1/map/path/:start/:end
Поиск кратчайшего пути между двумя комнатами.

### Response
```json
{
  "path": ["101", "102", "103"],
  "distance": 24.5
}
```

---

# Admin API

Требуют роль `admin`.

- POST /api/v1/admin/set-role
- POST /api/v1/admin/faculty
- POST /api/v1/admin/group
- POST /api/v1/admin/lesson
- PUT /api/v1/admin/*
- DELETE /api/v1/admin/*

---

# Ошибки
Единый формат ошибок:
```json
{
  "error": "string"
}
```