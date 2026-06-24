# AGENTS.md — Инструкция для ИИ-ассистентов

Этот файл содержит инструкции для AI-ассистентов, работающих с проектом QW Pay.
При начале работы ОБЯЗАТЕЛЬНО прочитайте этот файл целиком.

---

## 0. AI Скиллы

Для работы с проектом доступны специализированные скиллы в `skills/`:

| Скилл | Файл | Триггер |
|-------|------|---------|
| **AI Coordinator** | `skills/ai-coordinator/SKILL.md` | Координация всех ИИ, оркестрация задач |
| **Security Engineer** | `skills/security-engineer/SKILL.md` | Аудит безопасности, проверка уязвимостей |
| **API Platform** | `skills/api-platform/SKILL.md` | REST API дизайн, документация |
| **QA & Test Automation** | `skills/qa-test-automation/SKILL.md` | Тесты, покрытие, edge cases |
| **DevOps & Cloud** | `skills/devops-cloud/SKILL.md` | Docker, CI/CD, инфраструктура |
| **Data Analytics** | `skills/data-analytics/SKILL.md` | Аналитика, метрики, SQL запросы |
| **Frontend Architect** | `skills/frontend-architect/SKILL.md` | UI/UX, компоненты, стили |
| **Backend Architect** | `skills/backend-architect/SKILL.md` | Go архитектура, сервисы, паттерны |
| **App Architecture** | `skills/app-architecture/SKILL.md` | Архитектурные решения, паттерны |
| **App Structure** | `skills/app-structure/SKILL.md` | Навигация по проекту, расположение файлов |
| **Computer Science** | `skills/computer-science/SKILL.md` | Алгоритмы, структуры данных, сложность |

**Использование:** Прочитайте нужный SKILL.md перед началом работы в соответствующей области.

---

## 1. Обзор проекта

QW Pay — микросервисная платёжная система для внутренних банковских переводов.

**Стек:**
- Go 1.23+ с Gin framework
- PostgreSQL 16 (pgx v5 driver)
- Redis 7 (anti-fraud очереди)
- C++ и Python (anti-fraud движки)
- Docker + Docker Compose
- GitHub Actions CI/CD

**Структура:**
```
cmd/server/main.go          — точка входа
internal/
  account/                   — CRUD счетов, баланс
  antifraud/                 — Redis-клиент + handler для anti-fraud
  auth/                      — JWT + OTP авторизация + refresh tokens
  config/                    — загрузка конфига из env
  contextkeys/               — ключи для gin.Context
  database/                  — подключение к PostgreSQL
  errors/                    — типизированные ошибки приложения
  exchange/                  — реальные курсы валют (Frankfurter API)
  logger/                    — structured logging (slog)
  middleware/                 — AuthRequired, RequestID, CORS
  model/                     — модели данных (User, Account, Transaction, RefreshToken)
  ratelimit/                 — in-memory rate limiter
  response/                  — единый формат JSON-ответов
  transaction/               — логика переводов с optimistic locking
antifraud/
  cpp/                       — C++ anti-fraud движок
  orchestrator.py            — Python оркестратор
migrations/                  — SQL миграции (001_init.sql, 002_refresh_tokens.sql)
web/                         — demo frontend
```

---

## 2. Команды разработки

```bash
# Инфраструктура
make db                      # запустить PostgreSQL + Redis
make stop                    # остановить всё

# Сервер
make build                   # собрать бинарник
make run                     # собрать и запустить
make lint                    # golangci-lint
make test                    # go test -v ./...

# Anti-fraud
make antifraud-build         # собрать C++ движок
make antifraud               # запустить C++ + Python

# Проверка
go vet ./...                 # статический анализ
go test -race ./...          # тесты с race detector
~/go/bin/staticcheck ./...   # staticcheck
```

---

## 3. Конвенции кода

### Язык и стиль
- **Язык комментариев:** русский (для README, API docs). Английский — для кода.
- **Naming:** Go conventions — exported PascalCase, unexported camelCase.
- **Errors:** всегда оборачивать через `fmt.Errorf("context: %w", err)` или `apperr.Wrap()`.
- **Логирование:** только через `logger` package (slog). Никогда `fmt.Println` или `log.Println`.
- **Тесты:** `func TestXxx(t *testing.T)`. Моки — в том же package. Не использовать gomock, testify.

### Архитектурные правила
- **Dependency injection:** всё через constructor (`NewService(repo)`), никаких глобальных зависимостей внутри пакетов.
- **Interfaces:** минимальные (1-3 метода). Определять в consumer package, не в provider.
- **Error types:** использовать `internal/errors` — `apperr.NotFound()`, `apperr.BadRequest()`, `apperr.Wrap()` и т.д.
- **Response format:** только через `internal/response` — `response.OK()`, `response.Created()`, `response.Error()`, `response.Paginated()`.

### Деньги
- **ВСЕ денежные значения** — `shopspring/decimal.Decimal`. Никогда `float64`.
- В моделях: `Balance decimal.Decimal`, `Amount decimal.Decimal`.
- В config: `MaxTransferAmount decimal.Decimal`, `DailyLimit decimal.Decimal`.
- Сравнения: `amount.GreaterThan(limit)`, `amount.Add(other)`.
- JSON маршализация: автоматически как строка (`"100.50"`), это нормально.

### Транзакции
- DB транзакции через `db.Begin(ctx)` + `defer tx.Rollback(ctx)`.
- Optimistic locking: `WHERE version=$N`, проверка `RowsAffected() == 0` → `ErrOptimisticLock`.
- Кросс-валютные переводы: debit — оригинальная сумма, credit — `amount * exchangeRate`.

### JWT и авторизация
- Access token: короткий (по умолчанию 24ч), содержит `sub` (user ID).
- Refresh token: длинный (по умолчанию 30 дней), хранится в БД (hash SHA-256).
- При refresh: старый токен revokается, выдаётся новая пара.
- Эндпоинты: `/api/v1/login`, `/api/v1/refresh`, `/api/v1/logout`.

---

## 4. Безопасность

- **OTP** — никогда не логировать в открытом виде.
- **Passwords** — bcrypt, min 6 символов.
- **JWT Secret** — из env, не хардкодить.
- **SQL injection** — все запросы через параметризованные `$1, $2...`.
- **Rate limiting** — на auth endpoints (10 req/s, burst 20).
- **CORS** — настроен через middleware, ограничения по origins.
- **Anti-fraud timeout** — default approved (настраивается).

---

## 5. CI/CD

### GitHub Actions (`.github/workflows/ci.yml`)
- **Go version:** `1.23` (пин, НЕ `go-version-file: go.mod` — конфликтует с зависимостями).
- **Jobs:** Test → Lint → Build → Release.
- **Release:** тег `v1.0.${{ github.run_number }}`, binary upload.
- **Триггеры:** push to main, PR to main.

### При изменении зависимостей
1. `go mod tidy`
2. `go test -race ./...`
3. `go vet ./...`
4. `staticcheck ./...`
5. Если всё зелёное — коммитить.

---

## 6. Частые ошибки

| Ошибка | Решение |
|---|---|
| `float64` для денег | Заменить на `decimal.Decimal` |
| `log.Fatal` в пакетах | Использовать `logger.Error()` + `panic()` |
| Глобальные переменные | DI через constructor |
| OTP в логах | Убрать OTP из `slog.Info` |
| OFFSET/LIMIT пагинация | Для больших таблиц рассмотреть cursor-based |
| Горутина без остановки | Добавить `context.WithCancel` + `Stop()` |
| `go-version-file: go.mod` в CI | Использовать `go-version: '1.23'` (конфликт зависимостей) |

---

## 7. Тестирование

- **Unit tests** — в каждом package (кроме database, logger, contextkeys).
- **Моки** — в том же package, ручные (не codegen).
- **Init functions** — для настройки `config.C` в тестах.
- **Race detector** — `go test -race ./...` обязательно перед коммитом.
- **Интеграционные тесты** — требуют PostgreSQL + Redis (make db).

---

## 8. Git workflow

- **Branch:** `main` — стабильная, CI/CD автоматически деплоит.
- **Коммиты:** conventional commits — `feat:`, `fix:`, `refactor:`, `docs:`.
- **PR:** требуют прохождения Test + Lint + Build.
- **Релизы:** автоматические через GitHub Actions (v1.0.N).
- **Не коммитить:** `.env`, `logs/`, `*.log`, `vendor/`, `tmp/`.

---

## 9. Контакты и ссылки

- **Repo:** https://github.com/qqwozz/qw_pay
- **CI/CD:** https://github.com/qqwozz/qw_pay/actions
- **Releases:** https://github.com/qqwozz/qw_pay/releases
