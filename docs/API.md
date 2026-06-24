# Документация API

## Базовый URL

```
http://localhost:8080/api/v1
```

## Аутентификация

Для защищённых эндпоинтов передавайте JWT-токен в заголовке:

```
Authorization: Bearer <token>
```

---

## Публичные эндпоинты

### POST /register

Регистрация нового пользователя.

**Request:**
```json
{
  "email": "user@example.com",
  "phone": "+79001234567",
  "password": "securepass123"
}
```

**Response (201):**
```json
{
  "message": "Registration successful. Check logs for OTP code.",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Errors:**
- `400` — Невалидные данные или email уже занят

---

### POST /verify

Подтверждение аккаунта OTP-кодом.

**Request:**
```json
{
  "email": "user@example.com",
  "otp_code": "123456"
}
```

**Response (200):**
```json
{
  "message": "Account verified"
}
```

**Errors:**
- `400` — Неверный или просроченный OTP

---

### POST /login

Вход в систему. Возвращает JWT-токен.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepass123"
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4...",
  "token_type": "bearer",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Errors:**
- `401` — Неверные учётные данные
- `403` — Аккаунт не подтверждён

---

## Защищённые эндпоинты

### POST /accounts

Создание нового счёта. При создании начисляется бонус 100 единиц валюты.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "currency": "USD"
}
```

**Response (201):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "currency": "USD",
  "balance": "100.00",
  "version": 1,
  "status": "ACTIVE",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

### GET /accounts

Получение списка счетов пользователя.

**Headers:** `Authorization: Bearer <token>`

**Response (200):**
```json
[
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "currency": "USD",
    "balance": "100.00",
    "version": 1,
    "status": "ACTIVE",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

---

### POST /accounts/:id/block

Блокировка счёта.

**Headers:** `Authorization: Bearer <token>`

**Response (200):**
```json
{
  "message": "Account blocked"
}
```

**Errors:**
- `403` — Счёт не принадлежит пользователю
- `404` — Счёт не найден

---

### POST /transactions

Создание перевода между счетами.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "from_account_id": "660e8400-e29b-41d4-a716-446655440001",
  "to_account_id": "660e8400-e29b-41d4-a716-446655440002",
  "amount": 50.00,
  "idempotency_key": "unique-key-123"
}
```

**Response (201):**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440003",
  "idempotency_key": "unique-key-123",
  "from_account_id": "660e8400-e29b-41d4-a716-446655440001",
  "to_account_id": "660e8400-e29b-41d4-a716-446655440002",
  "amount": "50.00",
  "currency": "USD",
  "status": "EXECUTED",
  "created_at": "2024-01-15T10:35:00Z"
}
```

**Errors:**
- `400` — Недостаточно средств, превышен лимит, невалидные данные
- `403` — Счёт не принадлежит пользователю / заблокирован антифродом

---

### GET /transactions

История переводов пользователя.

**Headers:** `Authorization: Bearer <token>`

**Response (200):**
```json
{
  "transactions": [...],
  "total": 15,
  "page": 1,
  "page_size": 20
}
```

---

### POST /antifraud/block-user

Блокировка пользователя в антифрод-системе.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response (200):**
```json
{
  "message": "User blocked in anti-fraud",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### POST /antifraud/block-account

Блокировка счёта в антифрод-системе.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "account_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Response (200):**
```json
{
  "message": "Account blocked in anti-fraud",
  "account_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

---

## Курсы валют

| From | To | Rate |
|------|----|------|
| RUB | USD | 0.011 |
| USD | RUB | 90.91 |
| RUB | EUR | 0.010 |
| EUR | RUB | 100.0 |
| USD | EUR | 0.92 |
| EUR | USD | 1.09 |

---

## Ошибки

Все ошибки возвращаются в формате:

```json
{
  "error": "Описание ошибки"
}
```

| Код | Описание |
|-----|----------|
| `400` | Невалидные данные запроса |
| `401` | Не авторизован |
| `403` | Доступ запрещён |
| `404` | Ресурс не найден |
| `500` | Внутренняя ошибка сервера |
