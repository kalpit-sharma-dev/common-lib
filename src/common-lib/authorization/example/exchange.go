package example

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

// ExampleGetTokenForServiceToServiceCommunication demonstrates how to obtain JWT token for Service-to-service communication.
func ExampleGetTokenForServiceToServiceCommunication() {
	ctx := context.Background()

	log, err := logger.Create(logger.Config{})
	if err != nil {
		fmt.Println("Error creating logger:", err)
		return
	}

	authorizationCfg := token.AuthorizationConfig{
		URL:    "http://internal-intauthzservice-1540055544.ap-south-1.elb.amazonaws.com/authorizationservice/v1",
		Client: http.DefaultClient,
	}

	assumeRoleReq := token.AssumeRoleRequest{
		RoleName: "Admin", // Prototype role name
	}

	jwtExchanger := token.NewJWTExchanger(authorizationCfg, log)
	gwtToken, err := jwtExchanger.AssumeRole(ctx, assumeRoleReq)
	if err != nil {
		fmt.Println("Error obtaining token:", err)
		return
	}

	fmt.Printf("obtained JWT Token: %q", gwtToken)
}

// ExampleGetEnhancedJWT demonstrates how to obtain enhanced JWT token
func ExampleGetEnhancedJWT() {
	ctx := context.Background()

	log, err := logger.Create(logger.Config{})
	if err != nil {
		fmt.Println("Error creating logger:", err)
		return
	}

	sessionToken := "mock-session-token"
	authorizationCfg := token.AuthorizationConfig{
		URL:    "http://internal-intauthzservice-1540055544.ap-south-1.elb.amazonaws.com/authorizationservice/v1",
		Client: http.DefaultClient,
	}

	jwtExchanger := token.NewJWTExchanger(authorizationCfg, log)
	jwt, err := jwtExchanger.ExchangeEnhanced(ctx, sessionToken, "partnerID")
	if err != nil {
		fmt.Println("Error obtaining enhanced token:", err)
		return
	}

	fmt.Printf("obtained enhanced JWT: %q", jwt)
}
