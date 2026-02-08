package stories

import (
	"fmt"
	"strings"
)

const (
	DefaultEstimateScheme = "points"
)

var allowedEstimateSchemes = map[string]map[string]int16{
	"points": {
		"1": 1,
		"2": 2,
		"3": 3,
		"5": 5,
		"8": 8,
	},
	"hours": {
		"0.5": 1,
		"1":   2,
		"2":   3,
		"4":   5,
		"8":   8,
	},
	"tshirt": {
		"XS": 1,
		"S":  2,
		"M":  3,
		"L":  5,
		"XL": 8,
	},
	"ideal_days": {
		"0.5": 1,
		"1":   2,
		"2":   3,
		"3":   5,
		"5":   8,
	},
}

var estimateUnitToValue = map[string]map[int16]string{
	"points": {
		1: "1",
		2: "2",
		3: "3",
		5: "5",
		8: "8",
	},
	"hours": {
		1: "0.5",
		2: "1",
		3: "2",
		5: "4",
		8: "8",
	},
	"tshirt": {
		1: "XS",
		2: "S",
		3: "M",
		5: "L",
		8: "XL",
	},
	"ideal_days": {
		1: "0.5",
		2: "1",
		3: "2",
		5: "3",
		8: "5",
	},
}

func normalizeEstimateScheme(scheme string) string {
	scheme = strings.TrimSpace(strings.ToLower(scheme))
	if scheme == "" {
		return DefaultEstimateScheme
	}
	return scheme
}

func ValidateEstimateScheme(scheme string) error {
	normalized := normalizeEstimateScheme(scheme)
	if _, ok := allowedEstimateSchemes[normalized]; !ok {
		return fmt.Errorf("invalid estimate scheme: %s", scheme)
	}
	return nil
}

func NormalizeEstimateValue(scheme string, value *string) (*int16, error) {
	if value == nil {
		return nil, nil
	}

	normalizedScheme := normalizeEstimateScheme(scheme)
	allowedValues, ok := allowedEstimateSchemes[normalizedScheme]
	if !ok {
		return nil, fmt.Errorf("invalid estimate scheme: %s", scheme)
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}

	// Only tshirt keeps uppercase semantics.
	if normalizedScheme == "tshirt" {
		trimmed = strings.ToUpper(trimmed)
	}

	normalizedUnit, ok := allowedValues[trimmed]
	if !ok {
		return nil, fmt.Errorf("invalid estimate '%s' for scheme '%s'. Allowed values only", trimmed, normalizedScheme)
	}

	return &normalizedUnit, nil
}

func EstimateValueFromUnit(scheme string, estimateUnit *int16) *string {
	if estimateUnit == nil {
		return nil
	}

	normalizedScheme := normalizeEstimateScheme(scheme)
	valuesByUnit, ok := estimateUnitToValue[normalizedScheme]
	if !ok {
		normalizedScheme = DefaultEstimateScheme
		valuesByUnit = estimateUnitToValue[normalizedScheme]
	}

	value, ok := valuesByUnit[*estimateUnit]
	if !ok {
		return nil
	}
	return &value
}
