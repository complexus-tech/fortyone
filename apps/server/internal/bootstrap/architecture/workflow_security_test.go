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

func TestServerReleaseUsesExistingAWSDeploymentSecrets(t *testing.T) {
	t.Parallel()

	path := filepath.Join(serverDir(t), "..", "..", ".github", "workflows", "ecs-fargate-release.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read server release workflow: %v", err)
	}
	workflow := string(content)
	for _, required := range []string{
		"aws-access-key-id:",
		"aws-secret-access-key:",
		"secrets.AWS_ACCESS_KEY_ID",
		"secrets.AWS_SECRET_ACCESS_KEY",
		"vars.AWS_REGION",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("release workflow is missing existing AWS deployment input %q", required)
		}
	}
	for _, unsupported := range []string{
		"role-to-assume:",
		"vars.AWS_DEPLOY_ROLE_ARN",
		"id-token: write",
	} {
		if strings.Contains(workflow, unsupported) {
			t.Errorf("release workflow contains unconfigured AWS deployment input %q", unsupported)
		}
	}
}
