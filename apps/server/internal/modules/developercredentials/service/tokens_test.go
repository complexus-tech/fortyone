package developercredentials

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTokenIssueAndVerificationBindEverySecurityField(t *testing.T) {
	t.Parallel()
	ref := developercredentialsdomain.DigestKeyRef{ID: "2026-08", Version: 7}
	manager, err := newTokenManager(TokenKeyringConfig{
		Active: ref,
		Keys:   []DigestKey{{Ref: ref, Material: bytes.Repeat([]byte{0x42}, digestKeyBytes)}},
	}, bytes.NewReader(bytes.Repeat([]byte{0x7a}, lookupPrefixSize+secretSize)))
	require.NoError(t, err)
	credentialID := uuid.New()
	issued, err := manager.issue(developercredentialsdomain.CredentialPersonalAccessToken, credentialID)
	require.NoError(t, err)
	raw := issued.Plaintext.Reveal()
	require.True(t, strings.HasPrefix(raw, personalHeader))
	require.Len(t, issued.Material.LookupPrefix, encodedPrefixLen)
	require.Len(t, issued.Material.SecretDigest, 32)

	record := developercredentialsdomain.VerificationRecord{
		CredentialID: credentialID, CredentialKind: issued.Material.Kind,
		LookupPrefix: issued.Material.LookupPrefix, SecretDigest: issued.Material.SecretDigest,
		TokenVersion: issued.Material.TokenVersion, DigestKey: issued.Material.DigestKey,
	}
	require.NoError(t, manager.verify(raw, record))

	tampered := record
	tampered.CredentialID = uuid.New()
	require.ErrorIs(t, manager.verify(raw, tampered), developercredentialsdomain.ErrAuthenticationFailed)
	tampered = record
	tampered.CredentialKind = developercredentialsdomain.CredentialServiceAccountKey
	require.ErrorIs(t, manager.verify(raw, tampered), developercredentialsdomain.ErrAuthenticationFailed)
	tampered = record
	tampered.DigestKey.Version++
	require.ErrorIs(t, manager.verify(raw, tampered), developercredentialsdomain.ErrAuthenticationFailed)
	require.ErrorIs(t, manager.verify(raw[:len(raw)-1]+"A", record), developercredentialsdomain.ErrAuthenticationFailed)
}

func TestEncodedKeyringRejectsAmbiguityAndDevelopmentMaterial(t *testing.T) {
	t.Parallel()
	encodedKey := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x31}, digestKeyBytes))
	tests := []struct {
		name string
		raw  string
	}{
		{name: "duplicate reference", raw: `{"one@1":"` + encodedKey + `","one@1":"` + encodedKey + `"}`},
		{name: "trailing object", raw: `{"one@1":"` + encodedKey + `"}{}`},
		{name: "unknown field shape", raw: `{"one@1":42}`},
		{name: "short material", raw: `{"one@1":"c2hvcnQ="}`},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseEncodedTokenKeyring("one", 1, test.raw)
			require.Error(t, err)
		})
	}

	renamedDevelopment, err := ParseEncodedTokenKeyring(
		"production", 1,
		`{"production@1":"ZGV2ZWxvcGVyLWNyZWRlbnRpYWwtZGV2LWtleS0wMDE="}`,
	)
	require.NoError(t, err)
	require.True(t, ContainsDevelopmentDigestKey(renamedDevelopment))
}

func TestTokenManagerRejectsReusedRotationMaterial(t *testing.T) {
	t.Parallel()
	material := bytes.Repeat([]byte{0x61}, digestKeyBytes)
	_, err := NewTokenManager(TokenKeyringConfig{
		Active: developercredentialsdomain.DigestKeyRef{ID: "new", Version: 2},
		Keys: []DigestKey{
			{Ref: developercredentialsdomain.DigestKeyRef{ID: "old", Version: 1}, Material: material},
			{Ref: developercredentialsdomain.DigestKeyRef{ID: "new", Version: 2}, Material: append([]byte(nil), material...)},
		},
	})
	require.ErrorContains(t, err, "independent material")
}

func FuzzParseToken(fuzz *testing.F) {
	fuzz.Add("")
	fuzz.Add("f41_pat_v1_000000000000_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	fuzz.Add("f41_sak_v1_ffffffffffff____________________________________________")
	fuzz.Fuzz(func(t *testing.T, raw string) {
		parsed, err := parseToken(raw)
		if err != nil {
			return
		}
		require.Len(t, parsed.Prefix, encodedPrefixLen)
		require.Len(t, parsed.Secret, encodedSecretLen)
	})
}
