package user

import (
	"net/http"
)

type mock struct {
	name, partnerID, uid, token string
	hasNOCAccess                bool
}

func NewMock(
	name,
	partnerID,
	uid,
	token string,
	hasNOCAccess bool,
) Service {
	return mock{
		name:         name,
		partnerID:    partnerID,
		uid:          uid,
		token:        token,
		hasNOCAccess: hasNOCAccess,
	}
}

func (s mock) GetUser(r *http.Request, httpClient *http.Client) User {
	return user{
		name:         s.name,
		partnerID:    s.partnerID,
		uid:          s.uid,
		token:        s.token,
		hasNOCAccess: s.hasNOCAccess,
	}
}
