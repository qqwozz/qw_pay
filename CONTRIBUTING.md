# Contributing

Спасибо за интерес к QW Pay! Вот как начать.

## Быстрый старт

```bash
git clone https://github.com/qqwozz/qw_pay.git
cd qw_pay
make db          # запустить PostgreSQL + Redis
make run         # собрать и запустить сервер
```

## Вклад

### Ветки
- `main` — стабильная, CI/CD автоматически деплоит
- `feat/*` — новые фичи
- `fix/*` — баг-фиксы

### Коммиты
Conventional commits:
```
feat: добавить эндпоинт GET /users/me
fix: исправить race condition в транзакциях
docs: обновить README
refactor: вынести логику в отдельный сервис
test: добавить тесты для auth handler
```

### Pull Request
1. Создай ветку от `main`
2. Напиши тесты
3. Убедись что `make test` и `make lint` проходят
4. Опиши изменения в PR

## Тесты

```bash
make test                    # все тесты
go test -race ./...          # с race detector
go test -coverprofile=c.out  # покрытие
go tool cover -html=c.out
```

## Стиль кода

- Go conventions (PascalCase/camelCase)
- Ошибки через `apperr.Wrap()`
- Логирование через `logger` package
- Деньги через `decimal.Decimal`, никогда `float64`
- DI через constructor

## Безопасность

Если нашёл уязвимость — см. [SECURITY.md](SECURITY.md).
