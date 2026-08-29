package architecture_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryHasNoPublicSelfHostDistributionBundle(t *testing.T) {
	t.Parallel()

	root := filepath.Clean(filepath.Join(serverDir(t), "..", ".."))
	for _, relative := range []string{
		"LICENSE",
		"CONTRIBUTING.md",
		"CODE_OF_CONDUCT.md",
		"fortyone.env.example",
		filepath.Join(".github", "FUNDING.yml"),
		"docker-compose.yml",
		"docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
		filepath.Join("deployments", "docker", "dockerfile.web"),
		filepath.Join("apps", "server", "LICENSE"),
		filepath.Join("apps", "server", "CONTRIBUTING.md"),
		filepath.Join("apps", "server", "docker-compose.yml"),
		filepath.Join("apps", "server", "docker-compose.yaml"),
	} {
		path := filepath.Join(root, relative)
		if _, err := os.Stat(path); err == nil {
			t.Errorf("public distribution artifact must remain absent: %s", relative)
		} else if !os.IsNotExist(err) {
			t.Errorf("inspect %s: %v", relative, err)
		}
	}

	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read repository README: %v", err)
	}
	content := strings.ToLower(string(readme))
	for _, required := range []string{"private monorepo", "managed fortyone service", "no public installation or distribution bundle"} {
		if !strings.Contains(content, required) {
			t.Errorf("repository README is missing private distribution contract %q", required)
		}
	}
	for _, forbidden := range []string{"modern, open-source", "full-stack, open-source", "docker compose up"} {
		if strings.Contains(content, forbidden) {
			t.Errorf("repository README reintroduced public distribution copy %q", forbidden)
		}
	}

	makefile, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatalf("read repository Makefile: %v", err)
	}
	for _, forbidden := range []string{"docker compose", "fortyoneapp/", "build-web:", "push-web:"} {
		if strings.Contains(strings.ToLower(string(makefile)), forbidden) {
			t.Errorf("repository Makefile reintroduced public distribution command %q", forbidden)
		}
	}

	structuredData, err := os.ReadFile(filepath.Join(root, "apps", "landing", "src", "components", "shared", "json-ld.tsx"))
	if err != nil {
		t.Fatalf("read landing structured data: %v", err)
	}
	if strings.Contains(strings.ToLower(string(structuredData)), "github.com/complexus-tech/fortyone") {
		t.Error("public landing structured data must not advertise the private source repository")
	}
}

func TestJavaScriptWorkspacesCannotBePublished(t *testing.T) {
	t.Parallel()

	root := filepath.Clean(filepath.Join(serverDir(t), "..", ".."))
	patterns := []string{
		filepath.Join(root, "package.json"),
		filepath.Join(root, "apps", "*", "package.json"),
		filepath.Join(root, "packages", "*", "package.json"),
	}

	var manifests []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("resolve package manifest pattern %s: %v", pattern, err)
		}
		manifests = append(manifests, matches...)
	}
	if len(manifests) == 0 {
		t.Fatal("no JavaScript package manifests found")
	}

	for _, manifest := range manifests {
		manifest := manifest
		t.Run(filepath.ToSlash(strings.TrimPrefix(manifest, root+string(filepath.Separator))), func(t *testing.T) {
			t.Parallel()

			content, err := os.ReadFile(manifest)
			if err != nil {
				t.Fatalf("read package manifest: %v", err)
			}
			var metadata struct {
				Private bool `json:"private"`
			}
			if err := json.Unmarshal(content, &metadata); err != nil {
				t.Fatalf("parse package manifest: %v", err)
			}
			if !metadata.Private {
				t.Error("internal JavaScript workspace must set private=true")
			}
		})
	}
}

func TestProductionImagesStayInPrivateECR(t *testing.T) {
	t.Parallel()

	root := filepath.Clean(filepath.Join(serverDir(t), "..", ".."))
	path := filepath.Join(root, ".github", "workflows", "ecs-fargate-release.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read production release workflow: %v", err)
	}
	workflow := string(content)
	for _, required := range []string{
		"amazon-ecr-login",
		"ECR_SERVER_REPOSITORY",
		"ECR_WORKER_REPOSITORY",
		"${{ github.sha }}",
		"provenance: mode=max",
		"sbom: true",
		"aquasec/trivy:0.74.0@sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969",
		"--severity HIGH,CRITICAL",
		"--exit-code 1",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("production release is missing private image contract %q", required)
		}
	}
	for _, forbidden := range []string{
		"docker/login-action",
		"DOCKERHUB_",
		"fortyoneapp/server",
		"fortyoneapp/worker",
		":latest",
	} {
		if strings.Contains(workflow, forbidden) {
			t.Errorf("production release contains public or mutable image contract %q", forbidden)
		}
	}
	if got := strings.Count(workflow, "imageDetails[0].imageDigest"); got != 3 {
		t.Errorf("production release resolves %d image digests, want one each for migration, worker, and API", got)
	}
	if got := strings.Count(workflow, "@$image_digest"); got != 3 {
		t.Errorf("production release pins %d task images by digest, want three", got)
	}
	for _, mutableDeployment := range []string{
		"image: ${{ needs.registry.outputs.url }}/${{ env.ECR_SERVER_REPOSITORY }}:${{ env.IMAGE_TAG }}",
		"image: ${{ needs.registry.outputs.url }}/${{ env.ECR_WORKER_REPOSITORY }}:${{ env.IMAGE_TAG }}",
	} {
		if strings.Contains(workflow, mutableDeployment) {
			t.Errorf("production release deploys a mutable image tag %q", mutableDeployment)
		}
	}
}

func TestProductionReleaseKeepsOptionalMigrationBeforeCredentialCutoverAndAPI(t *testing.T) {
	t.Parallel()

	root := filepath.Clean(filepath.Join(serverDir(t), "..", ".."))
	path := filepath.Join(root, ".github", "workflows", "ecs-fargate-release.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read production release workflow: %v", err)
	}
	workflow := string(content)

	migrationStart := strings.Index(workflow, "\n  migrate-database:\n")
	workerStart := strings.Index(workflow, "\n  deploy-worker:\n")
	serverStart := strings.Index(workflow, "\n  deploy-server:\n")
	if migrationStart < 0 || workerStart < 0 || serverStart < 0 {
		t.Fatalf(
			"release workflow must declare migrate-database, deploy-worker, and deploy-server jobs; indexes = %d/%d/%d",
			migrationStart,
			workerStart,
			serverStart,
		)
	}
	if !(migrationStart < workerStart && workerStart < serverStart) {
		t.Fatalf("release jobs are not ordered migration -> worker cutover -> API")
	}

	migrationJob := workflow[migrationStart:workerStart]
	for _, required := range []string{
		"vars.RUN_PRODUCTION_MIGRATIONS == 'true'",
		"vars.RUN_PRODUCTION_MIGRATIONS != 'true'",
		"aws ecs register-task-definition",
		"aws ecs run-task",
		`command: ["/app/api", "-migrate"]`,
		"aws ecs wait tasks-stopped",
		`if [ "$exit_code" != "0" ]`,
	} {
		if !strings.Contains(migrationJob, required) {
			t.Errorf("optional migration release job is missing contract %q", required)
		}
	}

	workerJob := workflow[workerStart:serverStart]
	if !strings.Contains(workerJob, "- migrate-database") {
		t.Error("worker deployment must depend on the successful one-shot migration")
	}
	serverJob := workflow[serverStart:]
	if !strings.Contains(serverJob, "- deploy-worker") {
		t.Error("API deployment must depend on the replacement worker credential cutover")
	}
}

func TestProductionImagesRunAsNonRootWithTLSRoots(t *testing.T) {
	t.Parallel()

	root := filepath.Clean(filepath.Join(serverDir(t), "..", ".."))
	for _, relative := range []string{
		filepath.Join("deployments", "docker", "dockerfile.server"),
		filepath.Join("deployments", "docker", "dockerfile.worker"),
	} {
		content, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatalf("read %s: %v", relative, err)
		}
		dockerfile := string(content)
		if got := strings.Count(dockerfile, "@sha256:"); got != 3 {
			t.Errorf("%s pins %d image digests, want the Dockerfile frontend, builder, and runtime images pinned", relative, got)
		}
		for _, required := range []string{
			"apk add --no-cache ca-certificates tzdata",
			"adduser -S -D -H -u 10001",
			"--chown=fortyone:fortyone",
			"org.opencontainers.image.revision=\"$BUILD_VERSION\"",
			"USER fortyone",
		} {
			if !strings.Contains(dockerfile, required) {
				t.Errorf("%s is missing runtime hardening contract %q", relative, required)
			}
		}
	}
}

func TestProductionBuildContextExcludesLocalToolsAndSecrets(t *testing.T) {
	t.Parallel()

	root := filepath.Clean(filepath.Join(serverDir(t), "..", ".."))
	content, err := os.ReadFile(filepath.Join(root, ".dockerignore"))
	if err != nil {
		t.Fatalf("read production Docker ignore contract: %v", err)
	}
	dockerignore := string(content)
	for _, required := range []string{
		"**/.git",
		"**/.env*",
		"**/.tools",
		"**/tmp",
		"**/node_modules",
	} {
		if !strings.Contains(dockerignore, required) {
			t.Errorf("production Docker context is missing exclusion %q", required)
		}
	}
}

func TestDeveloperToolsAreProjectLocalAndVersionPinned(t *testing.T) {
	t.Parallel()

	root := serverDir(t)
	makefile, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatalf("read server Makefile: %v", err)
	}
	contract := string(makefile)
	for _, required := range []string{
		"AIR_VERSION := v1.64.5",
		"MIGRATE_VERSION := v4.18.3",
		"AIR := $(CURDIR)/.tools/bin/air",
		"MIGRATE := $(CURDIR)/.tools/bin/migrate",
		"dev: air-bootstrap",
		"migrate-create: migrate-bootstrap",
		"migrate-up: migrate-bootstrap",
		"migrate-down: migrate-bootstrap",
		"migrate-version: migrate-bootstrap",
		"migrate-force: migrate-bootstrap",
	} {
		if !strings.Contains(contract, required) {
			t.Errorf("developer tool contract is missing %q", required)
		}
	}
	for _, forbidden := range []string{"~/go/bin", "$(HOME)/go/bin", "@latest"} {
		if strings.Contains(contract, forbidden) {
			t.Errorf("server Makefile reintroduced non-reproducible developer tool lookup %q", forbidden)
		}
	}

	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read server README: %v", err)
	}
	for _, forbidden := range []string{"~/go/bin", "@latest"} {
		if strings.Contains(string(readme), forbidden) {
			t.Errorf("server README reintroduced non-reproducible developer setup %q", forbidden)
		}
	}
}
