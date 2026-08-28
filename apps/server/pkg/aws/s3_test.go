package aws

import (
	"context"
	"path/filepath"
	"testing"
)

func TestNewS3ServiceUsesDefaultCredentialChainWhenStaticKeysAreAbsent(t *testing.T) {
	disableAmbientAWSCredentials(t)

	service, err := NewS3Service(Config{Region: "us-east-1"}, nil)
	if err != nil {
		t.Fatalf("NewS3Service with workload credentials: %v", err)
	}
	if service == nil || service.client == nil {
		t.Fatal("NewS3Service returned no S3 client")
	}
}

func TestNewS3ServiceRejectsPartialStaticCredentials(t *testing.T) {
	disableAmbientAWSCredentials(t)

	for name, config := range map[string]Config{
		"access key only": {Region: "us-east-1", AccessKeyID: "local-access-key"},
		"secret key only": {Region: "us-east-1", SecretAccessKey: "local-secret-key"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewS3Service(config, nil)
			if err == nil {
				t.Fatal("partial static credentials were accepted")
			}
		})
	}
}

func TestNewS3ServiceUsesCompleteStaticCredentialsForLocalCompatibility(t *testing.T) {
	disableAmbientAWSCredentials(t)

	service, err := NewS3Service(Config{
		Region:          "us-east-1",
		AccessKeyID:     "local-access-key",
		SecretAccessKey: "local-secret-key",
	}, nil)
	if err != nil {
		t.Fatalf("NewS3Service with static credentials: %v", err)
	}
	credentials, err := service.client.Options().Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("retrieve configured credentials: %v", err)
	}
	if credentials.AccessKeyID != "local-access-key" || credentials.SecretAccessKey != "local-secret-key" {
		t.Fatalf("unexpected configured credentials: source=%q access_key=%q", credentials.Source, credentials.AccessKeyID)
	}
}

func TestNewS3ServiceUsesServiceSpecificBaseEndpoint(t *testing.T) {
	disableAmbientAWSCredentials(t)

	const endpoint = "http://127.0.0.1:9000"
	service, err := NewS3Service(Config{
		Region:         "us-east-1",
		Endpoint:       endpoint,
		ForcePathStyle: true,
	}, nil)
	if err != nil {
		t.Fatalf("NewS3Service with custom endpoint: %v", err)
	}
	options := service.client.Options()
	if options.BaseEndpoint == nil || *options.BaseEndpoint != endpoint {
		t.Fatalf("custom S3 endpoint = %v, want %q", options.BaseEndpoint, endpoint)
	}
	if !options.UsePathStyle {
		t.Fatal("custom S3 client did not retain path-style addressing")
	}
}

func TestNewS3ServiceRequiresRegion(t *testing.T) {
	disableAmbientAWSCredentials(t)

	if _, err := NewS3Service(Config{}, nil); err == nil {
		t.Fatal("blank AWS region was accepted")
	}
}

func disableAmbientAWSCredentials(t *testing.T) {
	t.Helper()
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_SESSION_TOKEN", "")
	t.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", "")
	t.Setenv("AWS_ROLE_ARN", "")
	t.Setenv("AWS_CONTAINER_CREDENTIALS_FULL_URI", "")
	t.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	t.Setenv("AWS_CONFIG_FILE", filepath.Join(t.TempDir(), "config"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "credentials"))
}
