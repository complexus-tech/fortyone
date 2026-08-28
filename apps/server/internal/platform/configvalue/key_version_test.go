package configvalue_test

import (
	"os"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/configvalue"
	configloader "github.com/josemukorivo/config"
	"github.com/stretchr/testify/require"
)

func TestKeyVersionSupportsConfigParserDefaults(t *testing.T) {
	const environmentKey = "FORTYONE_CONFIGVALUE_TEST_KEY_VERSION"
	previous, existed := os.LookupEnv(environmentKey)
	require.NoError(t, os.Unsetenv(environmentKey))
	t.Cleanup(func() {
		if existed {
			require.NoError(t, os.Setenv(environmentKey, previous))
			return
		}
		require.NoError(t, os.Unsetenv(environmentKey))
	})

	var cfg struct {
		Version configvalue.KeyVersion `default:"1" env:"FORTYONE_CONFIGVALUE_TEST_KEY_VERSION"`
	}
	require.NoError(t, configloader.Parse("", &cfg))
	require.Equal(t, uint32(1), cfg.Version.Uint32())
}

func TestKeyVersionRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "0", "-1", "4294967296", "not-a-number"} {
		value := value
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			var version configvalue.KeyVersion
			require.Error(t, version.Set(value))
		})
	}
}
