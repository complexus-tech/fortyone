package web

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestOptionalQueryParameterIsBoundedAndUnambiguous(t *testing.T) {
	t.Parallel()

	value, present, err := OptionalQueryParameter(url.Values{"search": {"typed API"}}, "search", 32)
	if err != nil || !present || value != "typed API" {
		t.Fatalf("OptionalQueryParameter() = %q, %v, %v", value, present, err)
	}
	if _, present, err := OptionalQueryParameter(url.Values{}, "search", 32); err != nil || present {
		t.Fatalf("absent parameter = present %v error %v", present, err)
	}
	if _, _, err := OptionalQueryParameter(url.Values{"search": {"first", "second"}}, "search", 32); !errors.Is(err, ErrRepeatedQueryParameter) {
		t.Fatalf("repeated parameter error = %v", err)
	}
	if _, _, err := OptionalQueryParameter(url.Values{"search": {strings.Repeat("x", 33)}}, "search", 32); !errors.Is(err, ErrQueryParameterTooLong) {
		t.Fatalf("oversized parameter error = %v", err)
	}
}

func TestOpaqueQueryParameterPreservesValidTokensAndRejectsControls(t *testing.T) {
	t.Parallel()

	want := "  signed-token-value  "
	got, present, err := OptionalOpaqueQueryParameter(url.Values{"state": {want}}, "state", 64)
	if err != nil || !present || got != want {
		t.Fatalf("OptionalOpaqueQueryParameter() = %q, %v, %v", got, present, err)
	}
	required, err := RequiredOpaqueQueryParameter(url.Values{"state": {want}}, "state", 64)
	if err != nil || required != want {
		t.Fatalf("RequiredOpaqueQueryParameter() = %q, %v", required, err)
	}

	for name, values := range map[string]url.Values{
		"missing":      {},
		"blank":        {"state": {"  "}},
		"nul":          {"state": {"secret\x00value"}},
		"control":      {"state": {"secret\nvalue"}},
		"invalid utf8": {"state": {string([]byte{0xff})}},
		"repeated":     {"state": {"one", "two"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := RequiredOpaqueQueryParameter(values, "state", 64)
			if err == nil {
				t.Fatal("expected opaque query error")
			}
		})
	}
}

func TestOptionalListQueryParameterIsBoundedAndDeterministic(t *testing.T) {
	t.Parallel()

	limits := QueryListLimits{MaxBytes: 64, MaxItemBytes: 16, MaxItems: 4}
	items, present, err := OptionalListQueryParameter(url.Values{
		"status": {"started, completed", "started", "paused"},
	}, "status", limits)
	if err != nil || !present {
		t.Fatalf("OptionalListQueryParameter() present = %v, error = %v", present, err)
	}
	want := []string{"started", "completed", "paused"}
	if strings.Join(items, ",") != strings.Join(want, ",") {
		t.Fatalf("items = %#v, want %#v", items, want)
	}
	if items, present, err := OptionalListQueryParameter(url.Values{}, "status", limits); err != nil || present || items != nil {
		t.Fatalf("absent list = %#v, %v, %v", items, present, err)
	}

	secretLikeValue := "sensitive-list-value"
	for name, test := range map[string]struct {
		values url.Values
		limits QueryListLimits
		cause  error
	}{
		"empty item": {
			values: url.Values{"status": {"started,,completed"}}, limits: limits, cause: ErrInvalidQueryParameter,
		},
		"empty parameter": {
			values: url.Values{"status": {""}}, limits: limits, cause: ErrInvalidQueryParameter,
		},
		"too many": {
			values: url.Values{"status": {"one,two,three"}}, limits: QueryListLimits{MaxBytes: 64, MaxItemBytes: 16, MaxItems: 2}, cause: ErrInvalidQueryParameter,
		},
		"item too long": {
			values: url.Values{"status": {secretLikeValue}}, limits: limits, cause: ErrQueryParameterTooLong,
		},
		"combined value too long": {
			values: url.Values{"status": {"123456789", "123456789"}}, limits: QueryListLimits{MaxBytes: 16, MaxItemBytes: 16, MaxItems: 4}, cause: ErrQueryParameterTooLong,
		},
		"invalid limits": {
			values: url.Values{"status": {"started"}}, limits: QueryListLimits{}, cause: ErrInvalidQueryParameter,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := OptionalListQueryParameter(test.values, "status", test.limits)
			if !errors.Is(err, test.cause) {
				t.Fatalf("error = %v, want %v", err, test.cause)
			}
			if strings.Contains(err.Error(), secretLikeValue) {
				t.Fatalf("error %q exposes query value", err)
			}
		})
	}
}

func TestOptionalTextQueryParameterValidatesTextWithoutEchoingIt(t *testing.T) {
	t.Parallel()

	value, present, err := OptionalTextQueryParameter(url.Values{"search": {"  roadmap  "}}, "search", 32, 16)
	if err != nil || !present || value != "roadmap" {
		t.Fatalf("OptionalTextQueryParameter() = %q, %v, %v", value, present, err)
	}
	if value, present, err := OptionalTextQueryParameter(url.Values{}, "search", 32, 16); err != nil || present || value != "" {
		t.Fatalf("absent text = %q, %v, %v", value, present, err)
	}

	for name, test := range map[string]url.Values{
		"invalid utf8":   {"search": {string([]byte{0xff, 0xfe})}},
		"nul":            {"search": {"sensitive\x00value"}},
		"too many runes": {"search": {"sensitive-text-value"}},
		"repeated":       {"search": {"one", "two"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := OptionalTextQueryParameter(test, "search", 64, 4)
			if err == nil {
				t.Fatal("expected text query error")
			}
			for _, supplied := range test["search"] {
				if supplied != "" && strings.Contains(err.Error(), supplied) {
					t.Fatalf("error %q exposes supplied text", err)
				}
			}
		})
	}
	if _, _, err := OptionalTextQueryParameter(url.Values{"search": {"ok"}}, "search", 32, 0); !errors.Is(err, ErrInvalidQueryParameter) {
		t.Fatalf("invalid limit error = %v", err)
	}
}

func TestRequiredTextQueryParameterRejectsMissingAndBlankValues(t *testing.T) {
	t.Parallel()

	value, err := RequiredTextQueryParameter(url.Values{"state": {"  opaque-state  "}}, "state", 64, 64)
	if err != nil || value != "opaque-state" {
		t.Fatalf("RequiredTextQueryParameter() = %q, %v", value, err)
	}

	for name, values := range map[string]url.Values{
		"missing":  {},
		"blank":    {"state": {"  "}},
		"repeated": {"state": {"first", "second"}},
		"oversized": {
			"state": {strings.Repeat("sensitive", 16)},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := RequiredTextQueryParameter(values, "state", 64, 64)
			if err == nil {
				t.Fatal("expected required text query error")
			}
			for _, supplied := range values["state"] {
				if supplied != "" && strings.Contains(err.Error(), supplied) {
					t.Fatalf("error %q exposes supplied text", err)
				}
			}
		})
	}
}

func TestParseOAuthCallbackQuery(t *testing.T) {
	t.Parallel()

	callback, err := ParseOAuthCallbackQuery(url.Values{
		"code":  {"authorization-code"},
		"state": {"opaque-state"},
	})
	if err != nil {
		t.Fatalf("ParseOAuthCallbackQuery() error = %v", err)
	}
	if callback.Code != "authorization-code" || callback.State != "opaque-state" || callback.ProviderError != "" {
		t.Fatalf("callback = %#v", callback)
	}

	errorCallback, err := ParseOAuthCallbackQuery(url.Values{
		"state": {"opaque-state"},
		"error": {"access_denied"},
	})
	if err != nil || errorCallback.ProviderError != "access_denied" || errorCallback.Code != "" {
		t.Fatalf("error callback = %#v, %v", errorCallback, err)
	}

	secretLikeValue := "sensitive-oauth-callback-value"
	for name, values := range map[string]url.Values{
		"missing state":  {"code": {"code"}},
		"missing code":   {"state": {"state"}},
		"repeated state": {"code": {"code"}, "state": {"first", "second"}},
		"repeated code on error": {
			"code": {"first", "second"}, "state": {"state"}, "error": {"access_denied"},
		},
		"oversized error": {
			"state": {"state"}, "error": {strings.Repeat(secretLikeValue, 16)},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseOAuthCallbackQuery(values)
			if err == nil {
				t.Fatal("expected OAuth callback query error")
			}
			if strings.Contains(err.Error(), secretLikeValue) {
				t.Fatalf("error %q exposes OAuth callback value", err)
			}
		})
	}
}

func TestOptionalIntegerQueryParameter(t *testing.T) {
	t.Parallel()

	value, present, err := OptionalIntegerQueryParameter(url.Values{"page": {" 25 "}}, "page", 20, 1, 100)
	if err != nil || !present || value != 25 {
		t.Fatalf("integer parameter = %d, %v, %v", value, present, err)
	}
	if _, present, err := OptionalIntegerQueryParameter(url.Values{}, "page", 20, 1, 100); err != nil || present {
		t.Fatalf("absent integer parameter = present %v error %v", present, err)
	}

	sensitiveValue := "invalid-sensitive-number"
	for name, query := range map[string]url.Values{
		"blank":         {"page": {""}},
		"malformed":     {"page": {sensitiveValue}},
		"too small":     {"page": {"0"}},
		"too large":     {"page": {"101"}},
		"overflow":      {"page": {strings.Repeat("9", 20)}},
		"repeated":      {"page": {"1", "2"}},
		"oversized raw": {"page": {strings.Repeat("1", 21)}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := OptionalIntegerQueryParameter(query, "page", 20, 1, 100)
			if err == nil || strings.Contains(err.Error(), sensitiveValue) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestOptionalBooleanQueryParameterUsesCanonicalProtocolValues(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		query       url.Values
		want        bool
		wantPresent bool
	}{
		"missing": {query: url.Values{}},
		"true": {
			query:       url.Values{"enabled": {"true"}},
			want:        true,
			wantPresent: true,
		},
		"false": {
			query:       url.Values{"enabled": {"false"}},
			wantPresent: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, present, err := OptionalBooleanQueryParameter(test.query, "enabled")
			if err != nil {
				t.Fatalf("OptionalBooleanQueryParameter() error = %v", err)
			}
			if got != test.want || present != test.wantPresent {
				t.Fatalf("OptionalBooleanQueryParameter() = %t, %t; want %t, %t", got, present, test.want, test.wantPresent)
			}
		})
	}
}

func TestOptionalBooleanQueryParameterRejectsAmbiguousValuesWithoutEchoingThem(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		query url.Values
		cause error
	}{
		"repeated": {
			query: url.Values{"enabled": {"true", "false"}},
			cause: ErrRepeatedQueryParameter,
		},
		"empty": {
			query: url.Values{"enabled": {""}},
			cause: ErrInvalidQueryParameter,
		},
		"non canonical": {
			query: url.Values{"enabled": {"TRUE"}},
			cause: ErrInvalidQueryParameter,
		},
		"surrounding whitespace": {
			query: url.Values{"enabled": {" true "}},
			cause: ErrInvalidQueryParameter,
		},
		"oversized": {
			query: url.Values{"enabled": {strings.Repeat("sensitive", DefaultMaxQueryParameterBytes)}},
			cause: ErrQueryParameterTooLong,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := OptionalBooleanQueryParameter(test.query, "enabled")
			if !errors.Is(err, test.cause) {
				t.Fatalf("OptionalBooleanQueryParameter() error = %v, want %v", err, test.cause)
			}
			for _, values := range test.query {
				for _, value := range values {
					if value != "" && strings.Contains(err.Error(), value) {
						t.Fatalf("error %q exposes query value", err)
					}
				}
			}
		})
	}
}

func TestOptionalUUIDQueryParameterRejectsInvalidValuesWithoutEchoingThem(t *testing.T) {
	t.Parallel()

	want := uuid.New()
	got, err := OptionalUUIDQueryParameter(url.Values{"teamId": {want.String()}}, "teamId")
	if err != nil || got == nil || *got != want {
		t.Fatalf("OptionalUUIDQueryParameter() = %v, %v, want %s", got, err, want)
	}
	if got, err := OptionalUUIDQueryParameter(url.Values{"teamId": {""}}, "teamId"); err != nil || got != nil {
		t.Fatalf("empty optional UUID = %v, %v", got, err)
	}
	secretLikeValue := "not-a-uuid-sensitive-value"
	if _, err := OptionalUUIDQueryParameter(url.Values{"teamId": {secretLikeValue}}, "teamId"); !errors.Is(err, ErrInvalidQueryParameter) || strings.Contains(err.Error(), secretLikeValue) {
		t.Fatalf("invalid UUID error = %v", err)
	}
}

func TestOptionalCommaSeparatedUUIDQueryParameter(t *testing.T) {
	t.Parallel()

	first, second := uuid.New(), uuid.New()
	got, err := OptionalCommaSeparatedUUIDQueryParameter(url.Values{
		"teamIds": {first.String() + ", ," + second.String()},
	}, "teamIds", 2)
	if err != nil {
		t.Fatalf("parse UUID list: %v", err)
	}
	if len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("UUID list = %#v", got)
	}

	for name, query := range map[string]url.Values{
		"repeated parameter": {"teamIds": {first.String(), second.String()}},
		"too many items":     {"teamIds": {first.String() + "," + second.String()}},
		"invalid UUID":       {"teamIds": {"not-a-uuid-sensitive-value"}},
		"zero UUID":          {"teamIds": {uuid.Nil.String()}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			maxItems := 2
			if name == "too many items" {
				maxItems = 1
			}
			_, err := OptionalCommaSeparatedUUIDQueryParameter(query, "teamIds", maxItems)
			if err == nil || strings.Contains(err.Error(), "not-a-uuid-sensitive-value") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
