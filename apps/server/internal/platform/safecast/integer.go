// Package safecast provides the checked integer conversions used at typed
// persistence boundaries. SQLC represents PostgreSQL smallint and integer as
// int16 and int32, while transport and domain validation commonly use Go int.
package safecast

import (
	"errors"
	"fmt"
	"math"
)

var ErrOutOfRange = errors.New("integer is outside the destination range")

// Int64 converts a database count to the platform-sized Go integer used by
// bounded domain collections and transport models.
func Int64(value int64) (int, error) {
	if value < int64(math.MinInt) || value > int64(math.MaxInt) {
		return 0, fmt.Errorf("convert %d to int: %w", value, ErrOutOfRange)
	}
	return int(value), nil
}

// Int32 converts a Go int to the PostgreSQL integer representation.
func Int32(value int) (int32, error) {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, fmt.Errorf("convert %d to int32: %w", value, ErrOutOfRange)
	}
	return int32(value), nil
}

// Int64ToInt32 converts an int64 to the PostgreSQL integer representation.
func Int64ToInt32(value int64) (int32, error) {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, fmt.Errorf("convert %d to int32: %w", value, ErrOutOfRange)
	}
	return int32(value), nil
}

// Uint32ToInt32 converts an unsigned key generation to a PostgreSQL integer.
func Uint32ToInt32(value uint32) (int32, error) {
	if value > math.MaxInt32 {
		return 0, fmt.Errorf("convert %d to int32: %w", value, ErrOutOfRange)
	}
	return int32(value), nil
}

// Int16 converts a Go int to the PostgreSQL smallint representation.
func Int16(value int) (int16, error) {
	if value < math.MinInt16 || value > math.MaxInt16 {
		return 0, fmt.Errorf("convert %d to int16: %w", value, ErrOutOfRange)
	}
	return int16(value), nil
}

// Int32ToInt16 converts a transport or PostgreSQL integer to a smallint-backed
// domain value without allowing wraparound.
func Int32ToInt16(value int32) (int16, error) {
	if value < math.MinInt16 || value > math.MaxInt16 {
		return 0, fmt.Errorf("convert %d to int16: %w", value, ErrOutOfRange)
	}
	return int16(value), nil
}
