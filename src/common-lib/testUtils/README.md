<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# TestUtil

Helper utility functions to ignore time difference's during comparision of two models.

### [Example](example/example.go)  

**Import Statement**

```go
import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/testutil"
```

## Ignore `time.time` fields and compare models

 Copied from reflect.DeepEqual and modified to consider fields of type time.Time as equal even if they contain different time values.

```go
    //Ignores equality for fields of type time.Time
    //Output - True if all fields of A and B are equal except time.time fields
    isEqual = DeepEqualIgnoringTime(A interface{},B interface{});
```

Often useful in unit-tests to compare models created using `time.Now()` which cannot now be compared by `reflect.DeepEqual`.
