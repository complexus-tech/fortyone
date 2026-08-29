package safecast

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckedIntegerConversions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		convert func() error
		wantErr bool
	}{
		{name: "int64 zero to int", convert: func() error { _, err := Int64(0); return err }},
		{name: "int32 minimum", convert: func() error { _, err := Int32(math.MinInt32); return err }},
		{name: "int32 maximum", convert: func() error { _, err := Int32(math.MaxInt32); return err }},
		{name: "int32 overflow", convert: func() error { _, err := Int64ToInt32(math.MaxInt32 + 1); return err }, wantErr: true},
		{name: "int32 underflow", convert: func() error { _, err := Int64ToInt32(math.MinInt32 - 1); return err }, wantErr: true},
		{name: "uint32 maximum", convert: func() error { _, err := Uint32ToInt32(math.MaxInt32); return err }},
		{name: "uint32 overflow", convert: func() error { _, err := Uint32ToInt32(math.MaxInt32 + 1); return err }, wantErr: true},
		{name: "int16 minimum", convert: func() error { _, err := Int16(math.MinInt16); return err }},
		{name: "int16 maximum", convert: func() error { _, err := Int16(math.MaxInt16); return err }},
		{name: "int16 overflow", convert: func() error { _, err := Int16(math.MaxInt16 + 1); return err }, wantErr: true},
		{name: "int16 underflow", convert: func() error { _, err := Int16(math.MinInt16 - 1); return err }, wantErr: true},
		{name: "int32 to int16 maximum", convert: func() error { _, err := Int32ToInt16(math.MaxInt16); return err }},
		{name: "int32 to int16 overflow", convert: func() error { _, err := Int32ToInt16(math.MaxInt16 + 1); return err }, wantErr: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := test.convert()
			if test.wantErr {
				require.ErrorIs(t, err, ErrOutOfRange)
				return
			}
			require.NoError(t, err)
		})
	}
}

func FuzzInt64ToInt32(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(math.MaxInt32))
	f.Add(int64(math.MaxInt32) + 1)
	f.Fuzz(func(t *testing.T, value int64) {
		converted, err := Int64ToInt32(value)
		if value < math.MinInt32 || value > math.MaxInt32 {
			if !errors.Is(err, ErrOutOfRange) {
				t.Fatalf("out-of-range value %d returned %v", value, err)
			}
			return
		}
		if err != nil || int64(converted) != value {
			t.Fatalf("in-range value %d converted to %d with %v", value, converted, err)
		}
	})
}
