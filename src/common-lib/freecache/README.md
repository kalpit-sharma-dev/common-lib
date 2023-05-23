<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# freecache

Common lib wrapper module to enable freecache usage

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/freecache"
```

**Functions**

```go
New(cacheSize int) *cache    //Returns the new instance of cache
```


```go
Set(key, value []byte, expireSeconds int) (err error)    //Insert a new value/update existing value in cache
```


```go
Get(key []byte) (value []byte, err error)    //Return the value from freecache by the key or not found error
```

```go
Del(key []byte) (affected bool)    //Deletes an item in the freecache by key and returns true or false if a delete occurred.
```


### Contribution

Any changes in this package should be communicated to Common Frameworks Team.
