package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestClassify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		code      string
		want      ErrorClass
		retryable bool
	}{
		{name: "unique", code: sqlStateUniqueViolation, want: ErrorClassUniqueViolation},
		{name: "foreign key", code: sqlStateForeignKeyViolation, want: ErrorClassForeignKeyViolation},
		{name: "not null", code: sqlStateNotNullViolation, want: ErrorClassNotNullViolation},
		{name: "check", code: sqlStateCheckViolation, want: ErrorClassCheckViolation},
		{name: "serialization", code: sqlStateSerializationFailure, want: ErrorClassSerializationFailure, retryable: true},
		{name: "deadlock", code: sqlStateDeadlock, want: ErrorClassDeadlock, retryable: true},
		{name: "unclassified PostgreSQL error", code: "22001", want: ErrorClassUnknown},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := fmt.Errorf("repository operation: %w", &pgconn.PgError{Code: test.code})

			if got := Classify(err); got != test.want {
				t.Fatalf("Classify() = %q, want %q", got, test.want)
			}
			if got := IsRetryableTransactionError(err); got != test.retryable {
				t.Fatalf("IsRetryableTransactionError() = %v, want %v", got, test.retryable)
			}
		})
	}
}

func TestSQLStateRejectsNonPostgreSQLError(t *testing.T) {
	t.Parallel()

	if code, ok := SQLState(errors.New("not PostgreSQL")); ok || code != "" {
		t.Fatalf("SQLState() = (%q, %v), want empty false", code, ok)
	}
	if got := Classify(nil); got != ErrorClassUnknown {
		t.Fatalf("Classify(nil) = %q, want unknown", got)
	}
}
