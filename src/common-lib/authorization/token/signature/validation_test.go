package signature

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token"
)

func TestSigValidator_Validate(t *testing.T) {
	var (
		cfg       = token.AuthorizationConfig{URL: "http://127.0.0.1", Client: http.DefaultClient}
		userID    = "uid"
		partnerID = "pid"
		kid       = "kid"

		coldStartKID = "cold"

		target         = NewValidator(cfg, nil)
		trustingTarget = NewTrustingValidator(cfg, nil)
	)

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	bb, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	pubKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: bb,
	})

	httpmock.Activate()
	defer httpmock.Deactivate()

	httpmock.RegisterResponder(
		http.MethodGet,
		fmt.Sprintf(authorizationPubKeyURLFmt, cfg.URL, kid),
		httpmock.NewJsonResponderOrPanic(http.StatusOK, struct {
			K string `json:"public_key"`
		}{K: string(pubKey)}),
	)

	coldStartConcurrency := 1000
	httpmock.RegisterResponder(
		http.MethodGet,
		fmt.Sprintf(authorizationPubKeyURLFmt, cfg.URL, coldStartKID),
		pauseUntilConcurrency(coldStartConcurrency, httpmock.NewJsonResponderOrPanic(http.StatusOK, struct {
			K string `json:"public_key"`
		}{K: string(pubKey)})),
	)

	claims := token.AuthClaims{
		CustomClaims: token.CustomClaims{
			UserID:    userID,
			PartnerID: partnerID,
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
			Issuer:    "test",
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	t.Run("positive", func(t *testing.T) {
		jwtToken.Header["kid"] = kid
		defer delete(jwtToken.Header, "kid")

		ss, err := jwtToken.SignedString(privKey)
		if err != nil {
			t.Fatal(err)
		}

		err = target.Validate(context.Background(), ss, partnerID, userID)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("positive_on_trusted_request_without_partner_id_or_user_id", func(t *testing.T) {
		jwtToken.Header["kid"] = kid
		defer delete(jwtToken.Header, "kid")

		ss, err := jwtToken.SignedString(privKey)
		if err != nil {
			t.Fatal(err)
		}

		err = trustingTarget.Validate(context.Background(), ss, "", "")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("negative_wrong_partner_id_and_user_id", func(t *testing.T) {
		jwtToken.Header["kid"] = kid
		defer delete(jwtToken.Header, "kid")

		ss, err := jwtToken.SignedString(privKey)
		if err != nil {
			t.Fatal(err)
		}

		err = target.Validate(context.Background(), ss, "", "")
		if err == nil {
			t.Fatal(err)
		}
	})

	t.Run("negative_invalid_identity", func(t *testing.T) {
		jwtToken.Header["kid"] = kid
		defer delete(jwtToken.Header, "kid")

		ss, err := jwtToken.SignedString(privKey)
		if err != nil {
			t.Fatal(err)
		}

		err = target.Validate(context.Background(), ss, "", "")
		if err == nil {
			t.Fatal(err)
		}
	})

	t.Run("negative_invalid_header", func(t *testing.T) {
		ss, err := jwtToken.SignedString(privKey)
		if err != nil {
			t.Fatal(err)
		}

		err = target.Validate(context.Background(), ss, partnerID, userID)
		if err == nil {
			t.Fatal(err)
		}
	})

	t.Run("positive_under_concurrent_writes_during_cold_start", func(t *testing.T) {
		jwtToken.Header["kid"] = coldStartKID
		defer delete(jwtToken.Header, "kid")

		ss, err := jwtToken.SignedString(privKey)
		require.NoError(t, err)
		wg := &sync.WaitGroup{}
		validate := func() {
			assert.NoError(t, trustingTarget.Validate(context.Background(), ss, "", ""))
			wg.Done()
		}
		for i := 0; i < coldStartConcurrency; i++ {
			wg.Add(1)
			go validate()
		}
		wg.Wait()
	})
}

func TestLocalSignatureValidator_ParseToken(t *testing.T) {
	var (
		cfg    = token.AuthorizationConfig{URL: "http://127.0.0.1", Client: http.DefaultClient}
		target = NewValidator(cfg, nil)
	)

	t.Run("positive", func(t *testing.T) {
		privKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		rawPublicKey, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
		require.NoError(t, err)

		pubKey := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: rawPublicKey,
		})

		httpmock.Activate()
		defer httpmock.Deactivate()

		httpmock.RegisterResponder(
			http.MethodGet,
			fmt.Sprintf(authorizationPubKeyURLFmt, cfg.URL, "kid"),
			httpmock.NewJsonResponderOrPanic(http.StatusOK, struct {
				K string `json:"public_key"`
			}{K: string(pubKey)}),
		)

		claims := token.AuthClaims{
			CustomClaims: token.CustomClaims{
				UserID:    "uid",
				PartnerID: "pid",
			},
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
				Issuer:    "test",
			},
		}

		jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		jwtToken.Header["kid"] = "kid"

		tkn, err := jwtToken.SignedString(privKey)
		require.NoError(t, err)

		res, err := target.Parse(context.Background(), fmt.Sprintf("Bearer %s", tkn))
		assert.NoError(t, err)
		assert.Equal(t, &claims.CustomClaims, res)
	})

	t.Run("negative_invalid_bearer_token", func(t *testing.T) {
		res, err := target.Parse(context.Background(), "Bearer")
		assert.EqualError(t, err, errInvalidBearerToken.Error())
		assert.Nil(t, res)
	})

	t.Run("negative_invalid_token", func(t *testing.T) {
		res, err := target.Parse(context.Background(), "AQIC5wM2LY4SfcySUwVGYiBG5DKD7gMpeAPFBS_YFsqQPj0.*AAJTSQACMDIAAlNLABQtMjM3NDU0MTc2MTE1NTAzMDc3OAACUzEAAjAx*")
		assert.EqualError(t, err, fmt.Errorf("token contains an invalid number of segments").Error())
		assert.Nil(t, res)
	})
}
