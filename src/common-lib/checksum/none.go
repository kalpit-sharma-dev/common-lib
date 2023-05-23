package checksum

import (
	"io"
)

//none is the struct for no checksum validator
type none struct{}

//Calculate is the method to get the Sha1 validator checksum
func (n none) Calculate(reader io.Reader) (string, error) {
	return "", nil
}

//Validate is to verify and validate checksum
func (n none) Validate(reader io.Reader, checksum string) (bool, error) { // nolint
	return true, nil
}
