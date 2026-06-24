# ROADMAP — QW Pay

## Текущее состояние

| Метрика | Значение |
|---------|----------|
| Покрытие тестов | ~25% |
| Govulncheck | FAIL (уязвимости в Go stdlib/deps) |
| Go версия | 1.23 |
| CI/CD | Test + Lint + Build + Release (PASS) |

---

## Phase 1: Качество кода (приоритет)

### 1.1 Обновление зависимостей
- [ ] Обновить Go до 1.24+
- [ ] Обновить gin, pgx, go-redis, jwt до последних версий
- [ ] Обновить golang.org/x/* пакеты
- [ ] Govulncheck должен проходить без ошибок

### 1.2 Покрытие тестов → 80%+
- [ ] Handler тесты (auth, account, transaction) — HTTP integration tests через httptest
- [ ] Repository тесты — require PostgreSQL (integration tests)
- [ ] Exchange provider — мокать HTTP, тестировать fallback
- [ ] Middleware тесты — CORS, rate limiter edge cases
- [ ] Anti-fraud handler тесты
- [ ] Concurrent transfer race condition тесты

### 1.3 Refactoring
- [ ] Вынести exchange rates в отдельный конфиг/файл
- [ ] Unified error handling: response.Error принимать *apperr.AppError
- [ ] Убрать дублирование JWT логики (service + middleware)
- [ ] Добавить request context timeout на handler level

---

## Phase 2: Безопасность

### 2.1 Критично
- [ ] Password validation: min 8 chars, uppercase, digit, special
- [ ] OTP rate limit: max 3 попытки в час на email
- [ ] Login rate limit: max 5 попыток в минуту на IP
- [ ] Account lockout после 10 неудачных попыток
- [ ] Audit log для финансовых операций

### 2.2 Важно
- [ ] Refresh token: одноразовый (rotate on refresh)
- [ ] Token blacklist в Redis (при logout — добавить в blacklist)
- [ ] Request signing для anti-fraud API
- [ ] HTTPS/TLS termination в Docker
- [ ] Secret rotation mechanism

### 2.3 Nice-to-have
- [ ] 2FA (TOTP) для админов
- [ ] IP whitelist для admin endpoints
- [ ] Anomaly detection на паттерны входов

---

## Phase 3: Функциональность

### 3.1 API
- [ ] `GET /accounts/:id` — получить конкретный счёт
- [ ] `GET /transactions/:id` — получить конкретный перевод
- [ ] `GET /exchange-rates` — текущие курсы валют
- [ ] `POST /accounts/:id/deposit` — пополнение счёта
- [ ] `GET /users/me` — профиль текущего пользователя

### 3.2 Бизнес-логика
- [ ] Scheduled transfers (cron)
- [ ] Transfer categories/tags
- [ ] Account statements (PDF/CSV export)
- [ ] Multi-user accounts (joint accounts)
- [ ] Recurring payments

### 3.3 Anti-fraud
- [ ] Geolocation-based risk scoring
- [ ] Device fingerprinting
- [ ] ML model integration (Python engine)
- [ ] Real-time alerts (webhook/email)

---

## Phase 4: DevOps & Infra

### 4.1 Мониторинг
- [ ] Prometheus metrics (request latency, error rate, transfer volume)
- [ ] Grafana dashboards
- [ ] Structured logging → ELK/Loki
- [ ] Alerting (PagerDuty/Slack)

### 4.2 Deployment
- [ ] Kubernetes manifests (Helm chart)
- [ ] Blue-green deployment
- [ ] Database migrations in CI/CD
- [ ] Health check → readiness/liveness probes

### 4.3 Performance
- [ ] Connection pooling tuning
- [ ] Redis caching для exchange rates
- [ ] Cursor-based pagination (вместо OFFSET)
- [ ] Batch operations для anti-fraud

---

## Phase 5: Developer Experience

### 5.1 Documentation
- [ ] OpenAPI/Swagger spec
- [ ] Postman collection
- [ ] Architecture Decision Records (ADR)
- [ ] Contributing guide

### 5.2 Tooling
- [ ] Makefile: `make migrate`, `make seed`
- [ ] Docker: multi-stage build optimization
- [ ] Pre-commit hooks (golangci-lint, gofmt)
- [ ] Dependabot/Renovate для автоматических обновлений

### 5.3 Testing
- [ ] E2E тесты (Playwright + Go)
- [ ] Load testing (k6/vegeta)
- [ ] Chaos engineering (Redis down, DB failover)

---

## Приоритеты

| Приоритет | Задача | Effort |
|-----------|--------|--------|
| P0 | Обновить зависимости (govulncheck) | 1-2 дня |
| P0 | Handler тесты → 50%+ | 3-5 дней |
| P1 | Password validation | 1 день |
| P1 | OTP rate limit | 1 день |
| P1 | GET /accounts/:id, GET /transactions/:id | 1 день |
| P2 | Prometheus metrics | 2-3 дня |
| P2 | OpenAPI spec | 2 дня |
| P3 | Kubernetes deployment | 1 неделя |
