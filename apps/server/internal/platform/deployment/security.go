package deployment

import (
	"errors"
	"fmt"
	"strings"
)

const MinimumProductionSecretBytes = 32

// SecretRequirement describes a secret that must be safe before a production
// process starts. Validation errors identify configuration keys but never
// include secret values.
type SecretRequirement struct {
	Name            string
	Value           string
	MinimumBytes    int
	ForbiddenValues []string
}

// ValidateProductionSecrets rejects missing, known-default, and undersized
// secrets in production while leaving local and test setup lightweight.
func ValidateProductionSecrets(mode Mode, requirements ...SecretRequirement) error {
	if !mode.IsProduction() {
		return nil
	}

	var validationErrors []error
	for _, requirement := range requirements {
		name := strings.TrimSpace(requirement.Name)
		if name == "" {
			validationErrors = append(validationErrors, errors.New("production secret requirement has no configuration key"))
			continue
		}

		value := strings.TrimSpace(requirement.Value)
		if value == "" {
			validationErrors = append(validationErrors, fmt.Errorf("%s is required in production", name))
			continue
		}

		for _, forbidden := range requirement.ForbiddenValues {
			if value == strings.TrimSpace(forbidden) {
				validationErrors = append(validationErrors, fmt.Errorf("%s must not use a known development default in production", name))
				break
			}
		}

		minimumBytes := requirement.MinimumBytes
		if minimumBytes <= 0 {
			minimumBytes = MinimumProductionSecretBytes
		}
		if len([]byte(value)) < minimumBytes {
			validationErrors = append(validationErrors, fmt.Errorf("%s must contain at least %d bytes in production", name, minimumBytes))
		}
	}

	return errors.Join(validationErrors...)
}

// TransportSecurity contains the resolved transport settings shared by the
// API and worker. The PostgreSQL mode must already be resolved through the
// database package so deprecated compatibility flags cannot bypass policy.
type TransportSecurity struct {
	PostgreSQLSSLMode string
	RedisTLSDisabled  bool
}

// ValidateProductionTransports fails closed when production connections do
// not authenticate and encrypt PostgreSQL and Redis traffic.
func ValidateProductionTransports(mode Mode, security TransportSecurity) error {
	if !mode.IsProduction() {
		return nil
	}

	var validationErrors []error
	if strings.TrimSpace(security.PostgreSQLSSLMode) != "verify-full" {
		validationErrors = append(validationErrors, errors.New("APP_DB_SSL_MODE must be verify-full in production"))
	}
	if security.RedisTLSDisabled {
		validationErrors = append(validationErrors, errors.New("APP_REDIS_DISABLE_TLS must be false in production"))
	}
	return errors.Join(validationErrors...)
}

// ValidateAWSCredentialSource keeps local S3-compatible development possible
// while requiring production to use the AWS workload identity chain. Static
// key fields must always be paired so a typo cannot silently select a different
// credential source.
func ValidateAWSCredentialSource(mode Mode, accessKeyID, secretAccessKey string) error {
	hasAccessKey := strings.TrimSpace(accessKeyID) != ""
	hasSecretKey := strings.TrimSpace(secretAccessKey) != ""
	if hasAccessKey != hasSecretKey {
		return errors.New("APP_AWS_ACCESS_KEY_ID and APP_AWS_SECRET_ACCESS_KEY must be configured together")
	}
	if mode.IsProduction() && hasAccessKey {
		return errors.New("production AWS storage must use the ECS task role; static APP_AWS credentials are forbidden")
	}
	return nil
}
