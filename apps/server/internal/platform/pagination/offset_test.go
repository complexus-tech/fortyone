package pagination

import (
	"errors"
	"math"
	"net/url"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/web"
)

func TestOffsetParamsOffset(t *testing.T) {
	t.Parallel()

	if got := (OffsetParams{Page: 3, PageSize: 25}).Offset(); got != 50 {
		t.Fatalf("Offset() = %d, want 50", got)
	}
	if got := (OffsetParams{Page: math.MaxInt, PageSize: 100}).Offset(); got != math.MaxInt {
		t.Fatalf("overflow Offset() = %d, want MaxInt", got)
	}
}

func TestOffsetRequested(t *testing.T) {
	t.Parallel()

	if OffsetRequested(url.Values{}) {
		t.Fatal("OffsetRequested() = true without pagination parameters")
	}
	if !OffsetRequested(url.Values{"pageSize": {""}}) {
		t.Fatal("OffsetRequested() = false for explicit pageSize")
	}
}

func TestParseOffsetQueryAppliesStrictInputAndDocumentedBounds(t *testing.T) {
	t.Parallel()

	config := OffsetQueryConfig{DefaultPageSize: 20, MaximumPageSize: 50, MaximumOffset: math.MaxInt32}
	for name, test := range map[string]struct {
		query url.Values
		want  OffsetParams
	}{
		"defaults": {
			query: url.Values{},
			want:  OffsetParams{Page: 1, PageSize: 20},
		},
		"valid": {
			query: url.Values{"page": {"3"}, "pageSize": {"25"}},
			want:  OffsetParams{Page: 3, PageSize: 25},
		},
		"page size cap": {
			query: url.Values{"pageSize": {"500"}},
			want:  OffsetParams{Page: 1, PageSize: 50},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseOffsetQuery(test.query, config)
			if err != nil {
				t.Fatalf("ParseOffsetQuery() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("ParseOffsetQuery() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseOffsetQueryRejectsAmbiguousMalformedAndOverflowingValues(t *testing.T) {
	t.Parallel()

	config := OffsetQueryConfig{DefaultPageSize: 20, MaximumPageSize: 50, MaximumOffset: math.MaxInt32}
	for name, test := range map[string]struct {
		query url.Values
		cause error
	}{
		"repeated page": {
			query: url.Values{"page": {"1", "2"}},
			cause: web.ErrRepeatedQueryParameter,
		},
		"empty page": {
			query: url.Values{"page": {""}},
			cause: web.ErrInvalidQueryParameter,
		},
		"malformed page": {
			query: url.Values{"page": {"sensitive-page-value"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"non-positive page size": {
			query: url.Values{"pageSize": {"0"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"oversized integer": {
			query: url.Values{"page": {strings.Repeat("9", maximumIntegerParameterBytes+1)}},
			cause: web.ErrQueryParameterTooLong,
		},
		"offset overflow": {
			query: url.Values{"page": {"2147483649"}, "pageSize": {"1"}},
			cause: web.ErrInvalidQueryParameter,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseOffsetQuery(test.query, config)
			if !errors.Is(err, test.cause) {
				t.Fatalf("ParseOffsetQuery() error = %v, want %v", err, test.cause)
			}
			for _, values := range test.query {
				for _, value := range values {
					if value != "" && err != nil && strings.Contains(err.Error(), value) {
						t.Fatalf("error %q exposes query value", err)
					}
				}
			}
		})
	}
}

func TestParseOffsetQueryRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	for _, config := range []OffsetQueryConfig{
		{},
		{DefaultPageSize: 20, MaximumPageSize: 10, MaximumOffset: math.MaxInt32},
		{DefaultPageSize: 20, MaximumPageSize: 50, MaximumOffset: -1},
	} {
		if _, err := ParseOffsetQuery(url.Values{}, config); !errors.Is(err, ErrInvalidOffsetQueryConfig) {
			t.Fatalf("ParseOffsetQuery(%#v) error = %v", config, err)
		}
	}
}
