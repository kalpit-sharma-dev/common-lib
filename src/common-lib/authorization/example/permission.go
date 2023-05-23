package example

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/auth"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/middleware"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token/permission"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

var (
	GetHandlerPermissions  = []string{"Read"}
	PostHandlerPermissions = []string{"Read", "Write"}
)

// New creates a router for URL-to-service mapping
func New(logger logger.Log) *mux.Router {
	router := mux.NewRouter()

	// init of permission middleware
	permissionMW := middleware.NewPermission(nil, logger)

	// register permission middleware by wrapping a required http.HandlerFunc, using specified permissions list for the route

	// registration of permission Assert middleware for GET
	// at least one of the caller permissions should match to the GetHandlerPermissions list using validationType `AnyOf`
	router.HandleFunc("/get", permissionMW.AssertHandler(GetHandler, GetHandlerPermissions, permission.AnyOf)).Methods(http.MethodGet)

	// registration of permission Assert middleware for POST
	// all of the caller permissions should match to the PostHandlerPermissions list using validationType `AllOf`
	router.HandleFunc("/post", permissionMW.AssertHandler(PostHandler, PostHandlerPermissions, permission.AllOf)).Methods(http.MethodPost)

	// registration of permissions Decode middleware for SpecificPermissionHandler
	// caller permissions will be passed to the request context and then can easily be retrieved and used in the custom handler logic
	router.HandleFunc("/", permissionMW.DecodeHandler(SpecificPermissionHandler)).Methods(http.MethodDelete)

	return router
}

func GetHandler(w http.ResponseWriter, r *http.Request) {}

func PostHandler(w http.ResponseWriter, r *http.Request) {}

func SpecificPermissionHandler(w http.ResponseWriter, r *http.Request) {
	permissions := r.Context().Value(auth.PermissionKey)
	// ...
	// some logic that works with permissions
	// ...
	fmt.Println(permissions)
}
