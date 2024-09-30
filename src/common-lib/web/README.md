<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Light weight wrapper of mux to provide middlware functionality

This is a common implementation for mux implementation which can be used by all the Go projects in the google.
It supports middleware functionalty.

### Third-Party Libraries

- [gorilla/mux](https://github.com/gorilla/mux) 
- **License** [bcd3-clouse] (https://github.com/gorilla/mux/blob/master/LICENSE) 
 - **Description** Package gorilla/mux implements a request router and dispatcher for matching incoming requests to their respective handler.

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
```

**Server Instance**

```go
//GetServer instance of HTTPServer
server := Create(&ServerConfig{ListenURL: ":8080"})
```

**Setup router and start server**

```go
//Get the Router and register route
r1 := server.GetRouter()
r1.AddHandle("/path-to-api", withMiddleware(protectedEndpoint), http.MethodGet)
r1.AddFunc("/path-to-api", withoutMiddleware, http.MethodGet)

//Start the Server
ctx1, _ := context.WithTimeout(context.Background(), 60*time.Second)
server.Start(ctx1)
```

### [Example](example/example.go)


```
Sample Output

juno@juno-VirtualBox:~/SourceCode/Juno/Go/src/gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6$ go run web/example/example.go 
*web.GorillaRouterauthMiddleware1 invoked
protected invoked
protected invoked
```

### Contribution

Any changes in this package should be communicated to Juno Team.
