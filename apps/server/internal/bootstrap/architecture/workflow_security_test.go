package architecture_test

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var immutableActionReference = regexp.MustCompile(`^[^@[:space:]]+@[0-9a-f]{40}$`)

func TestThirdPartyWorkflowActionsUseImmutableCommits(t *testing.T) {
	t.Parallel()

	workflowsRoot := filepath.Join(serverDir(t), "..", "..", ".github", "workflows")
	err := filepath.WalkDir(workflowsRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || (filepath.Ext(path) != ".yml" && filepath.Ext(path) != ".yaml") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for lineNumber := 1; scanner.Scan(); lineNumber++ {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "uses:") {
				continue
			}
			reference := strings.TrimSpace(strings.TrimPrefix(line, "uses:"))
			if comment := strings.Index(reference, " #"); comment >= 0 {
				reference = strings.TrimSpace(reference[:comment])
			}
			reference = strings.Trim(reference, `"'`)
			if strings.HasPrefix(reference, "./") {
				continue
			}
			if !immutableActionReference.MatchString(reference) {
				t.Errorf("%s:%d uses mutable third-party action %q; pin the reviewed 40-character commit and retain the release version in a comment", path, lineNumber, reference)
			}
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatalf("scan workflow actions: %v", err)
	}
}

func TestServerReleaseUsesOIDCInsteadOfLongLivedAWSKeys(t *testing.T) {
	t.Parallel()

	path := filepath.Join(serverDir(t), "..", "..", ".github", "workflows", "ecs-fargate-release.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read server release workflow: %v", err)
	}
	workflow := string(content)
	for _, forbidden := range []string{
		"aws-access-key-id:",
		"aws-secret-access-key:",
		"secrets.AWS_ACCESS_KEY_ID",
		"secrets.AWS_SECRET_ACCESS_KEY",
	} {
		if strings.Contains(workflow, forbidden) {
			t.Errorf("release workflow contains long-lived AWS credential input %q", forbidden)
		}
	}
	for _, required := range []string{"id-token: write", "role-to-assume: ${{ vars.AWS_DEPLOY_ROLE_ARN }}"} {
		if !strings.Contains(workflow, required) {
			t.Errorf("release workflow is missing OIDC contract %q", required)
		}
	}
}

func TestServerReleaseValidatesRequiredConfigurationBeforeAWS(t *testing.T) {
	t.Parallel()

	path := filepath.Join(serverDir(t), "..", "..", ".github", "workflows", "ecs-fargate-release.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read server release workflow: %v", err)
	}
	workflow := string(content)

	preflightStart := strings.Index(workflow, "\n  release-config:\n")
	qualityStart := strings.Index(workflow, "\n  server-quality:\n")
	if preflightStart < 0 || qualityStart < 0 || preflightStart >= qualityStart {
		t.Fatal("release workflow must validate production configuration before declaring release jobs")
	}

	preflight := workflow[preflightStart:qualityStart]
	for _, variable := range []string{
		"AWS_DEPLOY_ROLE_ARN",
		"AWS_REGION",
		"ECR_SERVER_REPOSITORY",
		"ECR_WORKER_REPOSITORY",
		"ECS_CLUSTER",
		"ECS_SERVER_SERVICE",
		"ECS_SERVER_TASK_DEFINITION",
		"ECS_SERVER_CONTAINER",
		"ECS_WORKER_SERVICE",
		"ECS_WORKER_TASK_DEFINITION",
		"ECS_WORKER_CONTAINER",
	} {
		if !strings.Contains(preflight, variable) {
			t.Errorf("release configuration preflight does not validate %s", variable)
		}
	}

	registryStart := strings.Index(workflow, "\n  registry:\n")
	buildStart := strings.Index(workflow, "\n  build-and-push:\n")
	if registryStart < 0 || buildStart < 0 || registryStart >= buildStart {
		t.Fatal("release workflow must declare registry before build-and-push")
	}
	registry := workflow[registryStart:buildStart]
	if !strings.Contains(registry, "needs: release-config") {
		t.Error("registry authentication must wait for release configuration preflight")
	}
}
