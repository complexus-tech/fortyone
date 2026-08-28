package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrorClass identifies PostgreSQL failures whose handling is stable across
// modules. Domain repositories remain responsible for mapping a class to a
// domain-specific error.
type ErrorClass string

const (
	ErrorClassUnknown              ErrorClass = "unknown"
	ErrorClassUniqueViolation      ErrorClass = "unique_violation"
	ErrorClassForeignKeyViolation  ErrorClass = "foreign_key_violation"
	ErrorClassNotNullViolation     ErrorClass = "not_null_violation"
	ErrorClassCheckViolation       ErrorClass = "check_violation"
	ErrorClassSerializationFailure ErrorClass = "serialization_failure"
	ErrorClassDeadlock             ErrorClass = "deadlock"
)

const (
	sqlStateUniqueViolation      = "23505"
	sqlStateForeignKeyViolation  = "23503"
	sqlStateNotNullViolation     = "23502"
	sqlStateCheckViolation       = "23514"
	sqlStateSerializationFailure = "40001"
	sqlStateDeadlock             = "40P01"
)

// Classify returns a stable category for a PostgreSQL error. Wrapped errors are
// supported; non-PostgreSQL failures remain unknown.
func Classify(err error) ErrorClass {
	code, ok := SQLState(err)
	if !ok {
		return ErrorClassUnknown
	}

	switch code {
	case sqlStateUniqueViolation:
		return ErrorClassUniqueViolation
	case sqlStateForeignKeyViolation:
		return ErrorClassForeignKeyViolation
	case sqlStateNotNullViolation:
		return ErrorClassNotNullViolation
	case sqlStateCheckViolation:
		return ErrorClassCheckViolation
	case sqlStateSerializationFailure:
		return ErrorClassSerializationFailure
	case sqlStateDeadlock:
		return ErrorClassDeadlock
	default:
		return ErrorClassUnknown
	}
}

func SQLState(err error) (string, bool) {
	var postgresErr *pgconn.PgError
	if !errors.As(err, &postgresErr) || postgresErr.Code == "" {
		return "", false
	}
	return postgresErr.Code, true
}

func IsRetryableTransactionError(err error) bool {
	class := Classify(err)
	return class == ErrorClassSerializationFailure || class == ErrorClassDeadlock
}
