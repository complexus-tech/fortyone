package deployment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateProductionSecrets(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateProductionSecrets(Development, SecretRequirement{
		Name:  "APP_AUTH_SECRET_KEY",
		Value: "secret",
	}))

	err := ValidateProductionSecrets(Production,
		SecretRequirement{
			Name:            "APP_AUTH_SECRET_KEY",
			Value:           "secret",
			ForbiddenValues: []string{"secret"},
		},
		SecretRequirement{
			Name:  "FEEDBACK_INGRESS_SECRET",
			Value: "",
		},
	)
	require.ErrorContains(t, err, "APP_AUTH_SECRET_KEY must not use a known development default")
	require.ErrorContains(t, err, "APP_AUTH_SECRET_KEY must contain at least 32 bytes")
	require.ErrorContains(t, err, "FEEDBACK_INGRESS_SECRET is required")
	require.NotContains(t, err.Error(), "secret\n")
}

func TestValidateProductionSecretsAcceptsStrongValues(t *testing.T) {
	t.Parallel()

	err := ValidateProductionSecrets(Production, SecretRequirement{
		Name:            "APP_AUTH_SECRET_KEY",
		Value:           "a-unique-production-secret-with-32-bytes",
		ForbiddenValues: []string{"secret"},
	})
	require.NoError(t, err)
}

func TestValidateProductionTransports(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateProductionTransports(Development, TransportSecurity{
		PostgreSQLSSLMode: "disable",
		RedisTLSDisabled:  true,
	}))

	err := ValidateProductionTransports(Production, TransportSecurity{
		PostgreSQLSSLMode: "require",
		RedisTLSDisabled:  true,
	})
	require.ErrorContains(t, err, "APP_DB_SSL_MODE must be verify-full")
	require.ErrorContains(t, err, "APP_REDIS_DISABLE_TLS must be false")

	require.NoError(t, ValidateProductionTransports(Production, TransportSecurity{
		PostgreSQLSSLMode: "verify-full",
	}))
}

func TestValidateAWSCredentialSource(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateAWSCredentialSource(Development, "", ""))
	require.NoError(t, ValidateAWSCredentialSource(Development, "local-access", "local-secret"))
	require.NoError(t, ValidateAWSCredentialSource(Production, "", ""))
	require.ErrorContains(t, ValidateAWSCredentialSource(Development, "local-access", ""), "must be configured together")
	require.ErrorContains(t, ValidateAWSCredentialSource(Production, "production-access", "production-secret"), "ECS task role")
}
