package main

import (
	"fmt"
	"os"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/checksum"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http/client"

	"github.com/cavaliercoder/grab"
)

func main() {
	count := 1
	for {
		fmt.Println("main : Retry number ", count)
		count++
		client := createClient()
		req, _ := createRequest(false)

		// start download
		fmt.Printf("Downloading %v...\n", req.URL())
		resp := client.Do(req)
		if resp != nil && resp.HTTPResponse != nil {
			fmt.Printf("%v\n", resp.HTTPResponse.Status)
		}

		// check for errors
		if err := handleResponse(resp); err != nil {
			fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
			continue
		}

		err := validate(client, resp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "validate failed: %v\n", err)
			continue
		}

		fmt.Printf("Download saved to %v \n", resp.Filename)

		break
	}
}

func validate(client *grab.Client, res *grab.Response) error {
	fmt.Println("validate : validating checksum")
	r, _ := createRequest(true)
	r.NoStore = true
	resp := client.Do(r)
	if resp != nil && resp.HTTPResponse != nil {
		fmt.Printf("%v\n", resp.HTTPResponse.Status)
	}

	fmt.Println("validate : Got Response")

	sum, err := resp.Bytes()
	if err != nil {
		return err
	}

	fmt.Println("validate : Got Response ", string(sum))

	service, err := checksum.GetService(checksum.MD5)
	if err != nil {
		return err
	}

	file, err := res.Open() //os.Open(res.Filename)
	if err != nil {
		return err
	}
	fmt.Println("validate : Opening File ")

	defer file.Close() // nolint
	_, err = service.Validate(file, string(sum))
	if err != nil {
		return err
	}
	return nil
}

func handleResponse(resp *grab.Response) error {
	t := time.NewTicker(400 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			if resp.BytesComplete() != resp.Size() {
				fmt.Printf("  transferred %v / %v bytes (%.2f%%) and Resumed %v\n", resp.BytesComplete(), resp.Size(), 100*resp.Progress(), resp.DidResume)
			}
		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	return resp.Err()
}

func createRequest(checksum bool) (*grab.Request, error) {
	url := "http://cdn.itsupport247.net/InstallJunoAgent/Plugin/Windows/platform-installation-manager/1.0.216/platform_installation_manager_windows32_1.0.216.zip"
	fileName := "/home/juno/Desktop/."
	if checksum {
		url += ".MD5"
		fileName += ".MD5"
	}
	req, err := grab.NewRequest(fileName, url)
	req.NoResume = false
	req.SkipExisting = false
	// req.SetChecksum(md5.New(), []byte(""), false)
	return req, err
}

func createClient() *grab.Client {
	c := client.TLS(&client.Config{Proxy: client.Proxy{Protocol: "http", Address: "localhost", Port: 25}}, true)

	return &grab.Client{
		HTTPClient: c,
		UserAgent:  "downloader",
		BufferSize: 32 * 1024,
	}
}
