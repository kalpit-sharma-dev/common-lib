<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Authorization

Includes a set of middlewares, all required interfaces and it's common implementations for:
 - Gateway services to authorize the user by acquiring and/or validating JWT token
 - Backend services to validate user permissions against accessed route

#### Authorization token (JWT) lifecycle

JWT Token is either passed by the UI or exchanged via the call to Authorization MS by given session token.

JWT lifecycle flows:
1. `UI -> cookie -> gateway -> header -> backend`
2. `gateway (token exchange) -> header -> backend`
3. `service -> Auth MS ("assume role and get JWT token") -> service` for service-to-service communication.

**Note:** The second flow is temporary until the UI will start passing JWT in the request cookie.


## Import Statement

For middlewares:
```go
import	(
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/middleware"
)
```

For JWT acquisition/permission validation implementations to build your own middleware:

- Permissions validation components:
```go
import	(
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token/permission"
)
```
- Signature/claims validation components:
```go
import	(
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token/signature"
)
```
- Token Exchange components:
```go
import	(
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token"
)
```

## Components description and usage

### JWT Validation middleware

This middleware should be used by gateway services to validate user authorization. There are two common use cases which are handled by the middleware:
- JWT is provided in the Authorization header by the caller
- JWT is not provided

#### Dependencies

It depends on the following components:
- signature.Validator
- token.Exchanger
- Cache
- Responder
- Logger
- token.AuthorizationConfig

The only `AuthorizationConfig` and `Logger` are required. For all other components default implementations will be used, in case it is not provided.

#### Usage

##### Prerequisites

To use that middleware the following requirements should be met:
1. During middleware initialization authorization service URL and logger are required parameters that should be passed
2. On the middleware execution phase inside `ServerHTTP()` PartnerID and UserID should be provided in the request context, usually it should be populated by authentication middleware.

##### Registration

1. Create new instance of validation middleware by using `middleware.NewTokenValidation()`
2. Since JWT validation middleware satisfies http.Handler interface, to start using it just register it in your handler stack. Usually this middleware should be executed after authentication.

```go
// validation middleware initialization, where:
// authConfig is a config to make calls to authorization service
// exchanger is to generate JWT token by given session token
// validator is for JWT signature and claims validation
// responder is used to write to http.Response error message
// cache to lower amount of call to Authorization MS and store JWT token in it
// log acts as a logger
validationMW := middleware.NewTokenValidation(authConfig, exchanger, validator, responder, cache, log)

// registration of the validation middleware
router.HandleFunc("/", validationMW.Handler(http.Handler)).Methods(http.MethodGet)
```

If using common lib's `src/web`, you can use `router.Use(validationMW.Middleware)` instead to apply this validation middleware globablly. If you're developing a service that's meant to be platform-compliant, you should use `NewTrustingTokenValidation` instead of `NewTokenValidation`, and omit the exchanger argument. Note that this validator should only be used if your service cannot be directly accessed in production, and must instead go through the reverse proxy or some other gateway that handles authentication.

##### Example

For working example please checkout example/validation.go

### Permission validation middleware

This middleware is for backend services to validate user permissions, parsed from JWT, against accessed route to allow or reject received call.

#### Dependencies

It depends on the following components:
- Responder
- Logger

The only `Logger` component is required.

#### Usage

##### Prerequisites

No specific prerequisites.

##### Registration

The middleware is designed as a handler wrapper, so only required route handlers can be wrapped without having a need to register that middleware separately.

To initialize an instance of permissions middleware, just call `middleware.NewPermission()`. Then call `instance.AssertHandler()` or `instance.DecodeHandler()` and provide required parameters.

The required parameters are the actual handler which is going to be wrapped and the `permission.ValidationType`, see `Types` section for more details.

```go
// init of permission middleware, where:
// responder is used to write to http.Response error message
// log acts as a logger
permissionMW := middleware.NewPermission(responder, log)

// permissions list of the specific handler
handlerPermissions := []string{"Read"}
// registration of permission Assert middleware for GET
// at least one of the caller permissions should match to the handlerPermissions list using validationType `AnyOf`
router.HandleFunc("/get", permissionMW.AssertHandler(GetHandler, handlerPermissions, permission.AnyOf)).Methods(http.MethodGet)
```

If using common lib's `src/web`, you can use something like this instead:

```go
router.AddFunc("/get", permissionMW.CommonAssertHandler(GetHandler, permission.AnyOf, "Read"), http.MethodGet)
```

If you're developing a service that's meant to be platform-compliant, you should use `NewTrustingPermission` instead of `NewPermission`. Note that this permission middleware should only be used if your service cannot be directly accessed in production, and must instead go through the reverse proxy or some other gateway that handles authentication.

##### Example

For working example please checkout `example/permission.go`.

## Types

```go
// PermissionValidationType describes how permissions can be validated
// AnyOf - at least one required permission should be present
// AllOf - all of the permissions should be present in the JWT
permission.ValidationType

// Exchanger is an interface that handles JWT token acquisition by given provided session token
token.Exchanger

// JWTExchanger is a default implementation of `Exchanger`
token.JWTExchanger

// Validator is an interface for JWT validation
signature.Validator

// TokenParser is an interface for bearer token parsing
signature.TokenParser

// LocalSignatureValidator is a local validator of JWT claims and signature
signature.LocalSignatureValidator

// TokenValidation is a JWT validation middleware
middleware.TokenValidation

// Responder is an interface to return http response back to the caller.
// Usually it should be implemented by the user of the middleware to include their own custom localisation and build custom error message if required.
middleware.Responder


```

## Errors

```go
// ErrInvalidAuthzResponse is thrown on the invalid response from authorization service
token.ErrInvalidAuthzResponse
// ErrNoPartnerID as a prerequisite partnerID should be specified in the request context
middleware.ErrNoPartnerID
// ErrNoUserID as a prerequisite userID should be specified in the request context
middleware.ErrNoUserID
// ErrBearerTokenInvalid indicates that JWT token is invalid
middleware.ErrBearerTokenInvalid
// ErrAccessForbidden indicates that user access by provided JWT token is forbidden
middleware.ErrAccessForbidden
```
