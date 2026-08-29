package web

import (
	"net/url"
	"strings"
	"testing"
)

func FuzzOptionalQueryParameterNeverPanics(f *testing.F) {
	f.Add("search", "roadmap", uint16(64), false)
	f.Add("teamId", "not-a-uuid", uint16(1), true)
	f.Fuzz(func(t *testing.T, name, value string, limit uint16, repeated bool) {
		maxBytes := int(limit) + 1
		values := url.Values{name: {value}}
		if repeated {
			values[name] = append(values[name], value)
		}
		_, _, _ = OptionalQueryParameter(values, name, maxBytes)
	})
}

func FuzzOptionalListQueryParameterNeverPanics(f *testing.F) {
	f.Add("started,completed", "paused", uint8(4), uint8(16), uint16(64))
	f.Add("sensitive-list-value", "", uint8(1), uint8(1), uint16(1))
	f.Fuzz(func(t *testing.T, first, second string, rawItems, rawItemBytes uint8, rawBytes uint16) {
		first = "sensitive-first<" + first + ">"
		second = "sensitive-second<" + second + ">"
		limits := QueryListLimits{
			MaxBytes:     int(rawBytes) + 1,
			MaxItemBytes: int(rawItemBytes) + 1,
			MaxItems:     int(rawItems) + 1,
		}
		_, _, err := OptionalListQueryParameter(url.Values{"filters": {first, second}}, "filters", limits)
		if err != nil {
			if strings.Contains(err.Error(), first) {
				t.Fatal("error exposes first supplied value")
			}
			if strings.Contains(err.Error(), second) {
				t.Fatal("error exposes second supplied value")
			}
		}
	})
}

func FuzzOptionalTextQueryParameterNeverPanics(f *testing.F) {
	f.Add("roadmap", uint16(64), uint16(32), false)
	f.Add("contains\x00nul", uint16(64), uint16(32), true)
	f.Fuzz(func(t *testing.T, value string, rawBytes, rawRunes uint16, repeated bool) {
		supplied := "sensitive-text<" + value + ">"
		values := url.Values{"search": {supplied}}
		if repeated {
			values["search"] = append(values["search"], supplied)
		}
		_, _, err := OptionalTextQueryParameter(values, "search", int(rawBytes)+1, int(rawRunes)+1)
		if err != nil && strings.Contains(err.Error(), supplied) {
			t.Fatal("error exposes supplied value")
		}
	})
}

func FuzzOptionalCommaSeparatedUUIDQueryParameterNeverPanics(f *testing.F) {
	f.Add("00000000-0000-4000-8000-000000000001", uint8(1))
	f.Add("not-a-uuid,00000000-0000-4000-8000-000000000001", uint8(10))
	f.Fuzz(func(t *testing.T, value string, rawMaxItems uint8) {
		maxItems := int(rawMaxItems) + 1
		_, _ = OptionalCommaSeparatedUUIDQueryParameter(url.Values{"ids": {value}}, "ids", maxItems)
	})
}

func FuzzOptionalIntegerQueryParameterNeverPanics(f *testing.F) {
	f.Add("25", int16(1), int16(100))
	f.Add("not-an-integer", int16(-10), int16(10))
	f.Fuzz(func(t *testing.T, value string, minimum, maximum int16) {
		_, _, _ = OptionalIntegerQueryParameter(
			url.Values{"page": {value}},
			"page",
			32,
			int(minimum),
			int(maximum),
		)
	})
}

func FuzzOptionalBooleanQueryParameterNeverPanics(f *testing.F) {
	f.Add("true")
	f.Add("false")
	f.Add("TRUE")
	f.Add(strings.Repeat("x", DefaultMaxQueryParameterBytes+1))

	f.Fuzz(func(t *testing.T, value string) {
		supplied := "sensitive-query-value<" + value + ">"
		_, _, err := OptionalBooleanQueryParameter(url.Values{"enabled": {supplied}}, "enabled")
		if err != nil && strings.Contains(err.Error(), supplied) {
			t.Fatal("error exposes supplied value")
		}
	})
}
