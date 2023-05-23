package permission

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token"
)

func generateToken(claims token.AuthClaims) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}

	return jwtToken.SignedString(privKey)
}

func TestValidatePermissions(t *testing.T) {
	var (
		uid = "uid"
		pid = "pid"
	)

	claims := token.AuthClaims{
		CustomClaims: token.CustomClaims{
			UserID:      uid,
			PartnerID:   pid,
			Permissions: []string{"one", "two", "three"},
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
			Issuer:    "test",
		},
	}

	token, err := generateToken(claims)
	assert.NoError(t, err)

	t.Run("positive_allof", func(t *testing.T) {
		routePermissions := []string{"one", "two", "three"}
		b, err := Validate(token, routePermissions, AllOf)
		if err != nil {
			t.Fatal(err)
		}

		if !b {
			t.Fatal("true expected")
		}
	})

	t.Run("negative_allof", func(t *testing.T) {
		routePermissions := []string{"four", "five"}
		b, err := Validate(token, routePermissions, AllOf)
		if err != nil {
			t.Fatal(err)
		}

		if b {
			t.Fatal("false expected")
		}
	})

	t.Run("positive_anyof", func(t *testing.T) {
		routePermissions := []string{"one", "two", "three"}
		b, err := Validate(token, routePermissions, AnyOf)
		if err != nil {
			t.Fatal(err)
		}

		if !b {
			t.Fatal("true expected")
		}
	})

	t.Run("negative_anyof", func(t *testing.T) {
		routePermissions := []string{"four", "five"}
		b, err := Validate(token, routePermissions, AnyOf)
		if err != nil {
			t.Fatal(err)
		}

		if b {
			t.Fatal("false expected")
		}
	})
}

func TestDecode(t *testing.T) {
	var (
		uid = "uid"
		pid = "pid"
	)

	expectedPermissions := []string{"one"}

	claims := token.AuthClaims{
		CustomClaims: token.CustomClaims{
			UserID:      uid,
			PartnerID:   pid,
			Permissions: expectedPermissions,
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
			Issuer:    "test",
		},
	}

	token, err := generateToken(claims)
	assert.NoError(t, err)

	actualPermissions, err := Decode(token)
	assert.NoError(t, err)
	assert.Equal(t, expectedPermissions, actualPermissions)
}

func TestDecodeClaims(t *testing.T) {
	var (
		uid         = "uid"
		pid         = "pid"
		permissions = []string{"one"}
	)

	claims := token.AuthClaims{
		CustomClaims: token.CustomClaims{
			UserID:      uid,
			PartnerID:   pid,
			Permissions: permissions,
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
			Issuer:    "test",
		},
	}

	token, err := generateToken(claims)
	assert.NoError(t, err)

	actualClaims, err := DecodeClaims(token)
	assert.NoError(t, err)
	assert.Equal(t, claims, actualClaims)

	_, err = DecodeClaims("")
	assert.NotNil(t, err)
}

func TestDecodeClaims_WithNOCClaims(t *testing.T) {
	var (
		uid         = "uid"
		pid         = "pid"
		permissions = []string{"one"}
	)

	claims := token.AuthClaims{
		CustomClaims: token.CustomClaims{
			UserID:      uid,
			PartnerID:   pid,
			Permissions: permissions,
			IsNoc:       new(bool),
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
			Issuer:    "test",
		},
	}

	token, err := generateToken(claims)
	assert.NoError(t, err)

	actualClaims, err := DecodeClaims(token)
	assert.NoError(t, err)
	assert.Equal(t, claims, actualClaims)

	_, err = DecodeClaims("")
	assert.NotNil(t, err)
}
