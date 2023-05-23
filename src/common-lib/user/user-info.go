package user

import (
	"github.com/gorilla/mux"
	"net/http"
)

const (
	realmHeader             = `realm`
	userNameHeader          = `username`
	nocRealm                = `/activedirectory`
	iPlanetDirectoryPro     = `iPlanetDirectoryPro`
	iPlanetDirectoryProPlus = `iPlanetDirectoryProPlus`
	uidHeader               = `uid`
)

type Service interface {
	GetUser(r *http.Request, httpClient *http.Client) User
}

type service struct{}

func NewService() Service {
	return service{}
}

type User interface {
	PartnerID() string
	Name() string
	UID() string
	Token() string
	HasNOCAccess() bool
}

type user struct {
	name         string
	partnerID    string
	uid          string
	token        string
	hasNOCAccess bool
}

func (s service) GetUser(r *http.Request, httpClient *http.Client) User {
	var (
		partnerID    = mux.Vars(r)["partnerID"]
		name         = r.Header.Get(userNameHeader)
		uid          = r.Header.Get(uidHeader)
		hasNOCAccess = false
	)

	token := r.Header.Get(iPlanetDirectoryProPlus)
	if len(token) == 0 {
		token = r.Header.Get(iPlanetDirectoryPro)
	}

	if realm := r.Header.Get(realmHeader); realm == nocRealm {
		hasNOCAccess = true
	}

	return user{
		name:         name,
		partnerID:    partnerID,
		uid:          uid,
		token:        token,
		hasNOCAccess: hasNOCAccess,
	}
}

func (u user) PartnerID() string {
	return u.partnerID
}

func (u user) Name() string {
	return u.name
}

func (u user) UID() string {
	return u.uid
}

func (u user) Token() string {
	return u.token
}

func (u user) HasNOCAccess() bool {
	return u.hasNOCAccess
}
