package integrationrequestsrepository

import (
	"testing"

	integrationrequestssql "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository/sqlc"
)

func TestIntegrationRequestMappingNormalizesNullMetadataToEmptyObject(t *testing.T) {
	t.Parallel()

	mapped, err := integrationRequestFromSQL(integrationrequestssql.IntegrationRequest{Metadata: []byte("null")})
	if err != nil {
		t.Fatalf("map integration request: %v", err)
	}
	if mapped.Metadata == nil || len(mapped.Metadata) != 0 {
		t.Fatalf("metadata = %#v, want initialized empty object", mapped.Metadata)
	}
}
