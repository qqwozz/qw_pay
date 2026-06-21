<div align="center">

# QW Pay

**Микросервисная платёжная система для внутренних банковских переводов**

[![CI/CD](https://github.com/qqwozz/qw_pay/actions/workflows/ci.yml/badge.svg)](https://github.com/qqwozz/qw_pay/actions)
[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis&logoColor=white)](https://redis.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

```
POST /register  →  OTP  →  POST /login  →  JWT + Refresh Token
POST /accounts  →  +100 bonus  →  POST /transactions  →  ACID transfer
```

</div>

---

## Возможности

| Компонент | Реализация |
|-----------|-----------|
| **Аутентификация** | JWT access + refresh tokens, OTP-подтверждение email |
| **Счета** | Мультивалютные (RUB/USD/EUR), optimistic locking, приветственный бонус |
| **Переводы** | ACID-транзакции, идемпотентность, конвертация по фиксированным курсам |
| **Антифрод** | Движок на C++ (velocity, блэклисты) + Python (scoring, паттерны) через Redis |
| **Деньги** | `shopspring/decimal` — точные вычисления без погрешности float64 |
| **API** | REST JSON, cursor-based пагинация, единый формат ответов |
| **CI/CD** | GitHub Actions → test → lint → build → auto-release |

---

## Быстрый старт

### Требования

- Go 1.23+
- Docker & Docker Compose

### Запуск

```bash
git clone https://github.com/qqwozz/qw_pay.git
cd qw_pay

make db              # PostgreSQL + Redis
make run             # Сервер на :8080
```

### Демо

Откройте http://localhost:8080/demo — интерактивная страница со всеми эндпоинтами.

---

## Архитектура

```
                    ┌──────────────────────┐
                    │      API Gateway      │
                    │   Gin + JWT + CORS    │
                    └──────────┬───────────┘
                               │
           ┌───────────────────┼───────────────────┐
           │                   │                   │
    ┌──────▼──────┐     ┌──────▼──────┐     ┌──────▼──────┐
    │  Auth       │     │  Account    │     │ Transaction │
    │  Service    │     │  Service    │     │  Service    │
    │             │     │             │     │             │
    │ • Register  │     │ • Create    │     │ • Transfer  │
    │ • Login     │     │ • List      │     │ • Idempotent│
    │ • OTP       │     │ • Block     │     │ • ACID      │
    │ • Refresh   │     │ • Bonus     │     │ • Cross-cur │
    └──────┬──────┘     └──────┬──────┘     └──────┬──────┘
           │                   │                   │
           └───────────────────┼───────────────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
  ┌──────▼──────┐      ┌──────▼──────┐      ┌───────▼──────┐
  │ PostgreSQL  │      │    Redis    │      │  Anti-Fraud  │
  │   (ACID)    │      │  (queues)   │      │  C++ + Python│
  └─────────────┘      └─────────────┘      └──────────────┘
```

---

## API Endpoints

### Публичные

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/register` | Регистрация (email + OTP) |
| `POST` | `/api/v1/verify` | Подтверждение OTP |
| `POST` | `/api/v1/login` | Вход → access + refresh tokens |
| `POST` | `/api/v1/refresh` | Обновление токенов |

### Защищённые (Bearer token)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/logout` | Выход (revokes refresh tokens) |
| `POST` | `/api/v1/accounts` | Создать счёт (+100 бонус) |
| `GET` | `/api/v1/accounts` | Список счетов |
| `POST` | `/api/v1/accounts/:id/block` | Заблокировать счёт |
| `POST` | `/api/v1/transactions` | Создать перевод |
| `GET` | `/api/v1/transactions?page=1&page_size=20` | История переводов |

---

## Структура проекта

```
qw_pay/
├── cmd/server/main.go           # Точка входа, DI, graceful shutdown
├── internal/
│   ├── account/                 # CRUD счетов, баланс, optimistic locking
│   ├── antifraud/               # Redis-клиент + anti-fraud handler
│   ├── auth/                    # JWT + OTP + refresh tokens
│   ├── config/                  # Конфигурация из env (decimal)
│   ├── contextkeys/             # Безопасное извлечение userID из context
│   ├── database/                # PostgreSQL пул соединений
│   ├── errors/                  # Типизированные ошибки (NotFound, BadRequest...)
│   ├── logger/                  # Structured logging (slog)
│   ├── middleware/              # AuthRequired, RequestID, CORS
│   ├── model/                   # Модели данных (decimal.Decimal)
│   ├── ratelimit/               # Token bucket rate limiter
│   ├── response/                # Единый JSON-формат ответов
│   └── transaction/             # Переводы с конвертацией, ACID
├── antifraud/
│   ├── cpp/                     # C++ движок (velocity, блэклисты)
│   └── orchestrator.py          # Python оркестратор (scoring)
├── migrations/                  # SQL миграции
├── web/                         # Demo frontend
├── Makefile
└── docker-compose.yml
```

---

## Конфигурация

| Переменная | По умолчанию | Описание |
|------------|-------------|----------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/qw_pay` | PostgreSQL |
| `JWT_SECRET` | — | Секрет для JWT (обязательно!) |
| `PORT` | `8080` | Порт сервера |
| `REDIS_ADDR` | `127.0.0.1:6379` | Redis |
| `MAX_TRANSFER_AMOUNT` | `10000000` | Лимит перевода |
| `DAILY_LIMIT` | `50000000` | Дневной лимит |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:8080` | Разрешённые origins |

---

## Команды

```bash
make help            # Список всех команд
make db              # PostgreSQL + Redis (Docker)
make run             # Собрать и запустить
make test            # go test -v ./...
make lint            # golangci-lint
make antifraud       # C++ + Python движки
make demo            # Открыть демо в браузере
make clean           # Удалить бинарники
```

---

## Антифрод-система

### C++ (быстрые проверки, <1ms)

- Velocity: макс 5 транзакций/мин, 20/час
- Лимит суммы: 500K за перевод, 2M в день
- Блэклисты пользователей и счетов
- Детекция круговых переводов и round amounts

### Python (глубокий scoring)

- Возраст аккаунта (< 24ч = +15 risk)
- Необычные часы (2-5 утра = +20 risk)
- Отклонение суммы от среднего
- Много уникальных получателей
- risk >= 60 → блокировка

---

## Лицензия

MIT License — [LICENSE](LICENSE)
