# Security Policy

## Поддерживаемые версии

| Версия | Поддержка |
|--------|-----------|
| 1.0.x | Active |

## Сообщения об уязвимостях

**НЕ публикуй уязвимости в GitHub Issues.**

Отправь на: `security@qw-pay.dev` (или создай приватный issue).

### Что указать
- Описание уязвимости
- Steps to reproduce
- Потенциальный impact
- Предлагаемое исправление (если есть)

### Ответ
- Acknowledge в течение 48 часов
- Fix в течение 7 дней для critical
- Credit в RELEASE_NOTES

## Безопасность по дизайну

- Passwords: bcrypt (cost 10)
- JWT: HS256,短期 access + долгий refresh
- OTP: constant-time comparison
- SQL: параметризованные запросы
- Rate limiting на auth endpoints
- Anti-fraud проверки перед каждой транзакцией
