package checksum

import (
	"errors"
	"io"
	"strings"
)

//go:generate mockgen -package mock -destination=mock/mocks.go . Service

//Service interface for methods for checksums
type Service interface {
	Calculate(reader io.Reader) (string, error)
	Validate(reader io.Reader, checksum string) (bool, error)
}

//Type is the checksum type
type Type struct {
	order int
	Name  string
	value string
}

var (
	//NONE is a type used for no check sum calculation and validation
	NONE = Type{order: 1, Name: "NONE", value: "\"NONE\""}

	//MD5 is a type used for md5 check sum calculation and validation
	MD5 = Type{order: 2, Name: "MD5", value: "\"MD5\""}

	//SHA1 is a type used for SHA1 check sum calculation and validation
	SHA1 = Type{order: 3, Name: "SHA1", value: "\"SHA1\""}

	//SHA256 is a type used for SHA256 check sum calculation and validation
	SHA256 = Type{order: 4, Name: "SHA256", value: "\"SHA256\""}

	//ErrChecksumInvalid is returned if validation failed du to invalid value
	ErrChecksumInvalid = "ErrChecksumInvalid"

	//ErrUnsupportedType is returned if invalid type is supplied
	ErrUnsupportedType = "ErrUnsupportedChecksumType"
)

// GetType - helper to get type by name
func GetType(data string) Type {
	tp := strings.ToUpper(strings.TrimSpace(data))
	switch tp {
	case MD5.Name:
		return MD5
	case SHA1.Name:
		return SHA1
	case SHA256.Name:
		return SHA256
	}
	return NONE
}

// UnmarshalJSON is a function to unmarshal Destination
// The default value is FILE
func (t *Type) UnmarshalJSON(data []byte) error {
	tp := strings.ToUpper(strings.TrimSpace(string(data)))
	typ := NONE
	switch tp {
	case MD5.value:
		typ = MD5
	case SHA1.value:
		typ = SHA1
	case SHA256.value:
		typ = SHA256
	}

	t.value = typ.value
	t.Name = typ.Name
	t.order = typ.order
	return nil
}

//MarshalJSON is a function to marshal Destination
func (t Type) MarshalJSON() ([]byte, error) {
	return []byte(t.value), nil
}

//GetService is the method to get the checksum type sum method
func GetService(cType Type) (Service, error) {
	switch cType {
	case MD5:
		return md5Impl{}, nil
	case SHA1:
		return sha1Impl{}, nil
	case SHA256:
		return sha256Impl{}, nil
	case NONE:
		return none{}, nil
	default:
		return nil, errors.New(ErrUnsupportedType)
	}
}
