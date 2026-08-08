package slack

import "testing"

func TestCredentialCodecRoundTripAndLegacyUpgrade(t *testing.T) {
	t.Parallel()

	codec, err := newCredentialCodec("test-secret")
	if err != nil {
		t.Fatalf("newCredentialCodec() error = %v", err)
	}
	sealed, version, err := codec.seal(slackCredential{AccessToken: "xoxb-secret"})
	if err != nil {
		t.Fatalf("seal() error = %v", err)
	}
	if version == 0 || sealed == "xoxb-secret" {
		t.Fatalf("seal() = (%q, %d), want encrypted current version", sealed, version)
	}
	credential, openedVersion, err := codec.open(sealed)
	if err != nil {
		t.Fatalf("open() error = %v", err)
	}
	if credential.AccessToken != "xoxb-secret" || openedVersion != version {
		t.Fatalf("open() = (%#v, %d), want access token and version %d", credential, openedVersion, version)
	}

	legacy, legacyVersion, err := codec.open("xoxb-legacy")
	if err != nil {
		t.Fatalf("open(legacy) error = %v", err)
	}
	if legacy.AccessToken != "xoxb-legacy" || legacyVersion != 0 {
		t.Fatalf("open(legacy) = (%#v, %d)", legacy, legacyVersion)
	}
}
