package stories

import (
	"fmt"
	"strings"
)

const (
	DefaultEstimateScheme = "hours"
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

var estimateValueToLabel = map[string]map[int16]string{
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

var estimateDurationMinutes = map[string]map[int16]int{
	"points": {
		1: 120,
		2: 240,
		3: 360,
		5: 600,
		8: 960,
	},
	"hours": {
		1: 30,
		2: 60,
		3: 120,
		5: 240,
		8: 480,
	},
	"tshirt": {
		1: 60,
		2: 120,
		3: 240,
		5: 480,
		8: 960,
	},
	"ideal_days": {
		1: 240,
		2: 480,
		3: 960,
		5: 1440,
		8: 2400,
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

func ValidateEstimateValue(scheme string, estimateValue *int16) error {
	if estimateValue == nil {
		return nil
	}

	normalizedScheme := normalizeEstimateScheme(scheme)
	labelsByValue, ok := estimateValueToLabel[normalizedScheme]
	if !ok {
		return fmt.Errorf("invalid estimate scheme: %s", scheme)
	}

	if _, ok := labelsByValue[*estimateValue]; !ok {
		return fmt.Errorf("invalid estimate value '%d' for scheme '%s'. Allowed values only", *estimateValue, normalizedScheme)
	}

	return nil
}

func EstimateLabelFromValue(scheme string, estimateValue *int16) *string {
	if estimateValue == nil {
		return nil
	}

	normalizedScheme := normalizeEstimateScheme(scheme)
	labelsByValue, ok := estimateValueToLabel[normalizedScheme]
	if !ok {
		normalizedScheme = DefaultEstimateScheme
		labelsByValue = estimateValueToLabel[normalizedScheme]
	}

	value, ok := labelsByValue[*estimateValue]
	if !ok {
		return nil
	}
	return &value
}

func EstimateDurationMinutes(scheme string, estimateValue *int16) int {
	if estimateValue == nil {
		return 0
	}

	normalizedScheme := normalizeEstimateScheme(scheme)
	durationsByValue, ok := estimateDurationMinutes[normalizedScheme]
	if !ok {
		durationsByValue = estimateDurationMinutes[DefaultEstimateScheme]
	}

	return durationsByValue[*estimateValue]
}
