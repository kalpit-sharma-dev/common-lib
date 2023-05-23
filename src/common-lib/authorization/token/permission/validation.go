package permission

import (
	"github.com/dgrijalva/jwt-go"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token"
)

// ValidationType describes how permissions can be validated
// AnyOf - at least one required permission should be present
// AllOf - all of the permissions should be present in the JWT
type ValidationType int

const (
	// permission validation types
	AnyOf ValidationType = iota
	AllOf
)

// Validate decodes caller's permissions from the jwt and validates it against registered route permissions according to
// given validation type
func Validate(jwtToken string, routePermissions []string, validationType ValidationType) (bool, error) {
	permissions, err := Decode(jwtToken)
	if err != nil {
		return false, err
	}

	return ValidatePermissions(permissions, routePermissions, validationType), nil
}

// Decode simply decodes provided jwt and returns decoded permissions
func Decode(jwtToken string) ([]string, error) {
	jp := new(jwt.Parser)
	var claims token.AuthClaims
	_, _, err := jp.ParseUnverified(jwtToken, &claims)
	if err != nil {
		return nil, err
	}

	return claims.Permissions, nil
}

// Decode simply decodes provided jwt and returns decoded claims
func DecodeClaims(jwtToken string) (token.AuthClaims, error) {
	jp := new(jwt.Parser)
	var claims token.AuthClaims
	_, _, err := jp.ParseUnverified(jwtToken, &claims)
	if err != nil {
		return claims, err
	}

	return claims, nil
}

func ValidatePermissions(tokenPermissions, routePermissions []string, validationType ValidationType) bool {
	switch validationType {
	case AnyOf:
		return anyOf(routePermissions, tokenPermissions)
	case AllOf:
		return allOf(routePermissions, tokenPermissions)
	default:
		return false
	}
}

func anyOf(slice, subSlice []string) bool {
	for _, sv := range slice {
		for _, ssv := range subSlice {
			if sv == ssv {
				return true
			}
		}
	}

	return false
}

func allOf(slice, subSlice []string) bool {
	for _, sv := range slice {
		var found bool
		for _, ssv := range subSlice {
			if sv == ssv {
				found = true
			}
		}

		if !found {
			return false
		}
	}

	return true
}
