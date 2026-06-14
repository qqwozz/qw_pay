# API Reference

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

For protected endpoints, pass the JWT token in the `Authorization` header:

```
Authorization: Bearer <token>
```

---

## Public Endpoints

### POST /register

Register a new user.

**Request:**
```json
{
  "email": "user@example.com",
  "phone": "+79001234567",
  "password": "securepass123"
}
```

**Response (200):**
```json
{
  "message": "Registration successful. Check logs for OTP code.",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Errors:**
- `400` — Invalid data or email already taken

---

### POST /verify

Verify account with OTP code.

**Request:**
```json
{
  "email": "user@example.com",
  "otp_code": "123456"
}
```

**Response (200):**
```json
{
  "message": "Account verified"
}
```

**Errors:**
- `400` — Invalid or expired OTP

---

### POST /login

Login and receive a JWT token.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepass123"
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "bearer",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "role": "USER"
}
```

**Errors:**
- `401` — Invalid credentials
- `403` — Account not verified

---

## Currency Endpoints

### GET /currencies

List supported currencies.

**Response (200):**
```json
{
  "currencies": ["RUB", "USD", "EUR"]
}
```

### GET /currencies/rates

Get all exchange rates.

**Response (200):**
```json
{
  "rates": [
    {"from_currency": "RUB", "to_currency": "USD", "rate": 0.011, "source": "initial"},
    {"from_currency": "USD", "to_currency": "RUB", "rate": 90.91, "source": "initial"}
  ]
}
```

### GET /currencies/rate?from=RUB&to=USD

Get a specific exchange rate.

**Response (200):**
```json
{
  "from_currency": "RUB",
  "to_currency": "USD",
  "rate": 0.011,
  "source": "initial"
}
```

### POST /currencies/convert

Convert amount between currencies.

**Request:**
```json
{
  "amount": 1000,
  "from": "RUB",
  "to": "USD"
}
```

**Response (200):**
```json
{
  "original_amount": 1000,
  "original_currency": "RUB",
  "converted_amount": 11,
  "target_currency": "USD",
  "exchange_rate": 0.011
}
```

### POST /currencies/rate *(Authenticated)*

Update an exchange rate.

**Request:**
```json
{
  "from": "RUB",
  "to": "USD",
  "rate": 0.012
}
```

**Response (200):**
```json
{
  "message": "Rate updated"
}
```

---

## Authenticated Endpoints

### POST /accounts

Create a new account. Welcome bonus: 100 units of the chosen currency.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "currency": "USD"
}
```

**Response (200):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "currency": "USD",
  "balance": 100.0,
  "version": 1,
  "status": "ACTIVE",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

### GET /accounts

List user's accounts.

**Headers:** `Authorization: Bearer <token>`

**Response (200):**
```json
[
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "currency": "USD",
    "balance": 100.0,
    "version": 1,
    "status": "ACTIVE",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

---

### POST /accounts/:id/block

Block own account.

**Headers:** `Authorization: Bearer <token>`

**Response (200):**
```json
{
  "message": "Account blocked"
}
```

**Errors:**
- `403` — Account doesn't belong to you
- `404` — Account not found

---

### POST /transactions

Create a transfer between accounts.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "from_account_id": "660e8400-e29b-41d4-a716-446655440001",
  "to_account_id": "660e8400-e29b-41d4-a716-446655440002",
  "amount": 50.00,
  "source_currency": "USD",
  "idempotency_key": "unique-key-123"
}
```

**Response (200):**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440003",
  "idempotency_key": "unique-key-123",
  "from_account_id": "660e8400-e29b-41d4-a716-446655440001",
  "to_account_id": "660e8400-e29b-41d4-a716-446655440002",
  "amount": 50.0,
  "currency": "EUR",
  "source_currency": "USD",
  "exchange_rate_used": 0.92,
  "converted_amount": 46.0,
  "status": "EXECUTED",
  "created_at": "2024-01-15T10:35:00Z"
}
```

**Errors:**
- `400` — Insufficient funds, limit exceeded, invalid data
- `403` — Account blocked / anti-fraud rejection

---

### GET /transactions

Transaction history with pagination.

**Headers:** `Authorization: Bearer <token>`

**Query params:** `page` (default: 1), `page_size` (default: 20, max: 100)

**Response (200):**
```json
{
  "transactions": [...],
  "total": 15,
  "page": 1,
  "page_size": 20
}
```

---

### POST /antifraud/block-user

Block a user in the anti-fraud system.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### POST /antifraud/block-account

Block an account in the anti-fraud system.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "account_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

---

## Admin Endpoints

All admin endpoints require `role=ADMIN` in the JWT token.

### GET /admin/audit-logs

View audit log entries.

**Headers:** `Authorization: Bearer <token>`

**Query params:**
- `page` (default: 1)
- `page_size` (default: 20, max: 100)
- `user_id` (optional, UUID filter)
- `action` (optional, filter by action type)

**Available actions:** `TRANSFER_COMPLETED`, `TRANSFER_BLOCKED_BY_FRAUD`, `ACCOUNT_CREATED`, `ACCOUNT_BLOCKED`, `USER_REGISTERED`, `USER_VERIFIED`

**Response (200):**
```json
{
  "logs": [
    {
      "id": "...",
      "user_id": "...",
      "action": "TRANSFER_COMPLETED",
      "entity_type": "transaction",
      "entity_id": "...",
      "ip_address": "127.0.0.1",
      "created_at": "2024-01-15T10:35:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "page_size": 20
}
```

---

### POST /admin/accounts/block

Admin block any account (bypasses ownership check).

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "account_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Response (200):**
```json
{
  "message": "Account blocked by admin"
}
```

---

## Exchange Rates

| From | To | Rate |
|------|----|------|
| RUB | USD | 0.011 |
| USD | RUB | 90.91 |
| RUB | EUR | 0.010 |
| EUR | RUB | 100.0 |
| USD | EUR | 0.92 |
| EUR | USD | 1.09 |

---

## Error Format

All errors return:

```json
{
  "error": "Description of the error"
}
```

| Code | Description |
|------|-------------|
| `400` | Bad request / validation error |
| `401` | Not authenticated |
| `403` | Forbidden (ownership / admin required / anti-fraud) |
| `404` | Resource not found |
| `500` | Internal server error |
