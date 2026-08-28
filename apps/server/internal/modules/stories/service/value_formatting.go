package stories

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (s *Service) formatValue(value any) string {
	if value == nil {
		return "nil"
	}
	switch v := value.(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
		return "nil"
	case *int16:
		if v != nil {
			return fmt.Sprintf("%d", *v)
		}
		return "nil"
	case *int:
		if v != nil {
			return strconv.Itoa(*v)
		}
		return "nil"
	case int:
		return strconv.Itoa(v)
	case int16:
		return fmt.Sprintf("%d", v)
	case *float64:
		if v != nil {
			return fmt.Sprintf("%.2f", *v)
		}
		return "nil"
	case *uuid.UUID:
		if v != nil {
			return v.String()
		}
		return "nil"
	case *time.Time:
		if v != nil {
			return v.Format(time.RFC3339)
		}
		return "nil"
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func sameUUIDSet(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	values := make(map[uuid.UUID]struct{}, len(left))
	for _, value := range left {
		values[value] = struct{}{}
	}
	for _, value := range right {
		if _, exists := values[value]; !exists {
			return false
		}
	}
	return true
}

func (s *Service) valuesEqual(oldValue, newValue any) bool {
	return normalizeComparableValue(oldValue) == normalizeComparableValue(newValue)
}

func normalizeComparableValue(value any) string {
	if value == nil {
		return "nil"
	}

	switch v := value.(type) {
	case string:
		return v
	case *string:
		if v == nil {
			return "nil"
		}
		return *v
	case uuid.UUID:
		return v.String()
	case *uuid.UUID:
		if v == nil {
			return "nil"
		}
		return v.String()
	case time.Time:
		return v.UTC().Format(time.RFC3339Nano)
	case *time.Time:
		if v == nil {
			return "nil"
		}
		return v.UTC().Format(time.RFC3339Nano)
	case int:
		return strconv.Itoa(v)
	case *int:
		if v == nil {
			return "nil"
		}
		return strconv.Itoa(*v)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case *int16:
		if v == nil {
			return "nil"
		}
		return strconv.FormatInt(int64(*v), 10)
	case bool:
		return strconv.FormatBool(v)
	case *bool:
		if v == nil {
			return "nil"
		}
		return strconv.FormatBool(*v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
