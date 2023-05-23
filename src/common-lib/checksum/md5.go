package checksum

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"strings"
)

type md5Impl struct{}

//Calculate is the method to get the MD5 validator checksum
func (c md5Impl) Calculate(reader io.Reader) (string, error) {
	h := md5.New()
	if _, err := io.Copy(h, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

//VerifyCheckSum is to verify and validate checksum
func (c md5Impl) Validate(reader io.Reader, checksum string) (bool, error) {
	calculatedVal, err := c.Calculate(reader)
	if err != nil {
		return false, fmt.Errorf("Checksum cannot be Calculated from downloaded component, Err : %v", err)
	}
	trimmedChecksum := strings.TrimSpace(checksum)
	trimmedCalculatedVal := strings.TrimSpace(calculatedVal)
	if strings.ToUpper(trimmedCalculatedVal) != strings.ToUpper(trimmedChecksum) {
		return false, errors.New("Checksum validation failed")
	}
	return true, nil
}
