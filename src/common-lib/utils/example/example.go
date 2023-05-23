package main

import (
	"fmt"
	"net/http"
	"os"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

// go run example.go test
// go run example.go 1
// go run example.go "Diskfull Error with FileSystem."
func main() {
	input := os.Args[1]
	fmt.Println("Raw Input:" + input)

	inputAsString := utils.ToString(input)
	fmt.Println("Input Parsed as String:", inputAsString)

	//This basically just checks if the input is a time which it won't be so it returns 0
	//Note: It does not auto implicitly convert the object to a time
	inputAsTime := utils.ToTime(input)
	fmt.Println("Input Parsed as Time:", inputAsTime)

	//This basically just checks if the input is a int64 which it won't be so it returns 0
	//Note: It does not auto implicitly convert the object to a int64
	inputAsInt64 := utils.ToInt64(input)
	fmt.Println("Input Parsed as int64:", inputAsInt64)

	//This basically just checks if the input is a int which it won't be so it returns 0
	//Note: It does not auto implicitly convert the object to a int
	inputAsInt := utils.ToInt(input)
	fmt.Println("Input Parsed as int:", inputAsInt)

	//This basically just checks if the input is a float64 which it won't be so it returns 0
	//Note: It does not auto implicitly convert the object to a float64
	inputAsFloat64 := utils.ToFloat64(input)
	fmt.Println("Input Parsed as float64:", inputAsFloat64)

	//This basically just checks if the input is a bool which it won't be so it returns false
	//Note: It does not auto implicitly convert the object to a bool
	inputAsBool := utils.ToBool(input)
	fmt.Println("Input Parsed as bool:", inputAsBool)

	//This basically just checks if the input is a string array which it won't be so it returns an empty array
	//Note: It does not auto implicitly convert the object to a string array
	inputAsStringArray := utils.ToStringArray(input)
	fmt.Println("Input Parsed as stringArray:", inputAsStringArray)

	//This basically just checks if the input is a string map which it isnt so it returns an empty map
	//Note: It does not auto implicitly convert the object to a string map
	inputAsStringMap := utils.ToStringMap(input)
	fmt.Println("Input Parsed as stringMap:", inputAsStringMap)

	//Getting a new Transaction ID
	transactionId := utils.GetTransactionID()
	fmt.Println("Newly Generated transaction ID:", transactionId)

	//Getting Transaction ID from request with no X-Request-Id Header
	request, _ := http.NewRequest(http.MethodGet, "fakeUrl.com?userInput="+inputAsString, nil)
	transactionIdFromRequest := utils.GetTransactionIDFromRequest(request)
	fmt.Println("Request had no transaction ID header so it generated a new one:", transactionIdFromRequest)

	//Getting Transaction ID from request with a X-Request-Id Header
	request.Header.Set("X-Request-Id", transactionId)
	transactionIdFromRequest = utils.GetTransactionIDFromRequest(request)
	fmt.Println("Request had a transaction ID header so it grabbed that one:", transactionIdFromRequest)

	//Getting Value of a specific header from request
	request.Header.Set("X-User-Input", inputAsString)
	userInputHeaderFromRequest := utils.GetValueFromRequestHeader(request, "X-User-Input")
	fmt.Println("Request User Input Header:", userInputHeaderFromRequest)

	//Getting specific Query Value fom request
	userInputQueryStringFromRequest := utils.GetQueryValuesFromRequest(request, "userInput")
	fmt.Println("Request User Input Query String:", userInputQueryStringFromRequest)

	//Getting Transaction ID from response with no X-Request-Id Header
	response := &http.Response{
		Header: make(http.Header),
	}
	transactionIdFromResponse := utils.GetTransactionIDFromResponse(response)
	fmt.Println("Response had no transaction ID header so it returned nothing:", transactionIdFromResponse)

	//Getting Transaction ID from response with a X-Request-Id Header
	response.Header.Set("X-Request-Id", transactionId)
	transactionIdFromResponse = utils.GetTransactionIDFromResponse(response)
	fmt.Println("Response had a transaction ID header so it grabbed that one:", transactionIdFromResponse)

	//Calculated Checksum for input
	inputAsByteArray := []byte(input)
	checkSum := utils.GetChecksum(inputAsByteArray)
	fmt.Println("CheckSum of Input:", checkSum)

	//Validate Input to make sure it's valid
	inputAsByteArray = []byte(input)
	succesfulValidation, checkSum := utils.ValidateMessage(inputAsByteArray, checkSum)
	if succesfulValidation {
		fmt.Println("Input validation was succesful")
	} else {
		fmt.Println("Input validation was unsuccesful. CheckSum was calculated as:", checkSum)
	}

	//Determines error response by parsing string
	mainError, subError := utils.DetermineErrorCodePair(inputAsString)
	fmt.Println("\nInput parsed for Errors \nMain Error:", mainError, "\nSub Error:", subError)

	//Determines whether two string arrays are equal case insensitive
	inputArray := []string{input}
	testArray := []string{"TEST"}
	arraysEqual := utils.EqualFold(inputArray, testArray)
	if arraysEqual {
		fmt.Println("\nThe two arrays are equal ignoring case")
	} else {
		fmt.Println("\nThe two arrays are different ignorning case")
	}

	//Determines the difference between two objects
	differenceBetweenInputAndArray := utils.Difference(input, testArray)
	fmt.Println("the changeset between", input, "and", testArray, "is", differenceBetweenInputAndArray)
}
