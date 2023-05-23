<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# IS

Helper functions to validate input strings are either Number, Phone Number, Alpha, Alpha Numeric, UUID etc 

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/validate/is"
```

Phone - Validate given string is Phone number or not

```go
is.Phone(s string) bool
```

Email - Validate given string is Email Address or not

```go
is.Email(s string) bool
```

UUID - Validate given string is UUID or not

```go
is.UUID(s string) bool
```

Alpha - Validate given string is Alpha or not

```go
is.Alpha(s string) bool
```

AlphaNumeric - Validate given string is Alpha Numeric or not

```go
is.AlphaNumeric(s string) bool
```

Number - Validate given string is Number or not

```go
is.Number(s string) bool
```

Identifier - Validate given string is Identifier (safe to use in as metric name) or not

```go
is.Identifier(s string) bool
```

**Banchmark**

```
BenchmarkPhone-2          	  100000	     19613 ns/op	       0 B/op	       0 allocs/op
BenchmarkEmail-2          	  200000	     10399 ns/op	       0 B/op	       0 allocs/op
BenchmarkUUID-2           	  200000	      7250 ns/op	       0 B/op	       0 allocs/op
BenchmarkAlpha-2          	  500000	      2898 ns/op	       0 B/op	       0 allocs/op
BenchmarkAlphaNumeric-2   	  500000	      4199 ns/op	       0 B/op	       0 allocs/op
BenchmarkNumber-2         	  200000	      6003 ns/op	       0 B/op	       0 allocs/op
BenchmarkIdentifier-2     	  300000	      5705 ns/op	       0 B/op	       0 allocs/op
```

### Contribution

Any changes in this package should be communicated to Juno Team.
