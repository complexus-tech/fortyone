package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlackRedirectURLDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Slack)
	redirectURL, found := configType.FieldByName("RedirectURL")
	require.True(t, found)
	require.Equal(t, "https://api.fortyone.app/integrations/slack/setup", redirectURL.Tag.Get("default"))
}
