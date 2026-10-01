package customfieldshttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefinitionIconOmissionAndResetRemainDistinct(t *testing.T) {
	for _, test := range []struct {
		body     string
		supplied bool
		icon     string
	}{
		{`{"name":"Amount","type":"number"}`, false, ""},
		{`{"name":"Amount","type":"number","icon":null}`, true, ""},
		{`{"name":"Amount","type":"number","icon":"star"}`, true, "star"},
	} {
		r := httptest.NewRequest(http.MethodPut, "/fields", strings.NewReader(test.body))
		r.Header.Set("Content-Type", "application/json")
		input, err := decodeDefinition(r)
		if err != nil || input.IconSet != test.supplied {
			t.Fatal("icon presence", input, err)
		}
		if test.icon == "" {
			if input.Icon != nil {
				t.Fatal("automatic icon was not null")
			}
		} else if input.Icon == nil || *input.Icon != test.icon {
			t.Fatal("selected icon was lost")
		}
	}
	for _, body := range []string{`{"name":"Amount","type":"number","icon":1}`, `{"name":"Amount","type":"number","icon":{},"unexpected":true}`, `{"name":"Amount","type":"number","unexpected":true}`} {
		r := httptest.NewRequest(http.MethodPut, "/fields", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if _, err := decodeDefinition(r); err == nil {
			t.Fatal("malformed or unknown definition property accepted")
		}
	}
}
