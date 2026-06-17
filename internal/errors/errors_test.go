package errors

import (
	"errors"
	"testing"
)

func TestAppError(t *testing.T) {
	t.Run("error message from message field", func(t *testing.T) {
		err := NotFound("not found resource")
		if err.Error() != "not found resource" {
			t.Errorf("expected 'not found resource', got '%s'", err.Error())
		}
	})

	t.Run("error message from wrapped error", func(t *testing.T) {
		err := &AppError{Err: ErrNotFound, Code: 404}
		if err.Error() != "resource not found" {
			t.Errorf("expected 'resource not found', got '%s'", err.Error())
		}
	})

	t.Run("unwrap returns inner error", func(t *testing.T) {
		inner := errors.New("inner")
		err := Wrap(inner, "outer")
		if !errors.Is(err, inner) {
			t.Error("expected errors.Is to find inner error")
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := NotFound("missing")
		if err.Code != 404 {
			t.Errorf("expected code 404, got %d", err.Code)
		}
		if !errors.Is(err, ErrNotFound) {
			t.Error("expected errors.Is to find ErrNotFound")
		}
	})

	t.Run("bad request", func(t *testing.T) {
		err := BadRequest("invalid input")
		if err.Code != 400 {
			t.Errorf("expected code 400, got %d", err.Code)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		err := Conflict("duplicate")
		if err.Code != 409 {
			t.Errorf("expected code 409, got %d", err.Code)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		err := Unauthorized("no token")
		if err.Code != 401 {
			t.Errorf("expected code 401, got %d", err.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		err := Forbidden("no access")
		if err.Code != 403 {
			t.Errorf("expected code 403, got %d", err.Code)
		}
	})

	t.Run("internal", func(t *testing.T) {
		err := Internal("server error")
		if err.Code != 500 {
			t.Errorf("expected code 500, got %d", err.Code)
		}
	})

	t.Run("custom", func(t *testing.T) {
		err := New(ErrNotFound, "custom message", 418)
		if err.Code != 418 {
			t.Errorf("expected code 418, got %d", err.Code)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", ErrNotFound},
		{"ErrAlreadyExists", ErrAlreadyExists},
		{"ErrUnauthorized", ErrUnauthorized},
		{"ErrForbidden", ErrForbidden},
		{"ErrBadRequest", ErrBadRequest},
		{"ErrInternal", ErrInternal},
		{"ErrOptimisticLock", ErrOptimisticLock},
		{"ErrAccountBlocked", ErrAccountBlocked},
		{"ErrSameAccount", ErrSameAccount},
		{"ErrAmountExceeded", ErrAmountExceeded},
		{"ErrDailyLimit", ErrDailyLimit},
		{"ErrNoExchangeRate", ErrNoExchangeRate},
		{"ErrOTPExpired", ErrOTPExpired},
		{"ErrOTPInvalid", ErrOTPInvalid},
		{"ErrAccountNotVerified", ErrAccountNotVerified},
	}

	for _, s := range sentinels {
		t.Run(s.name, func(t *testing.T) {
			if s.err == nil {
				t.Errorf("%s is nil", s.name)
			}
			if s.err.Error() == "" {
				t.Errorf("%s has empty error string", s.name)
			}
		})
	}
}
