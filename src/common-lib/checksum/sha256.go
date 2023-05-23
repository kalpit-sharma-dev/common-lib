package checksum

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"strings"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

// sha256Impl is the struct for Sha1 validator
type sha256Impl struct{}

// Calculate is the method to get the Sha1 validator checksum
func (s sha256Impl) Calculate(reader io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// Validate is to verify and validate checksum
func (s sha256Impl) Validate(reader io.Reader, checksum string) (bool, error) {
	sum, err := s.Calculate(reader)
	if err != nil {
		return false, fmt.Errorf("Checksum cannot be Calculated from downloaded component, Err : %v", err)
	}
	log := logger.Get()
	if strings.ToUpper(sum) != strings.ToUpper(checksum) {
		log.Trace("", "Invalid CheckSum :- Calculated checksum : %s & downloaded file checksum : %s", sum, checksum)
		return false, errors.New(ErrChecksumInvalid)
	}
	return true, nil
}
