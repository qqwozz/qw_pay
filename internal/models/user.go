package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole — роль пользователя в системе.
type UserRole string

const (
	RoleUser  UserRole = "USER"  // Обычный пользователь
	RoleAdmin UserRole = "ADMIN" // Администратор (audit log, блокировка счетов)
)

// User — модель пользователя системы.
type User struct {
	ID           uuid.UUID `json:"id"`            // Уникальный идентификатор
	Email        string    `json:"email"`         // Email (уникальный)
	Phone        string    `json:"phone"`         // Телефон (уникальный)
	PasswordHash string    `json:"-"`             // Хеш пароля (не возвращается в JSON)
	Role         UserRole  `json:"role"`          // Роль пользователя
	IsVerified   bool      `json:"is_verified"`   // Подтверждён ли аккаунт через OTP
	CreatedAt    time.Time `json:"created_at"`    // Дата создания
	UpdatedAt    time.Time `json:"updated_at"`    // Дата последнего обновления
}
