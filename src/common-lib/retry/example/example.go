package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	retry "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/retry"
)

type employee struct {
	Status string `json:"status"`
	Data   []struct {
		ID             string `json:"id"`
		EmployeeName   string `json:"employee_name"`
		EmployeeSalary string `json:"employee_salary"`
		EmployeeAge    string `json:"employee_age"`
		ProfileImage   string `json:"profile_image"`
	} `json:"data"`
}

func main() {
	retryExample()
}

func retryExample() {
	url := "http://dummy.restapiexample.com/api/v1/employees"
	var body []byte
	empList := employee{}

	// this variable and its configuration should be initialized in some central location and used every time we retry
	// you can change any attribute of this variable as per your needs, some changes have been demoed below; refer documentation for more details
	task := retry.NewTask()
	task.Attempts = 5
	task.BaseDelay = 20 * time.Second
	task.DelayType = retry.ExponentialDelay
	task.CheckIfRetry = func(err error) bool {
		//set this func if you wish to explicitly retry only in certain error and not retry in case of irrecoverable error
		return true
	}

	err := task.Do(
		func() error {
			resp, err := http.Get(url)

			if err == nil {
				defer func() {
					if err := resp.Body.Close(); err != nil {
						panic(err)
					}
				}()
				body, err = ioutil.ReadAll(resp.Body)
			}

			return err
		},
		"abcd")

	err = json.Unmarshal(body, &empList)
	if err != nil {
		fmt.Printf("failed to unmarshal response err=[%s]\n", err)
	}

	for ct, eachEmp := range empList.Data {
		fmt.Printf("rownum =%v\n", ct)
		fmt.Printf("ID     =%s\n", eachEmp.ID)
		fmt.Printf("Name   =%s\n", eachEmp.EmployeeName)
		fmt.Printf("Age    =%s\n", eachEmp.EmployeeAge)
		fmt.Printf("Salary =%s\n", eachEmp.EmployeeSalary)
		fmt.Println("-----------------------------------------")
	}

}
