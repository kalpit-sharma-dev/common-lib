package main

import (
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/timeutil"
)

func main() {
	currentTime := time.Now()
	_, _ = timeutil.GetCurrentTime("Asia/Calcutta")
	// output
	// &{13:21:53 06/28/2021 Monday}
	_, _ = timeutil.ToLongYYYYMMDD(&currentTime)
	// output [yyyymmdd]
	// 20210628
	_, _ = timeutil.ToLongYYYYMMDDHH(&currentTime)
	// output [yyyymmddhh]
	// 2021062813
	_ = timeutil.ToHourLong(currentTime, currentTime.AddDate(0, 0, 2))
	// output [yyyymmdd, yyyymm(dd+1), yyyymm(dd+2)]
	// [20210628 20210629 20210630]
}
