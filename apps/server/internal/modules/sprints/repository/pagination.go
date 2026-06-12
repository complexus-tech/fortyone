package sprintsrepository

func positiveIntFilter(filters map[string]any, key string) (int, bool) {
	value, ok := intFilter(filters, key)
	return value, ok && value > 0
}

func nonNegativeIntFilter(filters map[string]any, key string) (int, bool) {
	value, ok := intFilter(filters, key)
	return value, ok && value >= 0
}

func intFilter(filters map[string]any, key string) (int, bool) {
	switch value := filters[key].(type) {
	case int:
		return value, true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	default:
		return 0, false
	}
}
