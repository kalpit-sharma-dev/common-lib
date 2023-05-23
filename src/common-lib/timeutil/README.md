<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# TimeUtil

Helper utility functions to retrieve time in varying formats and at varying locations.

### [Example](example/example.go)  

**Import Statement**

```go
import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/timeutil"
```

## Retrieving time by location

**CurrentTime By Timezone**

```go
currentTime,err:=timeutil.GetCurrentTime("timezone")
// outputs reference to currentTime struct with the following fields

// in case provided timezone is incorrect it would output error  
// GetCurrentTime: Failed to load Location "timezone"

type CurrentTime struct {
 Time string `json:"Time"`   // timeFormat = "15:04:05"
 Date string `json:"Date"`   // dateFormat = "01/02/2006"
 Day  string `json:"Day"`    // dayFormat  = "Monday"
}
```

**Example's of timezones supported by GetCurrentTime**

- "Europe/Andorra"
- "Asia/Dubai"
- "Asia/Kabul"
- "Europe/Tirane"
- "Asia/Yerevan"
- "Antarctica/Casey"
- "Antarctica/Davis"

## Retrieving time in various formats  

**Time in integer format**

#### Year-Month-Day

```go
timeInIntegerFormat,error:=timeutil.ToLongYYYYMMDD(time)
// outputs the time in int format [yyyymmdd]
// 20210628
```

#### Year-Month-Day-Hour

```go
timeInIntegerFormat,error:=timeutil.ToLongYYYYMMDDHH(time)
// output the time in int format [yyyymmddhh]
// 2021062813
```

**Difference between time's**  

```go
timeInIntegerFormat,error:= timeutil.ToHourLong(timeA,timeB)
// outputs the difference in time between timeA and timeB
```

- If timeB is in the _future_ of timeA  

```go
timeInIntegerFormat,error:= timeutil.ToHourLong(timeA, (timeA + 2 days))
// output [yyyymmdd, yyyymm(dd+1), yyyymm(dd+2)]
```

- If timeB is in the _past_ of timeA  

```go
timeInIntegerFormat,error:= timeutil.ToHourLong(timeA, (timeA - 2 days))
// output []
```
