package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient"
)

func main() {

	TestWebClientInvalidHTTPMethod()
	TestWebClientEmptyContentType()
	TestWebClientNilData()
	TestWebClientInvalidMessageType()
	TestWebClientEmptyURLSuffix()
	TestWebClientInvalidURLSuffix()
	TestWebClientDataPost()
}

const (
	brokerURL = "http://localhost:8081"
)

func TestWebClientInvalidHTTPMethod() bool {

	message := &http.Request{}
	message.Method = "POST"
	message.Header.Add("Content-Type", "json/string")
	message.Body = ioutil.NopCloser(bytes.NewReader([]byte(`{"message":"Hello World"}`)))
	url, _ := url.Parse(brokerURL + "/broker/post")
	message.URL = url

	httpCommandFactory := new(webClient.HTTPClientFactoryImpl)
	httpClientFactory := new(webClient.ClientFactoryImpl)
	client := httpClientFactory.GetClientService(httpCommandFactory, webClient.ClientConfig{})

	_, err := client.Do(message)
	if err != nil {
		fmt.Println(err)
		fmt.Println("InvalidHTTPMethod Validated Successfully")
		return true
	}
	fmt.Println("InvalidHTTPMethod Validation failed")
	return false
}

func TestWebClientEmptyContentType() bool {
	message := &http.Request{}
	message.Method = "POST"
	message.Header.Add("Content-Type", "")
	message.Body = ioutil.NopCloser(bytes.NewReader([]byte(`{"message":"Hello World"}`)))
	url, _ := url.Parse(brokerURL + "/broker/post")
	message.URL = url

	httpCommandFactory := new(webClient.HTTPClientFactoryImpl)
	httpClientFactory := new(webClient.ClientFactoryImpl)
	client := httpClientFactory.GetClientService(httpCommandFactory, webClient.ClientConfig{})

	_, err := client.Do(message)
	if err != nil {
		fmt.Println(err)
		fmt.Println("EmptyContentType Validated Successfully")
		return true
	}
	fmt.Println("EmptyContentType Validation failed")
	return false
}

func TestWebClientNilData() bool {
	message := &http.Request{}
	message.Method = "POST"
	message.Header.Add("Content-Type", "json/string")
	message.Body = ioutil.NopCloser(bytes.NewReader(nil))
	url, _ := url.Parse(brokerURL + "/broker/post")
	message.URL = url

	httpCommandFactory := new(webClient.HTTPClientFactoryImpl)
	httpClientFactory := new(webClient.ClientFactoryImpl)
	client := httpClientFactory.GetClientService(httpCommandFactory, webClient.ClientConfig{})

	_, err := client.Do(message)
	if err != nil {
		fmt.Println(err)
		fmt.Println("NilData Successfully")
		return true
	}
	fmt.Println("NilData Successfully")
	return false
}

func TestWebClientInvalidMessageType() bool {
	message := &http.Request{}
	message.Method = "POST"
	message.Header.Add("Content-Type", "json/string")
	message.Body = ioutil.NopCloser(bytes.NewReader([]byte(`{"message":"Hello World"}`)))
	//message.MessageType = 0
	url, _ := url.Parse(brokerURL + "/broker/post")
	message.URL = url

	httpCommandFactory := new(webClient.HTTPClientFactoryImpl)
	httpClientFactory := new(webClient.ClientFactoryImpl)
	client := httpClientFactory.GetClientService(httpCommandFactory, webClient.ClientConfig{})

	_, err := client.Do(message)
	if err != nil {
		fmt.Println(err)
		fmt.Println("InvalidMessageType Validated Successfully")
		return true
	}
	fmt.Println("InvalidMessageType Validation failed")
	return false
}

func TestWebClientEmptyURLSuffix() bool {
	message := &http.Request{}
	message.Method = "POST"
	message.Header.Add("Content-Type", "json/string")
	message.Body = ioutil.NopCloser(bytes.NewReader([]byte(`{"message":"Hello World"}`)))
	url, _ := url.Parse(brokerURL + "")
	message.URL = url

	httpCommandFactory := new(webClient.HTTPClientFactoryImpl)
	httpClientFactory := new(webClient.ClientFactoryImpl)
	client := httpClientFactory.GetClientService(httpCommandFactory, webClient.ClientConfig{})

	_, err := client.Do(message)
	if err != nil {
		fmt.Println(err)
		fmt.Println("EmptyURLSuffix Validated Successfully")
		return true
	}
	fmt.Println("EmptyURLSuffix Validation failed")
	return false
}

func TestWebClientInvalidURLSuffix() bool {
	message := &http.Request{}
	message.Method = "POST"
	message.Header.Add("Content-Type", "json/string")
	message.Body = ioutil.NopCloser(bytes.NewReader([]byte(`{"message":"Hello World"}`)))
	url, _ := url.Parse(brokerURL + "/Invalid/Route")
	message.URL = url

	httpCommandFactory := new(webClient.HTTPClientFactoryImpl)
	httpClientFactory := new(webClient.ClientFactoryImpl)
	client := httpClientFactory.GetClientService(httpCommandFactory, webClient.ClientConfig{})

	resp, err := client.Do(message)
	if err != nil {
		fmt.Println("InvalidURLSuffix Validation failed")
		fmt.Println(err)
		return false
	}
	if resp.StatusCode == 404 {
		fmt.Println("InvalidURLSuffix Validated Successfully")
		return true
	}
	fmt.Println("InvalidURLSuffix Validation failed")
	return false
}

func TestWebClientDataPost() bool {
	message := &http.Request{}
	message.Method = "POST"
	message.Header.Add("Content-Type", "json/string")
	message.Body = ioutil.NopCloser(bytes.NewReader([]byte(`{"message":"Hello World"}`)))
	url, _ := url.Parse(brokerURL + "/broker/post")
	message.URL = url

	httpCommandFactory := new(webClient.HTTPClientFactoryImpl)
	httpClientFactory := new(webClient.ClientFactoryImpl)
	client := httpClientFactory.GetClientService(httpCommandFactory, webClient.ClientConfig{})

	resp, err := client.Do(message)

	if err != nil {
		fmt.Println("DataPost failed")
		fmt.Println(err)
		return false
	}
	if resp.StatusCode != 200 {
		fmt.Println("DataPost failed, server not running")
		fmt.Println(err)
		return false
	}
	fmt.Println("Data Posted Successfully")
	return true
}
