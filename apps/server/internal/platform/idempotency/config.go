package idempotency

import (
	"errors"
	"fmt"
	"time"
)

const (
	DefaultLeaseDuration     = 2 * time.Minute
	DefaultRetentionDuration = 24 * time.Hour
	MinLeaseDuration         = time.Second
	MaxLeaseDuration         = 15 * time.Minute
	MinRetentionDuration     = time.Hour
	MaxRetentionDuration     = 30 * 24 * time.Hour
	MaxPurgeBatchSize        = 1000
)

var ErrInvalidConfig = errors.New("invalid idempotency configuration")

type Config struct {
	LeaseDuration     time.Duration
	RetentionDuration time.Duration
}

func DefaultConfig() Config {
	return Config{
		LeaseDuration:     DefaultLeaseDuration,
		RetentionDuration: DefaultRetentionDuration,
	}
}

func (c Config) validate() error {
	if c.LeaseDuration < MinLeaseDuration || c.LeaseDuration > MaxLeaseDuration {
		return fmt.Errorf(
			"%w: lease duration must be between %s and %s",
			ErrInvalidConfig,
			MinLeaseDuration,
			MaxLeaseDuration,
		)
	}
	if c.RetentionDuration < MinRetentionDuration || c.RetentionDuration > MaxRetentionDuration {
		return fmt.Errorf(
			"%w: retention duration must be between %s and %s",
			ErrInvalidConfig,
			MinRetentionDuration,
			MaxRetentionDuration,
		)
	}
	if c.RetentionDuration <= c.LeaseDuration {
		return fmt.Errorf("%w: retention duration must exceed lease duration", ErrInvalidConfig)
	}
	return nil
}
