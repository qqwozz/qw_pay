# Руководство разработчика

## Требования

- Go 1.22+
- PostgreSQL 16+ (или Docker)
- Redis 7+ (для антифрода)
- g++ и libhiredis-dev (для C++ движка)
- Python 3.12+ (для Python анализа)

---

## Установка

### macOS

```bash
# Homebrew
brew install go postgresql redis python3

# libhiredis (для C++)
brew install hiredis
```

### Ubuntu/Debian

```bash
# Go
sudo snap install go --classic

# PostgreSQL + Redis
sudo apt install postgresql redis-server

# C++ зависимости
sudo apt install g++ libhiredis-dev

# Python
sudo apt install python3 python3-pip
pip3 install redis
```

### Arch Linux

```bash
# Go
sudo pacman -S go

# PostgreSQL + Redis
sudo pacman -S postgresql redis

# C++ зависимости
sudo pacman -S hiredis

# Python
sudo pacman -S python python-pip
pip install redis
```

---

## Настройка

### 1. Создайте .env файл

```bash
cp .env.example .env
```

### 2. Настройте переменные

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/qw_pay?sslmode=disable
JWT_SECRET=your-secret-key-here
PORT=8080
REDIS_ADDR=127.0.0.1:6379
```

### 3. Запустите PostgreSQL и Redis

```bash
make db
```

### 4. Инициализируйте БД

```bash
psql postgres://postgres:postgres@localhost:5432/qw_pay < migrations/001_init.sql
```

---

## Запуск

### Только сервер

```bash
make run
```

### Сервер + Антифрод

```bash
# Терминал 1: Сервер
make run

# Терминал 2: Антифрод
make antifraud
```

### Docker (всё вместе)

```bash
cd deploy
docker-compose up -d
```

---

## Структура кода

```
internal/
├── auth/              # Сервис аутентификации
│   ├── handler.go     # HTTP-хендлеры
│   ├── service.go     # Бизнес-логика
│   └── repository.go  # Доступ к БД
├── account/           # Сервис счетов
│   ├── handler.go
│   ├── service.go
│   └── repository.go
├── transaction/       # Сервис переводов
│   ├── handler.go
│   ├── service.go
│   └── repository.go
├── antifraud/         # Антифрод клиент
│   └── client.go
├── config/            # Конфигурация
├── database/          # Пул соединений
├── middleware/        # JWT middleware
└── model/             # Модели данных
```

---

## Добавление нового эндпоинта

### 1. Создайте репозиторий (если нужен доступ к БД)

```go
// internal/myfeature/repository.go
package myfeature

type Repository struct {
    db *pgxpool.Pool
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*model.Thing, error) {
    // SQL запрос
}
```

### 2. Создайте сервис

```go
// internal/myfeature/service.go
package myfeature

type Service struct {
    repo *Repository
}

func (s *Service) DoSomething(ctx context.Context, id uuid.UUID) error {
    // Бизнес-логика
}
```

### 3. Создайте хендлер

```go
// internal/myfeature/handler.go
package myfeature

type Handler struct {
    svc *Service
}

func (h *Handler) HandleRequest(c *gin.Context) {
    // Валидация, вызов сервиса, ответ
}
```

### 4. Зарегистрируйте в main.go

```go
myFeatureRepo := myfeature.NewRepository(database.Pool)
myFeatureSvc := myfeature.NewService(myFeatureRepo)
myFeatureH := myfeature.NewHandler(myFeatureSvc)

v1.POST("/my-endpoint", myFeatureH.HandleRequest)
```

---

## Тестирование

### Запуск тестов

```bash
make test
```

### Ручное тестирование

```bash
# Регистрация
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","phone":"+79001234567","password":"secret123"}'

# Получите OTP из логов (logs/server.log)

# Подтверждение
curl -X POST http://localhost:8080/api/v1/verify \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","otp_code":"123456"}'

# Вход
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"secret123"}'
# Сохраните access_token

# Создание счёта
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"currency":"USD"}'

# Перевод
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "from_account_id":"<from-uuid>",
    "to_account_id":"<to-uuid>",
    "amount":10,
    "idempotency_key":"test-key-1"
  }'
```

---

## Логи

Логи пишутся в `logs/server.log` и stdout.

```bash
# Смотреть логи в реальном времени
make logs

# Или
tail -f logs/server.log
```

---

## Антифрод

### C++ движок

```bash
# Собрать
make antifraud-build

# Запустить
./antifraud/cpp/fraud_engine
```

### Python движок

```bash
# Запустить
python3 antifraud/python/service.py
```

### Оркестратор (оба движка)

```bash
make antifraud
```

---

## Полезные команды

```bash
make help            # Все команды
make build           # Собрать сервер
make run             # Запустить сервер
make db              # Запустить БД + Redis
make stop            # Остановить контейнеры
make antifraud       # Запустить антифрод
make demo            # Открыть демо
make logs            # Смотреть логи
make test            # Запустить тесты
make clean           # Очистить
```

---

## Исправление проблем

### "connection refused" к PostgreSQL

```bash
# Проверьте, что PostgreSQL запущен
docker-compose ps

# Перезапустите
make stop && make db
```

### "connection refused" к Redis

```bash
# Запустите Redis
redis-server --daemonize yes

# Или через Docker
docker run -d -p 6379:6379 redis:7-alpine
```

### Ошибка сборки C++

```bash
# Установите зависимости
# macOS:
brew install hiredis

# Ubuntu:
sudo apt install g++ libhiredis-dev
```
