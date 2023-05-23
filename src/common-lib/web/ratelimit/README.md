<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Rate limiting

This is the implementation of a strategy for limiting network traffic

## Import Statement

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/ratelimit"
```

## Integration

### Initialize of middleware

```go
if err := ratelimit.Init(context.TODO(), &rateLimitResponder{}, redis.Client(), config.RateLimitConfig); err != nil {
    panic(err)
}

handler.mdw = ratelimit.Middleware()
```

### Use of middleware

Wrap required handlers with the correct parameters

```go
http.HandleFunc(
	"/route", 
	handler.mdw.CheckRateLimit(handler.routeHandler, "unique-group-name", "unique-key"),
)
```

### Responder interface

Responder is an interface to return http response back to the caller. It should be implemented by the user of the middleware to include their own custom localisation and build custom error message if required.

Implementation for the RespondRateLimit handler
```go
func (h *rateLimitResponder) RespondRateLimit(resp ratelimit.Response, r *http.Request, w http.ResponseWriter) {
	
	logger.Get().Error(utils.GetTransactionIDFromRequest(r), "Rate-limit", "%v", resp.Error)
	
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		w.WriteHeader(http.StatusTooManyRequests)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
```

### Configuration

Rate limiting contains next configuration:

- **enabled** - boolean, enable/disable limitation
- **intervalInSec** - time period of one bucket in seconds (related to all groups)
- **inMemoryCacheTTL** - in seconds (in-memory cache to store the configuration)
- **algorithm** - type of algorithm used to calculate the number of requests per time interval (currently only slidingWindow is available)
- **groups** - limits configuration for target groups

There are next limitations:

**intervalInSec**
- min - 30 seconds
- max - 1 day (86 400 seconds)

**inMemoryCacheTTL**
- min - 5 seconds
- default value - 60 seconds

Also, there is an ability to add specific "limit" to "overrides" map for group based on unique endpoint key.

```json
{
  "enabled": true,
  "intervalInSec": 300,
  "inMemoryCacheTTL": 120,
  "algorithm": "slidingWindow",
  "groups": {
    "unique-group-name": {
      "limit": 500,
      "overrides": {
        "unique-key": 1000
      }
    },
    "bulk-action": {
      "limit": 100
    }
  }
}
```

**Get the configuration**

You can use the GetConfig handler to get the current limiter configuration

```go
http.HandleFunc("/rate-limit-config", ratelimit.GetConfig)
```

**Update the configuration**

You can use the SetConfig handler to update the limiter configuration.
Send JSON with the new configuration to the defined route.

```go
http.HandleFunc("/update-rate-limit-config", ratelimit.SetConfig)
```

**Enable/Disable limiting**

You can use the SetEnabled handler to update the limiter status.
Send JSON to the defined route.

```json
{
  "enabled": true
}
```

```go
http.HandleFunc("/update-rate-limit-config/enabled", ratelimit.SetEnabled)
```
