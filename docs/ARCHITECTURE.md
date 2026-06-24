# Архитектура

## Обзор

QW Pay построен по принципу чистой архитектуры с разделением на сервисы. Каждый сервис имеет три слоя: Handler → Service → Repository.

---

## Микросервисная структура

### Go сервисы

```
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway                            │
│                  (cmd/server/main.go)                       │
│              Роутинг, JWT, инициализация                    │
└───────────┬──────────────┬───────────────┬─────────────────┘
            │              │               │
    ┌───────▼──────┐ ┌────▼─────┐ ┌───────▼──────┐
    │  Auth        │ │ Account  │ │ Transaction  │
    │  Service     │ │ Service  │ │   Service    │
    └───────┬──────┘ └────┬─────┘ └───────┬──────┘
            │              │               │
            │              │       ┌───────▼──────┐
            │              │       │   Exchange   │
            │              │       │   Provider   │
            │              │       │ (Frankfurter)│
            │              │       └──────────────┘
            └──────────────┼───────────────────┘
                           │
                ┌──────────▼──────────┐
                │     PostgreSQL      │
                └─────────────────────┘
```

### Антифрод (внешние сервисы)

```
┌─────────────────┐
│  Go Backend     │
│  (Transaction)  │
└────────┬────────┘
         │ Redis Queue
    ┌────▼────┐
    │  Redis  │
    └────┬────┘
    ┌────┴────────────────┐
    │                     │
┌───▼────────┐   ┌───────▼──────┐
│ C++ Engine │   │ Python Engine│
│ (fast)     │   │ (deep)       │
└────────────┘   └──────────────┘
```

---

## Слои архитектуры

### Handler (HTTP)

Отвечает за:
- Валидацию входящих запросов
- Маршрутизацию
- Формирование HTTP-ответов

### Service (Бизнес-логика)

Отвечает за:
- Бизнес-правила
- Оркестрацию операций
- Логирование

### Repository (Доступ к данным)

Отвечает за:
- SQL-запросы
- Работу с транзакциями БД
- Маппинг моделей

---

## Поток перевода

```
1. Клиент → POST /api/v1/transactions
2. Handler проверяет JWT (middleware)
3. Handler проверяет ownership счёта
4. Handler отправляет в Redis queue (anti-fraud)
5. C++ engine проверяет velocity, лимиты, блэклисты
6. Python engine проверяет паттерны, scoring
7. Результат → Redis verdict
8. Go backend получает verdict
9. Если approved → ACID транзакция:
   a. BEGIN
   b. UPDATE accounts (debit) WHERE version = X
   c. UPDATE accounts (credit) WHERE version = Y
   d. INSERT INTO transactions
   e. COMMIT
10. Клиент ← 200 OK / 403 Forbidden
```

---

## Модели данных

### users
- `id` (UUID, PK)
- `email` (VARCHAR, UNIQUE)
- `phone` (VARCHAR, UNIQUE)
- `password_hash` (VARCHAR)
- `role` (ENUM: USER, ADMIN)
- `is_verified` (BOOLEAN)
- `created_at`, `updated_at` (TIMESTAMPTZ)

### accounts
- `id` (UUID, PK)
- `user_id` (UUID, FK → users)
- `currency` (ENUM: RUB, USD, EUR)
- `balance` (NUMERIC(18,2))
- `version` (INT) — для optimistic locking
- `status` (ENUM: ACTIVE, BLOCKED)
- `created_at`, `updated_at` (TIMESTAMPTZ)

### transactions
- `id` (UUID, PK)
- `idempotency_key` (VARCHAR, UNIQUE)
- `from_account_id` (UUID, FK → accounts)
- `to_account_id` (UUID, FK → accounts)
- `amount` (NUMERIC(18,2))
- `currency` (VARCHAR(3))
- `source_currency` (VARCHAR(3))
- `exchange_rate_used` (NUMERIC(18,6))
- `status` (ENUM: PENDING, EXECUTED, REJECTED)
- `created_at` (TIMESTAMPTZ)

---

## Паттерны

### Optimistic Locking

Каждый счёт имеет поле `version`. При обновлении баланса:

```sql
UPDATE accounts
SET balance = balance - $1, version = version + 1
WHERE id = $2 AND version = $3 AND status = 'ACTIVE'
```

Если `RowsAffected() == 0` → конфликт, повторить.

### Идемпотентность

Каждый перевод имеет уникальный `idempotency_key`. При повторном запросе возвращается существующий результат без создания дубликата.

### ACID

Все операции списания/зачисления выполняются в одной транзакции БД:

```sql
BEGIN;
  -- Дебет отправителя
  UPDATE accounts SET balance = balance - 100, version = version + 1
  WHERE id = 'from_id' AND version = 1;

  -- Кредит получателя
  UPDATE accounts SET balance = balance + 100, version = version + 1
  WHERE id = 'to_id' AND version = 1;

  -- Запись транзакции
  INSERT INTO transactions (...) VALUES (...);

COMMIT;
```

---

## Антифрод

### C++ Engine

**Задача:** Быстрые проверки (< 5ms)

| Проверка | Лимит | Risk |
|----------|-------|------|
| Velocity (мин) | 5 транзакций | +85 |
| Velocity (час) | 20 транзакций | +80 |
| Сумма за перевод | 500,000 | +90 |
| Сумма за день | 2,000,000 | +75 |
| Блэклист юзера | — | +100 |
| Блэклист счёта | — | +100 |
| Круговые переводы | 10 уникальных | +70 |
| Round amount ≥ 1000 | — | +30 (flag) |

### Python Engine

**Задача:** Глубокий анализ и scoring

| Проверка | Risk |
|----------|------|
| Аккаунт < 1 часа | +40 |
| Аккаунт < 24 часов | +15 |
| Время 2:00-5:00 | +20 |
| Сумма > 5x от средней | +25 |
| > 15 уникальных получателей | +30 |
| Интервал < 10 сек | +35 |
| Round amount ≥ 10,000 | +20 |
| > 20 входящих в час | +25 |

**Вердикт:** risk >= 60 → BLOCK, risk < 60 → APPROVE

---

## Инфраструктура

### Docker Compose

```yaml
services:
  db:         PostgreSQL 16
  redis:      Redis 7
  app:        Go сервер
  antifraud-cpp:    C++ движок
  antifraud-python: Python анализ
```

### Логирование

Логи пишутся в:
- `stdout` (для Docker)
- `logs/server.log` (для разработки)

### Мониторинг

- Статус подключения к PostgreSQL
- Статус подключения к Redis
- Антифрод вердикты в логах
- Время обработки запросов
