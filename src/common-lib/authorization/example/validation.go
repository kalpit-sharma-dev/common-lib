package example

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/middleware"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token/signature"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

func NewRouter(logger logger.Log) *mux.Router {
	var (
		router = mux.NewRouter()

		// init validation middleware dependencies
		authorizationCfg = token.AuthorizationConfig{
			URL:    "http://internal-intauthzservice-1540055544.ap-south-1.elb.amazonaws.com/authorizationservice/v1",
			Client: http.DefaultClient,
		}
		sigValidator = signature.NewValidator(authorizationCfg, logger)
		jwtExchanger = token.NewJWTExchanger(authorizationCfg, logger)

		// init JWT validation middleware
		validationMW = middleware.NewTokenValidation(authorizationCfg, jwtExchanger, sigValidator, nil, nil, logger)
	)

	// registration of jwt validation middleware by wrapping given http.HandlerFunc
	router.HandleFunc("/", validationMW.Handler(Handler)).Methods(http.MethodGet)
	return router
}

func Handler(w http.ResponseWriter, r *http.Request) {}
