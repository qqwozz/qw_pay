# API Platform Builder

**ID:** api-platform
**Version:** 2.0
**Category:** API Design & Documentation
**Triggers:** API design, REST, OpenAPI, Swagger, endpoint creation, versioning, pagination, documentation

---

## Role

I am a senior API architect specializing in RESTful payment APIs. I design, document, and implement APIs following industry best practices.

---

## API Design Principles

### RESTful Resource Naming
```
POST   /api/v1/users              # Create user
GET    /api/v1/users/:id          # Get user
PUT    /api/v1/users/:id          # Update user
DELETE /api/v1/users/:id          # Delete user
```

### HTTP Methods & Status Codes

| Method | Purpose | Success Code | Error Codes |
|--------|---------|--------------|-------------|
| GET | Read resource | 200 OK | 404 Not Found |
| POST | Create resource | 201 Created | 400, 409 |
| PUT | Replace resource | 200 OK | 400, 404 |
| PATCH | Partial update | 200 OK | 400, 404 |
| DELETE | Remove resource | 204 No Content | 404 |

### Response Envelope
```go
type SuccessResponse struct {
    Data    interface{} `json:"data"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorResponse struct {
    Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type Meta struct {
    Total   int `json:"total"`
    Page    int `json:"page"`
    PerPage int `json:"per_page"`
    Pages   int `json:"pages"`
}
```

---

## QW Pay API Specification

### Base URL
```
http://localhost:8080/api/v1
```

### Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/register` | Registration (email + OTP) |
| `POST` | `/verify` | OTP verification |
| `POST` | `/login` | Login → access + refresh tokens |
| `POST` | `/refresh` | Token refresh |

### Protected Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/logout` | Logout (revoke refresh token) |
| `POST` | `/accounts` | Create account (+100 bonus) |
| `GET` | `/accounts` | List accounts |
| `POST` | `/accounts/:id/block` | Block account |
| `POST` | `/transactions` | Create transfer |
| `GET` | `/transactions?page=1&page_size=20` | Transaction history |
| `POST` | `/antifraud/block-user` | Block user in anti-fraud |
| `POST` | `/antifraud/block-account` | Block account in anti-fraud |

### Exchange Rates

| From | To | Rate |
|------|-----|------|
| RUB | USD | 0.011 |
| USD | RUB | 90.91 |
| RUB | EUR | 0.010 |
| EUR | RUB | 100.00 |
| USD | EUR | 0.92 |
| EUR | USD | 1.09 |

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_DATA` | 400 | Request body validation failed |
| `INSUFFICIENT_FUNDS` | 400 | Not enough balance |
| `LIMIT_EXCEEDED` | 400 | Transfer limit exceeded |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource doesn't exist |
| `CONFLICT` | 409 | Resource already exists |
| `RATE_LIMITED` | 429 | Too many requests |
| `ANTI_FRAUD_BLOCKED` | 422 | Blocked by anti-fraud |

---

## Implementation Checklist

- [ ] Define request/response schemas
- [ ] Choose correct HTTP method and status codes
- [ ] Add authentication middleware if protected
- [ ] Implement input validation
- [ ] Use `internal/response` for consistent responses
- [ ] Use `internal/errors` for error handling
- [ ] Add rate limiting if needed
- [ ] Document in `docs/API.md`
- [ ] Update `README.md` endpoint table
- [ ] Write unit tests
- [ ] Add integration tests
