package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulePrefix = "github.com/complexus-tech/projects-api/internal/"

func TestNoLegacyPackageNames(t *testing.T) {
	internalRoot := internalDir(t)
	fset := token.NewFileSet()

	err := filepath.WalkDir(internalRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly)
		if parseErr != nil {
			t.Fatalf("parse package for %s: %v", path, parseErr)
		}

		pkg := file.Name.Name
		if strings.HasSuffix(pkg, "grp") || strings.HasSuffix(pkg, "repo") {
			t.Errorf("legacy package name %q in %s", pkg, path)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk internal dir: %v", err)
	}
}

func TestServiceAndRepositoryDoNotImportHTTP(t *testing.T) {
	internalRoot := internalDir(t)
	fset := token.NewFileSet()

	err := filepath.WalkDir(internalRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		cleanPath := filepath.ToSlash(path)
		if !strings.Contains(cleanPath, "/service/") && !strings.Contains(cleanPath, "/repository/") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			t.Fatalf("parse imports for %s: %v", path, parseErr)
		}

		for _, imp := range file.Imports {
			importPath, unquoteErr := strconv.Unquote(imp.Path.Value)
			if unquoteErr != nil {
				t.Fatalf("unquote import in %s: %v", path, unquoteErr)
			}

			if !strings.HasPrefix(importPath, modulePrefix) {
				continue
			}

			if strings.Contains(importPath, "/http/") || strings.HasSuffix(importPath, "/http") || strings.Contains(importPath, "/platform/http/") {
				t.Errorf("layer violation: %s imports %s", path, importPath)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk internal dir: %v", err)
	}
}

func internalDir(t *testing.T) string {
	t.Helper()

	root := filepath.Clean(filepath.Join("..", "..", ".."))
	return filepath.Join(root, "internal")
}
