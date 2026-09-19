package notifications

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRegisterPushDeviceValidatesExpoTokenAndPlatform(t *testing.T) {
	t.Parallel()
	require.NoError(t, (RegisterPushDevice{
		UserID: uuid.New(), Token: "ExponentPushToken[abcdefghijklmnop]", Platform: PushPlatformIOS,
	}).Validate())
	require.ErrorIs(t, (RegisterPushDevice{
		UserID: uuid.New(), Token: "native-token-without-expo-prefix", Platform: PushPlatformAndroid,
	}).Validate(), ErrInvalid)
	require.ErrorIs(t, (RegisterPushDevice{
		UserID: uuid.New(), Token: "ExpoPushToken[abcdefghijklmnop]", Platform: "desktop",
	}).Validate(), ErrInvalid)
}
