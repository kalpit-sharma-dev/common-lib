package main

import (
	"fmt"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/checksum"
	"os"
)

// go run example.go 'test.txt' 'SHA256'
func main() {
	//Setup File Stream
	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Unable to open file")
		return
	}
	defer file.Close()

	//Setup CheckSum Service
	checkSumType := checksum.GetType(os.Args[2])
	checkSumService, err := checksum.GetService(checkSumType)
	if err != nil {
		fmt.Println("Unable to get CheckSum Service")
		return
	}

	//Create CheckSum For File
	checkSum, err := checkSumService.Calculate(file)
	if err != nil {
		fmt.Println("Unable to calculate checksum of file")
		return
	}
	fmt.Println("The checkSum for the file is " + checkSum)

	//Recreate the stream
	file.Close()
	file, err = os.Open(filePath)
	if err != nil {
		fmt.Println("Unable to open file")
		return
	}

	//Validate CheckSum For File
	_, err = checkSumService.Validate(file, checkSum)
	if err != nil {
		fmt.Println("Unable to validate the checkSum")
		return
	}
	fmt.Println("The checkSum for the file was succesfully validated")
}
