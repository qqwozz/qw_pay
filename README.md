<div align="center">

# 💳 QW Pay

**Платёжная система — упрощённый аналог внутренних банковских переводов**

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-24-2496ED?style=flat&logo=docker&logoColor=white)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

</div>

---

## 📋 Описание

QW Pay — микросервисная платёжная система для внутренних банковских переводов с поддержкой валютных счетов, идемпотентностью транзакций и optimistic locking.

### Текущий статус: MVP (без антифрода)

Реализовано:
- Регистрация пользователей с OTP-подтверждением
- JWT-аутентификация
- Управление счетами в RUB / USD / EUR
- Внутренние переводы с конвертацией по фиксированным курсам
- Идемпотентность (повторный запрос не создаёт дубликат)
- Optimistic locking через version (защита от гонок)
- ACID-транзакции при списании/зачислении
- Дневные лимиты переводов
- История операций с пагинацией

---

## 🏗 Архитектура

```
┌─────────────────────────────────────────────────────────┐
│                      API Gateway                        │
│                  (Gin + JWT middleware)                  │
└───────────┬─────────────┬───────────────┬───────────────┘
            │             │               │
    ┌───────▼──────┐ ┌────▼─────┐ ┌───────▼──────┐
    │    Auth      │ │ Account  │ │ Transaction  │
    │   Service    │ │ Service  │ │   Service    │
    │  (register,  │ │ (CRUD,   │ │ (transfer,   │
    │   login,     │ │  block,  │ │  idempotent, │
    │   OTP)       │ │  balance)│ │  ACID)       │
    └───────┬──────┘ └────┬─────┘ └───────┬──────┘
            │             │               │
            └─────────────┼───────────────┘
                          │
                ┌─────────▼─────────┐
                │    PostgreSQL     │
                │   (ACID + WAL)    │
                └───────────────────┘
```

---

## 🚀 Быстрый старт

### Предварительные требования

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 16+ (или через Docker)

### Запуск

```bash
# 1. Клонировать репозиторий
git clone https://github.com/your-org/qw_pay.git
cd qw_pay

# 2. Запустить PostgreSQL и aplicação
docker compose up -d

# 3. Сервер запустится на http://localhost:8080
```

### Запуск без Docker

```bash
# 1. Запустить PostgreSQL и создать БД
createdb qw_pay
psql qw_pay < migrations/001_init.sql

# 2. Настроить переменные окружения
cp .env.example .env
# отредактировать .env

# 3. Запустить сервер
go run ./cmd/server
```

---

## 📡 API Endpoints

### Публичные

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/register` | Регистрация пользователя |
| `POST` | `/api/v1/verify` | Подтверждение OTP-кодом |
| `POST` | `/api/v1/login` | Вход (получение JWT) |

### Защищённые (требуют `Authorization: Bearer <token>`)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/accounts` | Создать счёт |
| `GET` | `/api/v1/accounts` | Список счетов |
| `POST` | `/api/v1/accounts/:id/block` | Заблокировать счёт |
| `POST` | `/api/v1/transfers` | Создать перевод |
| `GET` | `/api/v1/transfers` | История переводов |

---

## 📖 Примеры использования

### Регистрация

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "phone": "+79001234567",
    "password": "securepass123"
  }'
# OTP-код будет выведен в логи сервера
```

### Подтверждение OTP

```bash
curl -X POST http://localhost:8080/api/v1/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "otp_code": "123456"
  }'
```

### Вход

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepass123"
  }'
# Ответ: { "access_token": "eyJ...", "token_type": "bearer", "user_id": "..." }
```

### Создание счёта

```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"currency": "RUB"}'
```

### Перевод

```bash
curl -X POST http://localhost:8080/api/v1/transfers \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "from_account_id": "uuid-from",
    "to_account_id": "uuid-to",
    "amount": 1000.00,
    "idempotency_key": "unique-key-123"
  }'
```

---

## 🗂 Структура проекта

```
qw_pay/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── config/
│   │   └── config.go            # Конфигурация
│   ├── database/
│   │   └── postgres.go          # Пул соединений
│   ├── models/                   # Модели данных
│   │   ├── user.go
│   │   ├── account.go
│   │   └── transaction.go
│   ├── services/                 # Бизнес-логика
│   │   ├── auth.go
│   │   ├── account.go
│   │   └── transaction.go
│   ├── handlers/                 # HTTP-хендлеры
│   │   ├── auth.go
│   │   ├── account.go
│   │   └── transaction.go
│   └── middleware/
│       └── auth.go              # JWT middleware
├── migrations/
│   └── 001_init.sql             # SQL-схема
├── docs/
│   └── TECHNICAL_SPECIFICATION.md  # ТЗ
├── docker-compose.yml
├── Dockerfile
├── .env
├── .gitignore
├── LICENSE
├── README.md
└── go.mod
```

---

## 🔧 Конфигурация

Переменные окружения (`.env`):

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/qw_pay?sslmode=disable` | Строка подключения к БД |
| `JWT_SECRET` | `change-me-in-production` | Секрет для JWT |
| `PORT` | `8080` | Порт сервера |

---

## 🛡 Безопасность

- Пароли хешируются через bcrypt
- JWT-токены с настраиваемым TTL
- Optimistic locking для защиты от race conditions
- ACID-транзакции для целостности данных
- Проверка принадлежности счетов перед операциями

---

## 📄 Лицензия

MIT License — см. файл [LICENSE](LICENSE).

---

## 🗺 План развития

- [ ] Антифрод-система (C++ engine + Python ML)
- [ ] Currency Service (внешний API курсов)
- [ ] Audit Log Service
- [ ] Kubernetes (Helm-чарты)
- [ ] Мониторинг (Prometheus + Grafana)
- [ ] Логирование (ELK)
