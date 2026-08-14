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

func TestMayaSenderAddressDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Email)
	mayaAddress, found := configType.FieldByName("MayaAddress")
	require.True(t, found)
	require.Equal(t, "maya@fortyone.app", mayaAddress.Tag.Get("default"))
}

func TestMayaSenderNameDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Email)
	mayaName, found := configType.FieldByName("MayaName")
	require.True(t, found)
	require.Equal(t, "Maya, AI Agent", mayaName.Tag.Get("default"))
}
