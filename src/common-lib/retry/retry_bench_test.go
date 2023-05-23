package retry

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkFixedDelayRetry(b *testing.B) {
	attemptNumList := make([]uint, 0)
	attemptErrList := make([]error, 0)
	retryIfErrList := make([]error, 0)
	trialTimestampList := make([]time.Time, 0)

	var countRetryIfCallback uint = 0
	var countTryFailedCallback uint = 0
	var countTaskExecute uint = 0

	tryFailedCallback := func(attemptNum uint, err error) {
		countTryFailedCallback++
		attemptErrList = append(attemptErrList, err)
		attemptNumList = append(attemptNumList, attemptNum)
	}

	retryIfCallback := func(err error) bool {
		countRetryIfCallback++
		retryIfErrList = append(retryIfErrList, err)
		return true
	}

	task := func() error {
		trialTimestampList = append(trialTimestampList, time.Now())
		countTaskExecute++
		return fmt.Errorf("error no - %d", countTaskExecute)
	}

	var attempts uint = 3
	config := NewTask()
	config.Attempts = attempts
	config.DelayType = FixedDelay
	//set lower timeout value to finish the test quicker
	config.BaseDelay = 1 * time.Second
	config.CheckIfRetry = retryIfCallback
	config.ExecuteFunctionOnFail = tryFailedCallback

	err := config.Do(task, "abcd")

	if err == nil {
		fmt.Printf("Retry function must return error when all retries is failed. However it returned nil\n")
		b.Fail()
	}

	if countTaskExecute != attempts {
		fmt.Printf("Count of task execute calls does not match expected. Expected %d, actual %d\n", attempts, countTaskExecute)
		b.Fail()
	}
	if countRetryIfCallback != attempts {
		fmt.Printf("Count of retry if callbacks does not match expected. Expected %d, actual %d\n", attempts, countRetryIfCallback)
		b.Fail()
	}
	if countTryFailedCallback != attempts {
		fmt.Printf("Count of retry failed callbacks does not match expected. Expected %d, actual %d", attempts, countTryFailedCallback)
		b.Fail()
	}

	if uint(len(attemptNumList)) != attempts {
		fmt.Printf("Number of Attempts does not match expected. Expected %d, actual %d\n", attempts, len(attemptNumList))
		b.Fail()
	}
	if uint(len(attemptErrList)) != attempts {
		fmt.Printf("Number of Errors during attempt does not match expected. Expected %d, actual %d\n", attempts, len(attemptErrList))
		b.Fail()
	}
	if uint(len(retryIfErrList)) != attempts {
		fmt.Printf("Number of Errors during retry check does not match expected. Expected %d, actual %d\n", attempts, len(retryIfErrList))
		b.Fail()
	}
	if uint(len(trialTimestampList)) != attempts {
		fmt.Printf("Number of trials does not match expected. Expected %d, actual %d\n", attempts, len(trialTimestampList))
		b.Fail()
	}

	for i := 1; i < len(trialTimestampList); i++ {
		diff := trialTimestampList[i].Sub(trialTimestampList[i-1])
		if uint(diff.Seconds()) != 1 {
			fmt.Printf("duration of retry does not match expected for retry number %d. Expected %d, actual %d\n", i, 1, uint(diff.Seconds()))
			b.Fail()
		}
	}
}

func BenchmarkSecondRetrySuccessful(b *testing.B) {
	attemptNumList := make([]uint, 0)
	attemptErrList := make([]error, 0)
	retryIfErrList := make([]error, 0)
	trialTimestampList := make([]time.Time, 0)

	var countRetryIfCallback uint = 0
	var countTryFailedCallback uint = 0
	var countTaskExecute uint = 0

	tryFailedCallback := func(attemptNum uint, err error) {
		countTryFailedCallback++
		attemptErrList = append(attemptErrList, err)
		attemptNumList = append(attemptNumList, attemptNum)
	}

	retryIfCallback := func(err error) bool {
		countRetryIfCallback++
		retryIfErrList = append(retryIfErrList, err)
		return true
	}

	task := func() error {
		trialTimestampList = append(trialTimestampList, time.Now())
		countTaskExecute++
		if countTaskExecute > 1 {
			return nil
		}
		return fmt.Errorf("error no - %d", countTaskExecute)
	}

	var attempts uint = 3
	config := NewTask()
	config.Attempts = attempts
	config.DelayType = FixedDelay
	//set lower timeout value to finish the test quicker
	config.BaseDelay = 1 * time.Second
	config.CheckIfRetry = retryIfCallback
	config.ExecuteFunctionOnFail = tryFailedCallback

	err := config.Do(task, "abcd")

	if err != nil {
		fmt.Printf("Retry function must not return error when retry is successful, no matter the attempt. However it returned error : %v\n", err)
		b.Fail()
	}

	//task is executed twice and failed once
	if countTaskExecute != 2 {
		fmt.Printf("Count of task execute calls does not match expected. Expected %d, actual %d\n", 2, countTaskExecute)
		b.Fail()
	}

	//task is failed and retried once
	if countRetryIfCallback != 1 {
		fmt.Printf("Count of retry if callbacks does not match expected. Expected %d, actual %d\n", 1, countRetryIfCallback)
		b.Fail()
	}
	//task is failed and retried once
	if countTryFailedCallback != 1 {
		fmt.Printf("Count of retry failed callbacks does not match expected. Expected %d, actual %d", 1, countTryFailedCallback)
		b.Fail()
	}

	//task is failed and retried once
	if uint(len(attemptNumList)) != 1 {
		fmt.Printf("Number of Attempts does not match expected. Expected %d, actual %d\n", 2, len(attemptNumList))
		b.Fail()
	}

	//task is executed twice and failed once
	if uint(len(trialTimestampList)) != 2 {
		fmt.Printf("Number of trials does not match expected. Expected %d, actual %d\n", 2, len(trialTimestampList))
		b.Fail()
	}

	//task is failed and retried once
	if uint(len(attemptErrList)) != 1 {
		fmt.Printf("Number of Errors during attempt does not match expected. Expected %d, actual %d\n", 1, len(attemptErrList))
		b.Fail()
	}
	//task is failed and retried once
	if uint(len(retryIfErrList)) != 1 {
		fmt.Printf("Number of Errors during retry check does not match expected. Expected %d, actual %d\n", 1, len(retryIfErrList))
		b.Fail()
	}

	for i := 1; i < len(trialTimestampList); i++ {
		diff := trialTimestampList[i].Sub(trialTimestampList[i-1])
		if uint(diff.Seconds()) != 1 {
			fmt.Printf("duration of retry does not match expected for retry number %d. Expected %d, actual %d\n", i, 1, uint(diff.Seconds()))
			b.Fail()
		}
	}
}

func BenchmarkRetryDeniedDueToIrrecoverableError(b *testing.B) {
	attemptNumList := make([]uint, 0)
	attemptErrList := make([]error, 0)
	retryIfErrList := make([]error, 0)
	trialTimestampList := make([]time.Time, 0)

	var countRetryIfCallback uint = 0
	var countTryFailedCallback uint = 0
	var countTaskExecute uint = 0

	tryFailedCallback := func(attemptNum uint, err error) {
		countTryFailedCallback++
		attemptErrList = append(attemptErrList, err)
		attemptNumList = append(attemptNumList, attemptNum)
	}

	retryIfCallback := func(err error) bool {
		countRetryIfCallback++
		retryIfErrList = append(retryIfErrList, err)

		if countRetryIfCallback > 1 {
			return false
		}
		return true
	}

	task := func() error {
		trialTimestampList = append(trialTimestampList, time.Now())
		countTaskExecute++
		return fmt.Errorf("error no - %d", countTaskExecute)
	}

	var attempts uint = 3
	config := NewTask()
	config.Attempts = attempts
	config.DelayType = FixedDelay
	//set lower timeout value to finish the test quicker
	config.BaseDelay = 1 * time.Second
	config.CheckIfRetry = retryIfCallback
	config.ExecuteFunctionOnFail = tryFailedCallback

	err := config.Do(task, "abcd")

	if err == nil {
		fmt.Printf("Retry function must return previous attempt error(s) when retry is denied, no matter the attempt. However it returned nil\n")
		b.Fail()
	}

	//task is executed twice
	if countTaskExecute != 2 {
		fmt.Printf("Count of task execute calls does not match expected. Expected %d, actual %d\n", 2, countTaskExecute)
		b.Fail()
	}

	//task is failed checked for retry twice
	if countRetryIfCallback != 2 {
		fmt.Printf("Count of retry if callbacks does not match expected. Expected %d, actual %d\n", 2, countRetryIfCallback)
		b.Fail()
	}
	//task is failed and retried once and second time retry was denied
	if countTryFailedCallback != 1 {
		fmt.Printf("Count of retry failed callbacks does not match expected. Expected %d, actual %d", 1, countTryFailedCallback)
		b.Fail()
	}

	//task is failed and retried once
	if uint(len(attemptNumList)) != 1 {
		fmt.Printf("Number of Attempts does not match expected. Expected %d, actual %d\n", 2, len(attemptNumList))
		b.Fail()
	}
	//task is failed and retried once
	if uint(len(attemptErrList)) != 1 {
		fmt.Printf("Number of Errors during attempt does not match expected. Expected %d, actual %d\n", 1, len(attemptErrList))
		b.Fail()
	}

	//task is executed twice
	if uint(len(trialTimestampList)) != 2 {
		fmt.Printf("Number of trials does not match expected. Expected %d, actual %d\n", 2, len(trialTimestampList))
		b.Fail()
	}

	//task is failed checked for retry twice
	if uint(len(retryIfErrList)) != 2 {
		fmt.Printf("Number of Errors during retry check does not match expected. Expected %d, actual %d\n", 2, len(retryIfErrList))
		b.Fail()
	}

	for i := 1; i < len(trialTimestampList); i++ {
		diff := trialTimestampList[i].Sub(trialTimestampList[i-1])
		if uint(diff.Seconds()) != 1 {
			fmt.Printf("duration of retry does not match expected for retry number %d. Expected %d, actual %d\n", i, 1, uint(diff.Seconds()))
			b.Fail()
		}
	}
}

func BenchmarkExponentialDelayRetry(b *testing.B) {
	attemptNumList := make([]uint, 0)
	attemptErrList := make([]error, 0)
	retryIfErrList := make([]error, 0)
	trialTimestampList := make([]time.Time, 0)

	var countRetryIfCallback uint = 0
	var countTryFailedCallback uint = 0
	var countTaskExecute uint = 0

	tryFailedCallback := func(attemptNum uint, err error) {
		countTryFailedCallback++
		attemptErrList = append(attemptErrList, err)
		attemptNumList = append(attemptNumList, attemptNum)
	}

	retryIfCallback := func(err error) bool {
		countRetryIfCallback++
		retryIfErrList = append(retryIfErrList, err)
		return true
	}

	task := func() error {
		trialTimestampList = append(trialTimestampList, time.Now())
		countTaskExecute++
		return fmt.Errorf("error no - %d", countTaskExecute)
	}

	var attempts uint = 4
	config := NewTask()
	config.Attempts = attempts
	config.DelayType = ExponentialDelay
	//set lower timeout value to finish the test quicker
	config.BaseDelay = 1 * time.Second
	config.CheckIfRetry = retryIfCallback
	config.ExecuteFunctionOnFail = tryFailedCallback

	err := config.Do(task, "abcd")

	if err == nil {
		fmt.Printf("Retry function must return error when all retries is failed. However it returned nil\n")
		b.Fail()
	}

	if countTaskExecute != attempts {
		fmt.Printf("Count of task execute calls does not match expected. Expected %d, actual %d\n", attempts, countTaskExecute)
		b.Fail()
	}
	if countRetryIfCallback != attempts {
		fmt.Printf("Count of retry if callbacks does not match expected. Expected %d, actual %d\n", attempts, countRetryIfCallback)
		b.Fail()
	}
	if countTryFailedCallback != attempts {
		fmt.Printf("Count of retry failed callbacks does not match expected. Expected %d, actual %d", attempts, countTryFailedCallback)
		b.Fail()
	}

	if uint(len(attemptNumList)) != attempts {
		fmt.Printf("Number of Attempts does not match expected. Expected %d, actual %d\n", attempts, len(attemptNumList))
		b.Fail()
	}
	if uint(len(attemptErrList)) != attempts {
		fmt.Printf("Number of Errors during attempt does not match expected. Expected %d, actual %d\n", attempts, len(attemptErrList))
		b.Fail()
	}
	if uint(len(retryIfErrList)) != attempts {
		fmt.Printf("Number of Errors during retry check does not match expected. Expected %d, actual %d\n", attempts, len(retryIfErrList))
		b.Fail()
	}
	if uint(len(trialTimestampList)) != attempts {
		fmt.Printf("Number of trials does not match expected. Expected %d, actual %d\n", attempts, len(trialTimestampList))
		b.Fail()
	}

	firstRetryDiff := trialTimestampList[1].Sub(trialTimestampList[0])
	secondRetryDiff := trialTimestampList[2].Sub(trialTimestampList[1])
	thirdRetryDiff := trialTimestampList[3].Sub(trialTimestampList[2])

	if uint(firstRetryDiff.Seconds()) != 1 {
		fmt.Printf("duration of first retry does not match expected. Expected %d, actual %d\n", 1, uint(firstRetryDiff.Seconds()))
		b.Fail()
	}
	if uint(secondRetryDiff.Seconds()) != 2 {
		fmt.Printf("duration of second retry does not match expected. Expected %d, actual %d\n", 2, uint(secondRetryDiff.Seconds()))
		b.Fail()
	}
	if uint(thirdRetryDiff.Seconds()) != 4 {
		fmt.Printf("duration of third retry does not match expected. Expected %d, actual %d\n", 4, uint(thirdRetryDiff.Seconds()))
		b.Fail()
	}
}
