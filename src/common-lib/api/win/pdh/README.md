<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# API - Windows - Performance Data Helper

Common lib wrapper module to collect windows performance data using PDH API's

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/api/win/pdh"
```

**Functions**

```go
OpenQuery(datasrc string, userData uint32, phQuery *windows.Handle) (ret uint32)    //OpenQuery Creates a new query that is used to manage the collection of performance data.
```


```go
CloseQuery(phQuery windows.Handle) (ret uint32)    //CloseQuery Closes all counters contained in the specified query, closes all handles related to the query, and frees all memory associated with the query.
```


```go
AddCounter(phQuery windows.Handle, szFullCounterPath string, userData uint32, phCounter *windows.Handle) (ret uint32)    //AddCounter Adds the specified counter to the query.
```

```go
CollectQueryData(phQuery windows.Handle) (ret uint32)    //CollectQueryData Collects the current raw data value for all counters in the specified query and updates the status code of each counter.
```

```go
GetFormattedCounterValueDouble(phCounter windows.Handle, counterType *uint32, pValue *FmtCounterValueDouble) (ret uint32)    //GetFormattedCounterValueDouble get the value
```

```go
GetFormattedCounterArrayDouble(phCounter windows.Handle, bufferSize *uint32, bufferCount *uint32, pValue *FmtCounterValueItemDouble) (ret uint32)    //GetFormattedCounterArrayDouble gets all the instance values in double
```

```go
GetFormattedCounterValueLarge(phCounter windows.Handle, counterType *uint32, pValue *FmtCounterValueLarge) (ret uint32)    //GetFormattedCounterValueLarge get the value in Large
```

```go
GetFormattedCounterArrayLarge(phCounter windows.Handle, bufferSize *uint32, bufferCount *uint32, pValue *FmtCounterValueItemLarge) (ret uint32)    //GetFormattedCounterArrayLarge gets all the instance values in large
```

```go
RemoveCounter(phCounter windows.Handle) (ret uint32)    //RemoveCounter removes the counter
```

```go
ParseCounterPath(FullPathBuffer string, pCounterPathElements *CounterPathElements, bufferSize *uint32, dwFlags uint32) (ret uint32)    //ParseCounterPath parses the given perf path
```

```go
QueryPerformanceData(perfPath []PerfPaths) (perfData []PerfData, ret uint32)    //QueryPerformanceData gets the given performance data
```

```go
GetCountersAndInstances(className string) (counters []string, instances []string, err error)    //GetCountersAndInstances gives counter and instance name list
```

### Contribution

Any changes in this package should be communicated to Common Frameworks Team.
