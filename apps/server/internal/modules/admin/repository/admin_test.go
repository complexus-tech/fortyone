package adminrepository

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuditLogJSONValuesScanNullAndDecode(t *testing.T) {
	row := dbAuditLog{}

	scanAuditJSONValue(t, &row.OldValue, nil)
	scanAuditJSONValue(t, &row.NewValue, []byte(`"Admin note added"`))

	auditLog := toAuditLog(row)

	require.Nil(t, auditLog.OldValue)
	require.Equal(t, "Admin note added", auditLog.NewValue)
}

func scanAuditJSONValue(t *testing.T, value any, src any) {
	t.Helper()

	scanner, ok := value.(sql.Scanner)
	require.True(t, ok)
	require.NoError(t, scanner.Scan(src))
}
