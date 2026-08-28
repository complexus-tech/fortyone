package deployment

import (
	"fmt"
	"strings"
)

// Mode is the authoritative runtime deployment mode shared by every process.
// Feature-specific environment flags must not be used to make security
// decisions.
type Mode string

const (
	Development Mode = "development"
	Test        Mode = "test"
	Staging     Mode = "staging"
	Production  Mode = "production"
)

// Parse normalizes and validates an APP_ENVIRONMENT value.
func Parse(value string) (Mode, error) {
	mode := Mode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case Development, Test, Staging, Production:
		return mode, nil
	default:
		return "", fmt.Errorf(
			"APP_ENVIRONMENT must be one of %q, %q, %q, or %q",
			Development,
			Test,
			Staging,
			Production,
		)
	}
}

func (m Mode) IsProduction() bool {
	return m == Production
}

func (m Mode) String() string {
	return string(m)
}
