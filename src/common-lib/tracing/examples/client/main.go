package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/tracing"
)

func main() {

	// setup tracing config
	tracingConfig := tracing.NewConfig()
	tracingConfig.ServiceName = util.Hostname(util.ProcessName()) + "_client"
	tracingConfig.ServiceVersion = "1.0.0"

	// setup http client
	client, _ := tracing.HTTPClient(tracingConfig, &http.Client{})

	// get context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get xray segment
	ctx, seg := xray.BeginSegment(ctx, "sample-xray-client")

	// add metadata (any value)
	metadata := struct{ Hello string }{Hello: "world"}
	tracing.AddMetadata(ctx, "metadata", metadata)

	// add annotation (string, number, or boolean)
	tracing.AddAnnotation(ctx, "annotation", "hello world")

	// prepare http request
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/webserver1", nil)

	// make request
	res, err := client.Do(req)

	// capture critical code
	tracing.Capture(ctx, "sample-xray-client-subsegment", func(ctx context.Context) error {
		<-time.After(1 * time.Second)
		return nil
	})

	// close segment
	seg.Close(err)

	// parse result
	if err != nil {
		fmt.Println(err)
	} else {
		defer res.Body.Close()
		resData, _ := ioutil.ReadAll(res.Body)
		resStr := string(resData)
		fmt.Printf("webserver1 response: %s \n", resStr)
	}
}
