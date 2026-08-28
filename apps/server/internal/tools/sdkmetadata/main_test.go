package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadMetadataAndRenderersUseOnlyContractValues(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "openapi.json")
	contents := `{"openapi":"3.1.0","info":{"version":"1.2.3-preview.4"},"servers":[{"url":"https://api.example.test"}]}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		t.Fatalf("open fixture root: %v", err)
	}
	t.Cleanup(func() { _ = root.Close() })
	contract, err := readMetadata(root, filepath.Base(path))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	for name, output := range map[string]string{
		"go":         string(renderGo(contract)),
		"typescript": string(renderTypeScript(contract)),
	} {
		for _, value := range []string{"1.2.3-preview.4", "https://api.example.test", "3.1.0"} {
			if !strings.Contains(output, value) {
				t.Fatalf("%s output does not contain %q: %s", name, value, output)
			}
		}
	}
}

func TestReadMetadataRejectsUnsupportedOrUnsafeContract(t *testing.T) {
	t.Parallel()
	for name, contents := range map[string]string{
		"unsupported OpenAPI": `{"openapi":"3.0.3","info":{"version":"1"},"servers":[{"url":"https://api.example.test"}]}`,
		"missing version":     `{"openapi":"3.1.0","info":{},"servers":[{"url":"https://api.example.test"}]}`,
		"insecure server":     `{"openapi":"3.1.0","info":{"version":"1"},"servers":[{"url":"http://api.example.test"}]}`,
		"credentialed server": `{"openapi":"3.1.0","info":{"version":"1"},"servers":[{"url":"https://user@api.example.test"}]}`,
		"missing server host": `{"openapi":"3.1.0","info":{"version":"1"},"servers":[{"url":"https://"}]}`,
	} {
		name, contents := name, contents
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "openapi.json")
			if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			root, err := os.OpenRoot(filepath.Dir(path))
			if err != nil {
				t.Fatalf("open fixture root: %v", err)
			}
			t.Cleanup(func() { _ = root.Close() })
			if _, err := readMetadata(root, filepath.Base(path)); err == nil {
				t.Fatal("readMetadata succeeded")
			}
		})
	}
}

func TestValidateTypeScriptPackage(t *testing.T) {
	t.Parallel()
	for name, testCase := range map[string]struct {
		contents string
		wantErr  bool
	}{
		"matching private preview": {contents: `{"version":"1.2.3-preview.4","private":true}`},
		"wrong version":            {contents: `{"version":"1.2.3","private":true}`, wantErr: true},
		"publishable preview":      {contents: `{"version":"1.2.3-preview.4","private":false}`, wantErr: true},
	} {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "package.json")
			if err := os.WriteFile(path, []byte(testCase.contents), 0o600); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			root, err := os.OpenRoot(filepath.Dir(path))
			if err != nil {
				t.Fatalf("open fixture root: %v", err)
			}
			t.Cleanup(func() { _ = root.Close() })
			err = validateTypeScriptPackage(root, filepath.Base(path), "1.2.3-preview.4")
			if (err != nil) != testCase.wantErr {
				t.Fatalf("validateTypeScriptPackage() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

func TestSDKMetadataFileAccessCannotEscapeConfiguredRoots(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	rootPath := filepath.Join(parent, "generated")
	if err := os.Mkdir(rootPath, 0o750); err != nil {
		t.Fatalf("create generated root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(parent, "outside.json"), []byte(`{"openapi":"3.1.0"}`), 0o600); err != nil {
		t.Fatalf("write outside fixture: %v", err)
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		t.Fatalf("open generated root: %v", err)
	}
	t.Cleanup(func() { _ = root.Close() })

	if _, err := readMetadata(root, "../outside.json"); err == nil {
		t.Fatal("readMetadata escaped its configured root")
	}
	if err := writeFile(root, "../outside.go", []byte("package outside\n")); err == nil {
		t.Fatal("writeFile escaped its configured root")
	}
	if _, err := os.Stat(filepath.Join(parent, "outside.go")); !os.IsNotExist(err) {
		t.Fatalf("outside output exists or stat failed unexpectedly: %v", err)
	}
}
