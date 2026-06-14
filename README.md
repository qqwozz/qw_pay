<div align="center">

# QW Pay

### Платёжная система для внутренних банковских переводов

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![C++](https://img.shields.io/badge/C++-17-00599C?style=for-the-badge&logo=cplusplus&logoColor=white)](https://isocpp.org/)
[![Python](https://img.shields.io/badge/Python-3.12-3776AB?style=for-the-badge&logo=python&logoColor=white)](https://www.python.org/)
[![Docker](https://img.shields.io/badge/Docker-24-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](LICENSE)

</div>

---

## О проекте

QW Pay — микросервисная платёжная система, демонстрирующая архитектурные паттерны enterprise-уровня: ACID-транзакции с optimistic locking, идемпотентность, антифрод-система с real-time C++ движком и ML-скорингом на Python, аудит-логирование и RBAC.

### Ключевые возможности

- **Регистрация и аутентификация** — OTP-подтверждение, JWT с ролями (USER/ADMIN)
- **Мультивалютные счета** — RUB/USD/EUR, конвертация по курсам из БД
- **ACID-переводы** — optimistic locking (retry до 3 попыток), идемпотентность
- **Антифрод** — C++ engine (<5ms) + Python ML-scoring через Redis очереди
- **Аудит-лог** — неизменяемый журнал всех операций
- **Админ-панель** — просмотр логов, блокировка счетов

---

## Архитектура

### Общая схема

```mermaid
graph TB
    Client["🖥️ Client<br/>(Browser / API)"]

    subgraph Go["Go Backend"]
        GW["API Gateway<br/>Gin + JWT"]
        Auth["Auth Service<br/>Register, OTP, JWT"]
        Acc["Account Service<br/>CRUD, Block, Bonus"]
        Tx["Transaction Service<br/>Transfer, Idempotent, ACID"]
        Cur["Currency Service<br/>Rates, Convert"]
        Audit["Audit Service<br/>Logging, Admin API"]
    end

    subgraph AntiFraud["Anti-Fraud System"]
        AFClient["Go Client<br/>(Redis Queue)"]
        CPP["C++ Engine<br/>Velocity, Limits, Blacklists<br/>< 5ms"]
        PY["Python Engine<br/>ML Scoring, Patterns<br/>< 10ms"]
    end

    subgraph Data["Data Layer"]
        PG[("PostgreSQL 16<br/>ACID, WAL")]
        RD[("Redis 7<br/>Queue, Cache")]
    end

    Client --> GW
    GW --> Auth
    GW --> Acc
    GW --> Tx
    GW --> Cur
    GW --> Audit

    Tx --> AFClient
    AFClient -->|"antifraud:queue"| RD
    AFClient -->|"antifraud:queue:python"| RD
    RD --> CPP
    RD --> PY
    CPP -->|"verdict"| RD
    PY -->|"verdict"| RD
    AFClient -->|"poll verdict"| RD

    Auth --> PG
    Acc --> PG
    Tx --> PG
    Cur --> PG
    Audit --> PG
```

### Поток перевода

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler
    participant AF as Anti-Fraud
    participant S as Service
    participant DB as PostgreSQL

    C->>H: POST /api/v1/transactions
    H->>H: Verify JWT + Ownership
    H->>AF: Check(from, to, amount)
    AF->>AF: C++ velocity + limits
    AF->>AF: Python ML scoring
    AF-->>H: Verdict (APPROVE/BLOCK)

    alt BLOCKED
        H-->>C: 403 Forbidden
    else APPROVED
        H->>S: Create Transfer
        S->>S: Validate + Check Limits
        S->>DB: BEGIN
        S->>DB: INSERT transaction (idempotency_key)
        S->>DB: UPDATE accounts SET balance = balance - amount WHERE version = X
        S->>DB: UPDATE accounts SET balance = balance + amount WHERE version = Y
        S->>DB: COMMIT
        S->>Audit: Log TRANSFER_COMPLETED
        S-->>C: 200 OK
    end
```

---

## Технологии

| Компонент | Технология | Зачем |
|-----------|-----------|-------|
| API Gateway | Go + Gin | Высокая производительность, стандартная экосистема |
| Аутентификация | JWT (HS256) + OTP | Stateless-авторизация, OTP для верификации |
| БД | PostgreSQL 16 | ACID, WAL, NUMERIC для финансовых данных |
| Кэш/Очередь | Redis 7 | Антифрод очереди, blacklist-кэш, rate limiting |
| Антифрод (fast) | C++ 17 + hiredis | <5ms latency, velocity checks, blacklists |
| Антифрод (ML) | Python 3.12 | RandomForest scoring, паттерн-анализ |
| Контейнеры | Docker + Compose | Локальная разработка, деплой |
| Миграции | SQL файлы | Версионирование схемы БД |

---

## Структура проекта

```
qw_pay/
├── cmd/server/                  # Точка входа
│   └── main.go                  # DI, routing, startup
├── internal/
│   ├── auth/                    # Аутентификация
│   │   ├── handler.go           # Register, VerifyOTP, Login
│   │   ├── service.go           # OTP, JWT, bcrypt
│   │   └── repository.go        # Users CRUD
│   ├── account/                 # Счета
│   │   ├── handler.go           # Create, List, Block, AdminBlock
│   │   ├── service.go           # Business logic, WelcomeBonus
│   │   └── repository.go        # Accounts CRUD + optimistic lock
│   ├── transaction/             # Переводы
│   │   ├── handler.go           # Create, List (pagination)
│   │   ├── service.go           # ACID, idempotency, retry loop
│   │   └── repository.go        # Debit/Credit with version check
│   ├── currency/                # Курсы валют
│   │   ├── handler.go           # Rates CRUD, Convert
│   │   ├── service.go           # Conversion logic
│   │   └── repository.go        # Exchange rates DB
│   ├── audit/                   # Аудит-лог
│   │   ├── handler.go           # Admin: list logs
│   │   ├── service.go           # Log events, fraud verdicts
│   │   └── repository.go        # Audit log + fraud_checks DB
│   ├── antifraud/               # Антифрод клиент
│   │   └── client.go            # Redis queue, verdict polling
│   ├── config/                  # Конфигурация
│   │   └── config.go            # .env loading
│   ├── database/                # Пул соединений
│   │   └── postgres.go          # pgxpool
│   ├── middleware/               # HTTP middleware
│   │   └── auth.go              # JWT, AdminRequired
│   └── model/                   # Данные
│       ├── user.go              # User + roles
│       ├── account.go           # Account + status/currency enums
│       ├── transaction.go       # Transaction + status
│       └── audit.go             # AuditLog, FraudCheck
├── antifraud/
│   ├── cpp/
│   │   ├── fraud_engine.cpp     # C++ real-time engine
│   │   ├── Makefile
│   │   └── Dockerfile
│   ├── python/
│   │   └── service.py           # Python ML scoring
│   └── orchestrator.py          # Запуск обоих движков
├── migrations/
│   ├── 001_init.sql             # users, accounts, transactions
│   ├── 002_exchange_rates.sql   # exchange_rates + seed data
│   └── 003_audit_fraud_tables.sql  # audit_log, fraud_checks
├── web/
│   └── index.html               # SPA demo (dark theme)
├── deploy/
│   ├── docker-compose.yml       # Full stack orchestration
│   └── Dockerfile               # Go app multi-stage build
├── docs/
│   ├── API.md                   # API reference
│   ├── ARCHITECTURE.md          # Architecture deep-dive
│   ├── DEVELOPMENT.md           # Developer guide
│   └── TECHNICAL_SPECIFICATION.md  # Requirements
├── docker-compose.yml           # Dev: PostgreSQL + Redis + App
├── Dockerfile                   # Multi-stage Go build
├── Makefile                     # Dev commands
└── .env.example                 # Configuration template
```

---

## Быстрый старт

### Требования

- Go 1.22+
- Docker & Docker Compose
- g++ и libhiredis-dev (для C++ антифрода)
- Python 3.12+ и `pip install redis` (для Python антифрода)

### Запуск

```bash
# 1. Клонировать и настроить
git clone <repo-url> && cd qw_pay
cp .env.example .env

# 2. Запустить PostgreSQL + Redis
make db

# 3. Запустить сервер
make run

# 4. (Опционально) Антифрод движки
make antifraud

# 5. Открыть демо
open http://localhost:8080/demo
```

### Docker (полный стек)

```bash
cd deploy
docker-compose up -d
```

---

## API Endpoints

### Публичные

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/register` | Регистрация (email + OTP) |
| `POST` | `/api/v1/verify` | Подтверждение OTP |
| `POST` | `/api/v1/login` | Вход → JWT |
| `GET` | `/api/v1/currencies` | Список валют |
| `GET` | `/api/v1/currencies/rates` | Все курсы |
| `GET` | `/api/v1/currencies/rate?from=RUB&to=USD` | Курс пары |
| `POST` | `/api/v1/currencies/convert` | Конвертация |

### Защищённые (Bearer JWT)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/accounts` | Создать счёт (+100 бонус) |
| `GET` | `/api/v1/accounts` | Мои счета |
| `POST` | `/api/v1/accounts/:id/block` | Заблокировать свой счёт |
| `POST` | `/api/v1/transactions` | Создать перевод |
| `GET` | `/api/v1/transactions?page=1&page_size=20` | История переводов |
| `POST` | `/api/v1/antifraud/block-user` | Блокировка юзера (AF) |
| `POST` | `/api/v1/antifraud/block-account` | Блокировка счёта (AF) |

### Админ (Bearer JWT + role=ADMIN)

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/api/v1/admin/audit-logs?action=&user_id=&page=1` | Аудит-лог |
| `POST` | `/api/v1/admin/accounts/block` | Блокировка счёта (админ) |

---

## Антифрод-система

```mermaid
graph LR
    TX["Transfer<br/>Request"] --> Queue["Redis Queue"]
    Queue --> CPP["C++ Engine<br/>< 5ms"]
    Queue --> PY["Python Engine<br/>< 10ms"]
    CPP --> Verdict["Verdict<br/>APPROVE/BLOCK"]
    PY --> Verdict
    Verdict --> Go["Go Backend<br/>5s timeout"]
```

### C++ Engine — быстрые проверки

| Проверка | Лимит | Risk |
|----------|-------|------|
| Velocity (мин) | 5 транзакций | +85 |
| Velocity (час) | 20 транзакций | +80 |
| Сумма за перевод | 500,000 | +90 |
| Сумма за день | 2,000,000 | +75 |
| Блэклист | — | +100 |
| Круговые переводы | 10+ уникальных | +70 |

### Python Engine — ML скоринг

| Проверка | Risk |
|----------|------|
| Аккаунт < 1 часа | +40 |
| Ночные переводы (2-5 утра) | +20 |
| Сумма > 5x от средней | +25 |
| > 15 уникальных получателей | +30 |
| Интервал < 10 сек | +35 |
| Крупные круглые суммы | +20 |

**Вердикт:** risk ≥ 60 → BLOCK

---

## Паттерны

### Optimistic Locking

```sql
UPDATE accounts
SET balance = balance - $1, version = version + 1
WHERE id = $2 AND version = $3 AND status = 'ACTIVE'
```

При конкурентных обновлениях — retry до 3 попыток с перечтением версии.

### Идемпотентность

Уникальный `idempotency_key` на каждый перевод. Повторный запрос возвращает существующий результат.

### ACID

Все debit/credit выполняются в одной транзакции БД: `BEGIN → INSERT → UPDATE (debit) → UPDATE (credit) → COMMIT`.

---

## Команды

```bash
make help              # Все команды
make db                # PostgreSQL + Redis
make build             # Собрать сервер
make run               # Запустить сервер
make lint              # golangci-lint
make test              # Тесты
make antifraud         # C++ + Python движки
make demo              # Открыть демо
make logs              # Логи сервера
make clean             # Очистить
```

---

## Конфигурация

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/qw_pay?sslmode=disable` | PostgreSQL |
| `JWT_SECRET` | `change-me-in-production` | JWT signing key |
| `PORT` | `8080` | HTTP port |
| `REDIS_ADDR` | `127.0.0.1:6379` | Redis address |

---

## Документация

- [**API Reference**](docs/API.md) — Все эндпоинты с примерами запросов/ответов
- [**Architecture**](docs/ARCHITECTURE.md) — Глубокое погружение в архитектуру
- [**Development Guide**](docs/DEVELOPMENT.md) — Установка, запуск, добавление фич
- [**Technical Specification**](docs/TECHNICAL_SPECIFICATION.md) — Техническое задание

---

## Лицензия

MIT License — см. [LICENSE](LICENSE).
