package entityreference

import (
	"fmt"
)

// ConflictHardReference is an errorMsg
const ConflictHardReference = "conflict"

// ConflictReferenceError is a custom error type
type ConflictReferenceError struct {
	ErrMsg     string
	References []*ReferenceResponse
}

func (c ConflictReferenceError) Error() string {
	return c.ErrMsg
}

// NewConflictReferenceError is a constructor for custom error
func NewConflictReferenceError(references []*Reference) ConflictReferenceError {
	err := ConflictReferenceError{ErrMsg: ConflictHardReference}
	for _, reference := range references {
		if reference.Type == Hard {
			err.References = append(err.References, &ReferenceResponse{
				EntityID:            reference.EntityID,
				ReferencingObjectID: reference.ReferencingObjectID,
				Service:             reference.Service,
				Type:                reference.Type,
			})
		}
	}
	return err
}

// MsgError represents error message
type MsgError struct {
	Message string `json:"error"`
	Desc    string `json:"description"`
}

// NotFoundError represents not found error
type NotFoundError struct {
	MsgError
}

func (e MsgError) Error() string {
	return fmt.Sprintf("%s, %s", e.Message, e.Desc)
}
