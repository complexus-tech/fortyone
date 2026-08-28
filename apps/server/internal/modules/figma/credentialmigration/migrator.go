// Package credentialmigration contains the only code allowed to open the
// retired provider-local Figma credential format.
package credentialmigration

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

const defaultBatchSize int32 = 100

type Store interface {
	ListLegacyCredentials(
		ctx context.Context,
		after *uuid.UUID,
		limit int32,
	) ([]figmadomain.LegacyCredential, error)
	UpgradeLegacyCredential(
		ctx context.Context,
		record figmadomain.LegacyCredential,
		nextPayload string,
	) (bool, error)
}

type Vault interface {
	Seal(binding credentialvault.Context, plaintext []byte) (string, error)
}

type Report struct {
	Scanned  int
	Migrated int
	Stale    int
}

type Migrator struct {
	store        Store
	vault        Vault
	legacySecret string
	batchSize    int32
}

func New(store Store, vault Vault, legacySecret string) (*Migrator, error) {
	if store == nil || vault == nil || strings.TrimSpace(legacySecret) == "" {
		return nil, errors.New("figma legacy credential migration is not configured")
	}
	return &Migrator{
		store: store, vault: vault, legacySecret: legacySecret,
		batchSize: defaultBatchSize,
	}, nil
}

func (migrator *Migrator) Run(ctx context.Context) (Report, error) {
	if migrator == nil || migrator.store == nil || migrator.vault == nil {
		return Report{}, errors.New("figma legacy credential migration is not configured")
	}
	var report Report
	var after *uuid.UUID
	for {
		records, err := migrator.store.ListLegacyCredentials(ctx, after, migrator.batchSize)
		if err != nil {
			return report, fmt.Errorf("list legacy Figma credentials: %w", err)
		}
		for _, record := range records {
			report.Scanned++
			plaintext, err := openLegacyCredential(migrator.legacySecret, record.Payload)
			if err != nil {
				return report, fmt.Errorf("open legacy Figma credential %s: %w", record.ID, err)
			}
			envelope, err := migrator.vault.Seal(
				figmaprovider.CredentialContext(
					record.WorkspaceID,
					record.ID,
					record.InstallationGeneration,
				),
				plaintext,
			)
			clear(plaintext)
			if err != nil {
				return report, fmt.Errorf("seal Figma credential %s: %w", record.ID, err)
			}
			replaced, err := migrator.store.UpgradeLegacyCredential(ctx, record, envelope)
			if err != nil {
				return report, fmt.Errorf("upgrade legacy Figma credential %s: %w", record.ID, err)
			}
			if replaced {
				report.Migrated++
			} else {
				report.Stale++
			}
		}
		if len(records) < int(migrator.batchSize) {
			return report, nil
		}
		cursor := records[len(records)-1].ID
		after = &cursor
	}
}

func openLegacyCredential(secret, encoded string) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, errors.New("legacy Figma credential envelope is malformed")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, errors.New("legacy Figma credential cipher is unavailable")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(raw) < gcm.NonceSize() {
		return nil, errors.New("legacy Figma credential envelope is malformed")
	}
	plaintext, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return nil, errors.New("legacy Figma credential authentication failed")
	}
	decoder := json.NewDecoder(bytes.NewReader(plaintext))
	decoder.DisallowUnknownFields()
	var token figmadomain.Token
	if err := decoder.Decode(&token); err != nil ||
		strings.TrimSpace(token.AccessToken) == "" || token.ExpiresAt.IsZero() {
		clear(plaintext)
		return nil, errors.New("legacy Figma credential payload is invalid")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		clear(plaintext)
		return nil, errors.New("legacy Figma credential payload is invalid")
	}
	return plaintext, nil
}
