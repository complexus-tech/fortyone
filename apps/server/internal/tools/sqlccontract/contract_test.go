package sqlccontract

import (
	"context"
	"testing"
	"time"

	contractsql "github.com/complexus-tech/projects-api/internal/tools/sqlccontract/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	_ uuid.UUID                       = contractsql.SqlcTypeContract{}.ID
	_ *uuid.UUID                      = contractsql.SqlcTypeContract{}.NullableID
	_ time.Time                       = contractsql.SqlcTypeContract{}.OccurredAt
	_ *time.Time                      = contractsql.SqlcTypeContract{}.NullableOccurredAt
	_ time.Time                       = contractsql.SqlcTypeContract{}.LocalAt
	_ *time.Time                      = contractsql.SqlcTypeContract{}.NullableLocalAt
	_ time.Time                       = contractsql.SqlcTypeContract{}.DueDate
	_ *time.Time                      = contractsql.SqlcTypeContract{}.NullableDueDate
	_ contractsql.SqlcContractStatus  = contractsql.SqlcTypeContract{}.Status
	_ *contractsql.SqlcContractStatus = contractsql.SqlcTypeContract{}.NullableStatus
	_ pgtype.Numeric                  = contractsql.SqlcTypeContract{}.Amount
	_ []byte                          = contractsql.SqlcTypeContract{}.Payload
	_ []uuid.UUID                     = contractsql.SqlcTypeContract{}.RelatedIds
	_ uuid.UUID                       = contractsql.GetTypeContractParams{}.ID
)

func TestGeneratedManyQueryReturnsNonNilEmptySlice(t *testing.T) {
	contracts, err := contractsql.New(emptyDB{}).ListTypeContracts(
		context.Background(),
		contractsql.ListTypeContractsParams{},
	)
	if err != nil {
		t.Fatalf("list empty type contracts: %v", err)
	}
	if contracts == nil || len(contracts) != 0 {
		t.Fatalf("empty contracts = %#v, want non-nil empty slice", contracts)
	}
}

func TestGeneratedNullableEnumUsesPointer(t *testing.T) {
	var status *contractsql.SqlcContractStatus
	if status != nil {
		t.Fatal("zero nullable enum must be nil")
	}
}

type emptyDB struct{}

func (emptyDB) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	panic("unexpected Exec call")
}

func (emptyDB) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return emptyRows{}, nil
}

func (emptyDB) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	panic("unexpected QueryRow call")
}

type emptyRows struct{}

func (emptyRows) Close()                                       {}
func (emptyRows) Err() error                                   { return nil }
func (emptyRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (emptyRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (emptyRows) Next() bool                                   { return false }
func (emptyRows) Scan(...any) error                            { panic("unexpected Scan call") }
func (emptyRows) Values() ([]any, error)                       { panic("unexpected Values call") }
func (emptyRows) RawValues() [][]byte                          { return nil }
func (emptyRows) Conn() *pgx.Conn                              { return nil }
