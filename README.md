<div align="center">

<picture>
  <source height="125" media="(prefers-color-scheme: dark)" srcset="web/logo-dark.svg">
  <img height="125" alt="QW Pay" src="web/logo.svg">
</picture>

<br/>

# **QW Pay**

### Микросервисная платёжная система для внутренних банковских переводов

<br/>

[![CI/CD](https://img.shields.io/github/actions/workflow/status/qqwozz/qw_pay/ci.yml?branch=main&style=for-the-badge&logo=github&label=CI%2FCD&logoColor=white)](https://github.com/qqwozz/qw_pay/actions)
[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-22c55e?style=for-the-badge)](#)
[![Coverage](https://img.shields.io/badge/Coverage-25%25-yellow?style=for-the-badge)](#)

<br/>

[**Быстрый старт**](#-быстрый-старт) · [**API**](#-api-endpoints) · [**Архитектура**](#-архитектура) · [**Документация**](docs/)

</div>

---

<br/>

## Возможности

<table>
<tr>
<td width="33%">

### 🔐 Аутентификация

- JWT access + refresh tokens
- OTP-подтверждение email
- Роли: USER, ADMIN
- Rate limiting

</td>
<td width="33%">

### 💰 Счета

- Мультивалютные (RUB/USD/EUR)
- Optimistic locking
- Приветственный бонус +100
- Блокировка счёта

</td>
<td width="33%">

### 🔄 Переводы

- ACID-транзакции
- Идемпотентность
- Конвертация по реальным курсам (Frankfurter API)
- Пагинация

</td>
</tr>
<tr>
<td>

### 🛡️ Антифрод

- C++ движок (<1ms)
- Python scoring
- Velocity checks
- Blacklists

</td>
<td>

### 🧮 Технологии

- `shopspring/decimal`
- REST JSON API
- Structured logging
- Token bucket

</td>
<td>

### 🚀 DevOps

- GitHub Actions CI/CD
- Docker Compose
- Auto-release
- Health checks

</td>
</tr>
</table>

---

<br/>

## Быстрый старт

### Требования

```
Go 1.23+  •  Docker  •  Make (опционально)
```

### Установка

```bash
# Клонируем репозиторий
git clone https://github.com/qqwozz/qw_pay.git
cd qw_pay

# Запускаем инфраструктуру
make db

# Собираем и запускаем сервер
make run
```

### Демо

Откройте **http://localhost:8080/demo** — интерактивная страница со всеми эндпоинтами.

---

<br/>

## Архитектура

```
                          ┌─────────────────────┐
                          │    🌐 API Gateway    │
                          │  Gin + JWT + CORS    │
                          └──────────┬──────────┘
                                     │
             ┌───────────────────────┼───────────────────────┐
             │                       │                       │
      ┌──────▼──────┐        ┌──────▼──────┐        ┌──────▼──────┐
      │    🔑 Auth   │        │   💳 Account │        │ 💸 Trans.   │
      │   Service    │        │   Service    │        │   Service   │
      │              │        │              │        │             │
      │  • Register  │        │  • Create    │        │  • Transfer │
      │  • Login     │        │  • List      │        │  • ACID     │
      │  • OTP       │        │  • Block     │        │  • Idempot. │
      │  • Refresh   │        │  • Bonus     │        │  • Cross-cur│
      └──────┬──────┘        └──────┬──────┘        └──────┬──────┘
             │                       │                       │
             └───────────────────────┼───────────────────────┘
                                     │
           ┌─────────────────────────┼─────────────────────────┐
           │                         │                         │
    ┌──────▼──────┐           ┌──────▼──────┐           ┌──────▼──────┐
    │ 🐘 Postgres │           │  🔴 Redis   │           │ 🛡️ Anti-    │
    │   (ACID)    │           │  (queues)   │           │   Fraud     │
    └─────────────┘           └─────────────┘           └─────────────┘
```

---

<br/>

## Технологический стек

<p align="center">
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  </a>
  <a href="https://gin-gonic.com/">
    <img src="https://img.shields.io/badge/Gin-000000?style=for-the-badge&logo=gin&logoColor=white" alt="Gin">
  </a>
  <a href="https://www.postgresql.org/">
    <img src="https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
  </a>
  <a href="https://redis.io/">
    <img src="https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white" alt="Redis">
  </a>
  <a href="https://www.docker.com/">
    <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
  </a>
  <a href="https://github.com/shopspring/decimal">
    <img src="https://img.shields.io/badge/shopspring/decimal-5947AB?style=for-the-badge" alt="Decimal">
  </a>
</p>

---

<br/>

## API Endpoints

### 🌐 Публичные

| Метод | Путь | Описание | Тело запроса |
|:-----:|------|----------|:------------|
| `POST` | `/api/v1/register` | Регистрация | `email`, `phone`, `password` |
| `POST` | `/api/v1/verify` | Подтверждение OTP | `email`, `otp_code` |
| `POST` | `/api/v1/login` | Вход в систему | `email`, `password` |
| `POST` | `/api/v1/refresh` | Обновление токенов | `refresh_token` |

### 🔒 Защищённые

| Метод | Путь | Описание | Тело запроса |
|:-----:|------|----------|:------------|
| `POST` | `/api/v1/logout` | Выход | — |
| `POST` | `/api/v1/accounts` | Создать счёт | `currency` |
| `GET` | `/api/v1/accounts` | Список счетов | — |
| `POST` | `/api/v1/accounts/:id/block` | Заблокировать счёт | — |
| `POST` | `/api/v1/transactions` | Создать перевод | `from_account_id`, `to_account_id`, `amount`, `idempotency_key` |
| `GET` | `/api/v1/transactions` | История переводов | `?page=1&page_size=20` |
| `POST` | `/api/v1/antifraud/block-user` | Блокировка пользователя | `user_id` |
| `POST` | `/api/v1/antifraud/block-account` | Блокировка счёта | `account_id` |

---

<br/>

## Структура проекта

```
qw_pay/
│
├── cmd/server/main.go              # 🚀 Точка входа, DI, graceful shutdown
│
├── internal/
│   ├── account/                     # 💳 CRUD счетов, optimistic locking
│   ├── antifraud/                   # 🛡️ Redis-клиент + anti-fraud handler
│   ├── auth/                        # 🔑 JWT + OTP + refresh tokens
│   ├── config/                      # ⚙️ Конфигурация из env
│   ├── contextkeys/                 # 📌 Ключи для gin.Context
│   ├── database/                    # 🐘 PostgreSQL пул соединений
│   ├── errors/                      # ❌ Типизированные ошибки
│   ├── exchange/                    # 💱 Реальные курсы валют (Frankfurter API)
│   ├── logger/                      # 📝 Structured logging (slog)
│   ├── middleware/                   # 🛡️ AuthRequired, RequestID, CORS
│   ├── model/                       # 📦 Модели данных
│   ├── ratelimit/                   # ⏱️ Token bucket rate limiter
│   ├── response/                    # 📨 Единый JSON-формат ответов
│   └── transaction/                 # 🔄 Переводы с конвертацией, ACID
│
├── antifraud/
│   ├── cpp/                         # ⚡ C++ движок (velocity, блэклисты)
│   └── orchestrator.py              # 🐍 Python оркестратор (scoring)
│
├── migrations/                      # 📊 SQL миграции
├── web/                             # 🌐 Demo frontend
├── docs/                            # 📚 Документация
│   ├── API.md                       # API reference
│   ├── ARCHITECTURE.md              # Архитектура системы
│   ├── DEVELOPMENT.md               # Руководство разработчика
│   └── TECHNICAL_SPECIFICATION.md   # Техническое задание
│
├── Makefile                         # 🔧 Команды сборки
├── Dockerfile                       # 🐳 Docker образ
├── docker-compose.yml               # 🐳 Локальная инфраструктура
└── .github/workflows/ci.yml         # ⚙️ GitHub Actions CI/CD
```

---

<br/>

## Конфигурация

| Переменная | По умолчанию | Описание |
|------------|-------------|:--------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/qw_pay` | Строка подключения к PostgreSQL |
| `JWT_SECRET` | — | Секрет для подписи JWT (обязательно!) |
| `PORT` | `8080` | Порт сервера |
| `REDIS_ADDR` | `127.0.0.1:6379` | Адрес Redis |
| `MAX_TRANSFER_AMOUNT` | `10000000` | Максимальная сумма перевода |
| `DAILY_LIMIT` | `50000000` | Дневной лимит |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:8080` | Разрешённые origins |

---

<br/>

## Команды

```bash
make help            # 📋 Список всех команд
make db              # 🐳 Запустить PostgreSQL + Redis (Docker)
make run             # 🚀 Собрать и запустить сервер
make test            # 🧪 Запустить тесты
make lint            # 🔍 Запустить linter
make antifraud       # 🛡️ Запустить C++ + Python движки
make demo            # 🌐 Открыть демо в браузере
make clean           # 🧹 Удалить бинарники
```

---

<br/>

## Антифрод-система

<table>
<tr>
<th width="50%">⚡ C++ Engine — Быстрые проверки (<1ms)</th>
<th width="50%">🧠 Python Engine — Глубокий scoring</th>
</tr>
<tr>
<td>

- Velocity: макс **5 транзакций/мин**, **20/час**
- Лимит суммы: **500K** за перевод, **2M** в день
- Блэклисты пользователей и счетов
- Детекция круговых переводов
- Round amount detection

</td>
<td>

- Возраст аккаунта (< 24ч = **+15 risk**)
- Необычные часы (2-5 утра = **+20 risk**)
- Отклонение суммы от среднего
- Много уникальных получателей
- **risk >= 60 → блокировка**

</td>
</tr>
</table>

---

<br/>

## Тестирование

```bash
# Все тесты
make test

# С race detector
go test -race ./...

# Конкретный пакет
go test -v ./internal/auth/...

# С покрытием
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

<br/>

## Курсы валют

Курсы загружаются в реальном времени из **Frankfurter API** (84 центральных банка, 200+ валют).

API бесплатное, не требует регистрации и обновляется ежедневно.

- Источник: [frankfurter.dev](https://frankfurter.dev/)
- Кэширование: 1 час
- Fallback: статические курсы при недоступности API

---

<br/>

## Статус проекта

<table>
<tr>
<td align="center" width="20%">
<img src="https://img.shields.io/badge/Аутентификация-✅_Готово-22c55e?style=for-the-badge" alt="Auth">
</td>
<td align="center" width="20%">
<img src="https://img.shields.io/badge/Счета-✅_Готово-22c55e?style=for-the-badge" alt="Accounts">
</td>
<td align="center" width="20%">
<img src="https://img.shields.io/badge/Переводы-✅_Готово-22c55e?style=for-the-badge" alt="Transfers">
</td>
<td align="center" width="20%">
<img src="https://img.shields.io/badge/Антифрод-✅_Готово-22c55e?style=for-the-badge" alt="Anti-fraud">
</td>
<td align="center" width="20%">
<img src="https://img.shields.io/badge/CI%2FCD-✅_Готово-22c55e?style=for-the-badge" alt="CI/CD">
</td>
</tr>
</table>

---

<br/>

## Лицензия

Распространяется под лицензией **MIT**. Смотрите [LICENSE](LICENSE) для подробной информации.

---

<br/>

<div align="center">

**[Наверх ↑](#qw-pay)**

<br/>

<a href="https://github.com/qqwozz/qw_pay">
  <img src="https://img.shields.io/github/stars/qqwozz/qw_pay?style=social" alt="GitHub Stars">
</a>

<br/>

Made with ❤️ by [qqwozz](https://github.com/qqwozz)

</div>
