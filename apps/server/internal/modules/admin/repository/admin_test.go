package adminrepository

import (
	"testing"

	adminsql "github.com/complexus-tech/projects-api/internal/modules/admin/repository/sqlc"
	"github.com/stretchr/testify/require"
)

func TestAuditLogJSONValuesDecodeNullAndString(t *testing.T) {
	row := adminsql.ListAdminAuditLogsRow{
		OldValue: []byte("null"), NewValue: []byte(`"Admin note added"`),
		Metadata: []byte(`{"source":"test"}`),
	}
	auditLog, err := auditLogFromRow(row)

	require.NoError(t, err)
	require.Nil(t, auditLog.OldValue)
	require.Equal(t, "Admin note added", auditLog.NewValue)
	require.Equal(t, map[string]any{"source": "test"}, auditLog.Metadata)
}

func TestDecodeJSONRejectsCorruptAuditData(t *testing.T) {
	_, err := decodeJSON([]byte(`{"broken"`))
	require.Error(t, err)
}

func TestDecodeJSONTreatsSQLNullAsNil(t *testing.T) {
	value, err := decodeJSON(nil)
	require.NoError(t, err)
	require.Nil(t, value)
}
