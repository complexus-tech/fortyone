package developeroauthrepository

import (
	"errors"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapApplicationDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	if platformdatabase.IsRetryableTransactionError(err) {
		return developeroauthdomain.ErrConcurrentUpdate
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return err
	}
	switch postgresError.ConstraintName {
	case "oauth_client_secrets_lookup_prefix_key":
		return developeroauthdomain.ErrSecretPrefixCollision
	case "oauth_client_secrets_single_rotation_key",
		"oauth_application_installations_active_identity_key",
		"oauth_application_installations_principal_id_key",
		"oauth_applications_client_id_key":
		return developeroauthdomain.ErrConcurrentUpdate
	case "oauth_application_installation_scopes_scope_check":
		return developeroauthdomain.ErrInvalidScope
	case "oauth_application_installations_principal_kind_check",
		"principals_identity_shape_check",
		"principals_machine_role_check":
		return developeroauthdomain.ErrAccessDenied
	default:
		return err
	}
}
