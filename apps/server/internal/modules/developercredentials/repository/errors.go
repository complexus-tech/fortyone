package developercredentialsrepository

import (
	"errors"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return developercredentialsdomain.ErrCredentialNotFound
	}
	if platformdatabase.IsRetryableTransactionError(err) {
		return developercredentialsdomain.ErrConcurrentUpdate
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return err
	}
	switch postgresError.ConstraintName {
	case "api_credentials_lookup_prefix_key":
		return developercredentialsdomain.ErrTokenPrefixCollision
	case "api_credentials_single_rotation_key":
		return developercredentialsdomain.ErrCredentialRotationConflict
	case "api_credential_scopes_scope_check":
		return developercredentialsdomain.ErrInvalidScope
	case "api_credential_scopes_service_account_management_check":
		return developercredentialsdomain.ErrInvalidScope
	case "api_credentials_principal_kind_check":
		return developercredentialsdomain.ErrInvalidCredentialKind
	case "api_credential_team_restrictions_team_workspace_fkey":
		return developercredentialsdomain.ErrTeamRestrictionNotAllowed
	case "principals_service_account_role_check", "principals_machine_role_check", "principals_identity_shape_check":
		return developercredentialsdomain.ErrInvalidServiceAccountRole
	default:
		return err
	}
}
