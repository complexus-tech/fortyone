package usersrepository

import (
	"fmt"
	"math/rand"
	"strings"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
)

// Repository errors
var (
	ErrNotFound = users.ErrNotFound
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

// generateOTP creates a 6-digit numeric OTP
func generateOTP() (string, error) {
	// Generate 6-digit numeric OTP (000000-999999)
	otp := rand.Intn(1000000)
	return fmt.Sprintf("%06d", otp), nil
}

// isUniqueConstraintViolation checks if the error is a unique constraint violation
func isUniqueConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate key value violates unique constraint") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "violates unique constraint")
}
