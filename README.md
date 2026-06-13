<div align="center">

# QW Pay

**Микросервисная платёжная система для внутренних банковских переводов**

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis&logoColor=white)](https://redis.io/)
[![C++](https://img.shields.io/badge/C++-17-00599C?style=flat&logo=cplusplus&logoColor=white)](https://isocpp.org/)
[![Python](https://img.shields.io/badge/Python-3.12-3776AB?style=flat&logo=python&logoColor=white)](https://www.python.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

</div>

---

## Описание

QW Pay — микросервисная платёжная система для внутренних банковских переводов с:

- Регистрацией пользователей с OTP-подтверждением
- JWT-аутентификацией
- Счетами в RUB / USD / EUR с приветственным бонусом 100
- Внутренними переводами с конвертацией по фиксированным курсам
- Идемпотентностью транзакций
- Optimistic locking через version
- ACID-транзакциями при списании/зачислении
- Антифрод-системой на C++ и Python

---

## Архитектура

```
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway                              │
│                    (Go + Gin + JWT)                             │
└───────────┬──────────────┬───────────────┬─────────────────────┘
            │              │               │
    ┌───────▼──────┐ ┌────▼─────┐ ┌───────▼──────┐
    │  Auth        │ │ Account  │ │ Transaction  │
    │  Service     │ │ Service  │ │   Service    │
    │  (register,  │ │ (CRUD,   │ │ (transfer,   │
    │   login,     │ │  block,  │ │  idempotent, │
    │   OTP, JWT)  │ │  bonus)  │ │  ACID)       │
    └───────┬──────┘ └────┬─────┘ └───────┬──────┘
            │              │               │
            └──────────────┼───────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
   ┌─────▼─────┐   ┌──────▼──────┐   ┌──────▼──────┐
   │ PostgreSQL│   │    Redis    │   │ Anti-Fraud  │
   │  (ACID)   │   │  (queue)    │   │ C++ + Python│
   └───────────┘   └─────────────┘   └─────────────┘
```

---

## Быстрый старт

### Требования

- Go 1.22+
- Docker & Docker Compose
- Redis (для антифрода)
- g++ и libhiredis-dev (для C++ движка)
- Python 3.12+ (для Python анализа)

### Запуск

```bash
# 1. Клонировать репозиторий
git clone https://github.com/your-org/qw_pay.git
cd qw_pay

# 2. Запустить PostgreSQL и Redis
make db

# 3. Запустить сервер
make run

# 4. (Опционально) Запустить антифрод
make antifraud

# 5. Открыть демо
make demo
```

### Демо-страница

Откройте http://localhost:8080/demo в браузере для интерактивного демо.

---

## Структура проекта

```
qw_pay/
├── cmd/server/              # Точка входа
│   └── main.go
├── internal/
│   ├── auth/                # Сервис аутентификации
│   │   ├── handler.go       # HTTP-хендлеры
│   │   ├── service.go       # Бизнес-логика
│   │   └── repository.go    # Работа с БД
│   ├── account/             # Сервис счетов
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   ├── transaction/         # Сервис переводов
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   ├── antifraud/           # Антифрод клиент
│   │   └── client.go
│   ├── config/              # Конфигурация
│   ├── database/            # Пул соединений
│   ├── middleware/          # JWT middleware
│   └── model/               # Модели данных
├── antifraud/
│   ├── cpp/                 # C++ движок (velocity, лимиты)
│   ├── python/              # Python анализ (scoring, паттерны)
│   └── orchestrator.py      # Оркестратор движков
├── web/                     # Фронтенд
├── migrations/              # SQL-схема
├── deploy/                  # Docker-конфигурации
├── docs/                    # Документация
├── Makefile                 # Команды разработчика
└── README.md
```

---

## API Endpoints

### Публичные

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/register` | Регистрация |
| `POST` | `/api/v1/verify` | Подтверждение OTP |
| `POST` | `/api/v1/login` | Вход (JWT) |

### Защищённые (Bearer token)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/accounts` | Создать счёт (+100 бонус) |
| `GET` | `/api/v1/accounts` | Список счетов |
| `POST` | `/api/v1/accounts/:id/block` | Заблокировать счёт |
| `POST` | `/api/v1/transactions` | Создать перевод |
| `GET` | `/api/v1/transactions` | История переводов |
| `POST` | `/api/v1/antifraud/block-user` | Заблокировать юзера в AF |
| `POST` | `/api/v1/antifraud/block-account` | Заблокировать счёт в AF |

---

## Антифрод-система

### C++ движок (быстрые проверки)

- Velocity: макс 5 транзакций/мин, 20/час
- Лимит суммы: 500K за перевод, 2M в день
- Блэклисты пользователей и счетов
- Подозрительные круговые переводы
- Детекция round amount (структуринг)

### Python анализ (глубокий scoring)

- Возраст аккаунта (< 24ч = +15 risk)
- Необычные часы (2-5 утра = +20 risk)
- Отклонение суммы от среднего
- Много уникальных получателей
- Скоринг: risk >= 60 → блокировка

---

## Команды

```bash
make help            # Показать все команды
make db              # Запустить PostgreSQL + Redis
make build           # Собрать сервер
make run             # Запустить сервер
make antifraud       # Запустить антифрод движки
make demo            # Открыть демо в браузере
make logs            # Смотреть логи
make test            # Запустить тесты
make clean           # Очистить бинарники
```

---

## Конфигурация

Переменные окружения (`.env`):

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/qw_pay?sslmode=disable` | Строка подключения к БД |
| `JWT_SECRET` | `change-me-in-production` | Секрет для JWT |
| `PORT` | `8080` | Порт сервера |
| `REDIS_ADDR` | `127.0.0.1:6379` | Адрес Redis |

---

## Документация

- [Документация API](docs/API.md)
- [Архитектура](docs/ARCHITECTURE.md)
- [Руководство разработчика](docs/DEVELOPMENT.md)
- [Техническое задание](docs/TECHNICAL_SPECIFICATION.md)

---

## Лицензия

MIT License — см. файл [LICENSE](LICENSE).
