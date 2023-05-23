# contextutil

`contextutil` provides a standard means of storing and accessing common transactional data in our platform within the scope of a single request.

By annotating the Golang [context.Context](https://golang.org/pkg/context/) with this data, and encouraging passing the context through an entire call chain in a program, we can automatically enable our additional libraries to be able to access this metadata about a request.

## Usage

Basic use:

```golang
// create context data with key data, pass "" if not available
ctxData := context.NewContextData(transactionID, partnerID, userID)
// add additional optional data
ctxData.CompanyID = companyID
ctxData.SiteID = siteID
// persist data onto the context
ctx = contextutil.WithValue(ctx, ctxData)

// ensure the ctx value is passed throughout the rest of the program's calls
```

In an HTTP server, as middleware:

```golang
import (
    "github.com/dgrijalva/jwt-go"
    "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/contextutil"
    "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
)

func contextMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        vars := web.GetRequestContext(w, r).GetVars()
        // get transaction ID from incoming header
        transactionID := r.Header.Get("X-Request-Id")

        // get partner ID from API path
        partnerID := vars["partnerId"]

        // get company ID from API path
        companyID := vars["companyId"]

        //get user ID from a decoded JWT authorization header
        authToken := r.Header.Get("Authorization")
        token, err := jwt.ParseWithClaims(authToken, &jwt.StandardClaims{}, 
            func(token *jwt.Token) (interface{}, error) {
                return []byte("<YOUR VERIFICATION KEY>"), nil
            })
        userID = claims.StandardClaims.Subject
        
        ctxData := context.NewContextData(transactionID, partnerID, userID)
        ctxData.CompanyID = companyID
        ctx = contextutil.WithValue(ctx, ctxData)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handle(w http.ResponseWriter, r *http.Request) {
    ctxData := contextutil.GetData(ctx)
    fmt.Printf("Transaction ID is %s", ctxData.TransactionID)
    fmt.Printf("Partner ID is %s", ctxData.PartnerID)
    fmt.Printf("User ID is %s", ctxData.UserID)
    fmt.Printf("Company ID is %s", ctxData.CompanyID)
}

server := web.Create(&web.ServerConfig{ListenURL: ":8080"})
r := server.GetRouter()
r.AddFunc("/", contextMiddleware(handler), http.MethodGet)
r.AddFunc("/partners", contextMiddleware(handler), http.MethodGet)
r.AddFunc("/partners/{partnerId}", contextMiddleware(handler), http.MethodGet)
r.AddFunc("/partners/{partnerId}/companies", contextMiddleware(handler), http.MethodGet)
r.AddFunc("/partners/{partnerId}/companies/{companyId}", contextMiddleware(handler), http.MethodGet)
server.Start(context.Background())
```

In other use cases, such as a `main` function or polling for messages:

```golang
import (
    "fmt"
    "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/contextutil"
)

type Message struct {
    Content string
    PartnerID string
    TransactionID string
}

// ... code to poll messages, such as Kafka ...

func handleMessage(m Message) {
    ctx := ctx.Background()
    ctxData := contextutil.NewContextData(m.TransactionID, m.PartnerID, "")
    ctx = contextutil.WithValue(ctx, ctxData)

    // important to continue passing context through rest of call chain to persist this data
    additionalMessageHandling(ctx, m)
}

func additionalMessageHandling(ctx context.Context, m Message) {
    ctxData := contextutil.GetData(ctx)
    fmt.Printf("Transaction ID is %s", ctxData.TransactionID)
    fmt.Printf("Partner ID is %s", ctxData.PartnerID)
}

```

## Context-Based Programming

The `context.Context` carries many useful functions for deadlines and cancellation across goroutines and other asynchronous behavior, but also can store additional metadata about a request that can then be accessed later in the application.

For a good idea of how the context should be used, from [How to correctly use context.Context in Go 1.7](https://medium.com/@cep21/how-to-correctly-use-context-context-in-go-1-7-8f2c0fafdf39):

> A great mental model of using Context is that it should flow through your program. Imagine a river or running water. This generally means that you don’t want to store it somewhere like in a struct. Nor do you want to keep it around any more than strictly needed. Context should be an interface that is passed from function to function down your call stack, augmented as needed. Ideally, a Context object is created with each request and expires when the request is over.

This means that when writing our application code, we need to keep this in mind and write the functions in a way that it accepts context from call to call.

From the [Golang blog on the usage of context](https://blog.golang.org/context):

> At Google, we require that Go programmers pass a Context parameter as the first argument to every function on the call path between incoming and outgoing requests. This allows Go code developed by many different teams to interoperate well.

To adopt this mentality for us and to incorporate the changes to enable our other libraries to gain the benefits of the context, existing code will need to be reworked to continue to pass a `context` value throughout the chain of calls in the program. New code should design with context in mind.

### Patterns

When updating interfaces that previously did not support context, it is common for many libraries to provide an alternative function that does accept context as a first parameter.

From the [sql](https://golang.org/pkg/database/sql/) package, for example, it has the following functions:

- `func (db *DB) Prepare(query string) (*Stmt, error)`
- `func (db *DB) PrepareContext(ctx context.Context, query string) (*Stmt, error)`
- `func (db *DB) Query(query string, args ...interface{}) (*Rows, error)`
- `func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error)`
- etc

It's important to note that the context should be used to store metadata-like data about the request, and should not be used as an object to assist in passing data that a function needs in order to work or data that would affect business or domain logic.

[Go Code Review Comment Guidelines](https://github.com/golang/go/wiki/CodeReviewComments#contexts) provides some general rules for context usage to keep to, one of them being:

> If you have application data to pass around, put it in a parameter, in the receiver, in globals, or, if it truly belongs there, in a Context value.
