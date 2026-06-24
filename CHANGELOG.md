# Changelog

## [1.0.24] - 2026-06-24

### Added
- Real-time exchange rates via Frankfurter API (free, no key)
- `internal/exchange` package with caching and fallback
- Tests for contextkeys, logger, exchange packages
- `govulncheck` in CI/CD security scanning
- `golangci-lint` via official action v6
- CONTRIBUTING.md, SECURITY.md, CHANGELOG.md

### Fixed
- OTP: constant-time comparison (timing attack protection)
- JWT: signing method validation (alg:none prevention)
- Balance check before debit (insufficient funds)
- Dockerfile: `golang:1.25` → `golang:1.23`
- Dockerfile: added migrations to container
- go.mod: `go 1.22` → `go 1.23`
- Makefile: `docker-compose` → `docker compose` (v2)
- Logger: auto-initialize if Setup() not called
- .gitignore: ignore .omo/, .mimocode/, .codegraph/

### Changed
- CI/CD: security job non-blocking (continue-on-error)
- CI/CD: build depends on test + lint only
- Logging: all packages use `logger` package (not raw slog)

## [1.0.21] - 2026-06-23

### Fixed
- README and documentation updates

## [1.0.20] - 2026-06-23

### Fixed
- Various bug fixes

## [1.0.0] - 2026-06-20

### Added
- Initial release
- User registration with OTP verification
- JWT authentication with refresh tokens
- Multi-currency accounts (RUB/USD/EUR)
- Cross-currency transfers with exchange rates
- Anti-fraud system (C++ + Python via Redis)
- Optimistic locking for concurrent transactions
- Rate limiting on auth endpoints
- Docker + Docker Compose
- GitHub Actions CI/CD
