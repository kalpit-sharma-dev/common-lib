package uuid

import (
	"github.com/google/uuid"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exception"
)

const (
	//ErrCantParseUUIDString : Error if parse fails
	ErrCantParseUUIDString = "ErrCantParseUUIDString"
)

// NewRandomUUID : Generates the new random UUID
func NewRandomUUID() (newuuid uuid.UUID, err error) {
	return uuid.New(), nil
}

// ParseUUID : Parses the given string UUID
func ParseUUID(xxxx string) (parseuuid uuid.UUID, err error) {
	parseuuid, err = uuid.Parse(xxxx)
	if err != nil {
		err = exception.New(ErrCantParseUUIDString, nil)
	}
	return
}
