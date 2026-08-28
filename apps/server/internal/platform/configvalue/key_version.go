// Package configvalue contains strongly typed values used by process
// configuration schemas.
package configvalue

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// KeyVersion is a positive, 32-bit keyring version.
//
// Set implements github.com/josemukorivo/config.Setter. The configuration
// parser does not natively decode unsigned integers, so using uint32 directly
// would silently leave the field at zero.
type KeyVersion uint32

// Set parses and validates a key version from process configuration.
func (version *KeyVersion) Set(value string) error {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
	if err != nil {
		return fmt.Errorf("parse key version: %w", err)
	}
	if parsed == 0 {
		return errors.New("key version must be positive")
	}

	*version = KeyVersion(parsed)
	return nil
}

// Uint32 returns the version in the representation consumed by keyrings.
func (version KeyVersion) Uint32() uint32 {
	return uint32(version)
}
