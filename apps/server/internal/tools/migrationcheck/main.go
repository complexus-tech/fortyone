// Command migrationcheck validates the machine-readable post-baseline
// migration contract and its generated operator documentation.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	migrationcontract "github.com/complexus-tech/projects-api/internal/migrations/contract"
)

func main() {
	serverRoot := flag.String("root", ".", "path to the apps/server root")
	writeDocumentation := flag.Bool("write", false, "regenerate the migration operator guide")
	flag.Parse()
	if flag.NArg() != 0 {
		fatalf("migrationcheck does not accept positional arguments")
	}

	root, err := filepath.Abs(*serverRoot)
	if err != nil {
		fatalf("resolve server root: %v", err)
	}
	serverDirectory, err := os.OpenRoot(root)
	if err != nil {
		fatalf("open server root: %v", err)
	}
	defer serverDirectory.Close()

	manifest, err := migrationcontract.LoadManifest(serverDirectory.FS(), migrationcontract.ManifestFile)
	if err != nil {
		fatalf("%v", err)
	}
	if err := migrationcontract.ValidateManifest(manifest); err != nil {
		fatalf("migration manifest is invalid:\n%v", err)
	}
	if err := migrationcontract.ValidateMigrationDirectory(serverDirectory.FS(), migrationcontract.MigrationsDirectory, manifest); err != nil {
		fatalf("migration files violate the manifest:\n%v", err)
	}

	if *writeDocumentation {
		if err := writeFileAtomically(serverDirectory, migrationcontract.DocumentationFile, migrationcontract.RenderDocumentation(manifest)); err != nil {
			fatalf("write migration operator guide: %v", err)
		}
	} else if err := migrationcontract.ValidateDocumentation(serverDirectory.FS(), migrationcontract.DocumentationFile, manifest); err != nil {
		fatalf("%v", err)
	}

	head := manifest.BaselineVersion
	if len(manifest.Migrations) != 0 {
		head = manifest.Migrations[len(manifest.Migrations)-1].Version
	}
	fmt.Printf("migration contract valid for %d post-baseline migrations through %06d\n", len(manifest.Migrations), head)
}

func writeFileAtomically(root *os.Root, path string, contents []byte) (returnErr error) {
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return fmt.Errorf("generate temporary filename: %w", err)
	}
	temporaryPath := filepath.Join(filepath.Dir(path), ".migration-operations-"+hex.EncodeToString(random))
	temporary, err := root.OpenFile(temporaryPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if err := root.Remove(temporaryPath); err != nil && !os.IsNotExist(err) && returnErr == nil {
			returnErr = fmt.Errorf("remove temporary file: %w", err)
		}
	}()

	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set temporary file permissions: %w", err)
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}
	if err := root.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace operator guide: %w", err)
	}
	return nil
}

func fatalf(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "migrationcheck: "+format+"\n", arguments...)
	os.Exit(1)
}
