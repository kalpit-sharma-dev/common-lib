package tracing

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2plugin "github.com/aws/aws-xray-sdk-go/awsplugins/ec2"
	ecsplugin "github.com/aws/aws-xray-sdk-go/awsplugins/ecs"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// Type denotes the tracing type
type Type string

const (
	// KeyTracingEnabled key for tracing status
	KeyTracingEnabled = "TRACING_ENABLED"
	// KeyAwsXraySdkDisabled key for xray sdk disabled
	KeyAwsXraySdkDisabled = "AWS_XRAY_SDK_DISABLED"
	// AwsXrayTracingType denotes aws xray tracing
	AwsXrayTracingType Type = "xray"
	// AwsHostPlatformECS denotes aws ecs host platform
	AwsHostPlatformECS Type = "ecs"
	// AwsHostPlatformEC2 denotes aws ec2 host platform
	AwsHostPlatformEC2 Type = "ec2"
)

// Configure configures xray configuration values
func Configure(config *Config) {
	if config.Enabled && config.Type == AwsXrayTracingType {
		os.Setenv("AWS_XRAY_CONTEXT_MISSING", "LOG_ERROR")
		switch config.HostPlatform {
		case AwsHostPlatformECS:
			ecsplugin.Init()
		case AwsHostPlatformEC2:
			ec2plugin.Init()
		}
		xray.Configure(xray.Config{
			DaemonAddr:     config.Address,
			ServiceVersion: config.ServiceVersion,
		})
	}
	handleTracingStatus(config.Enabled)
}

func handleTracingStatus(isEnabled bool) {
	os.Setenv(KeyTracingEnabled, strconv.FormatBool(isEnabled))
	os.Setenv(KeyAwsXraySdkDisabled, strconv.FormatBool(!isEnabled))
}

// TraceEnabled Check if tracing is enabled
func TraceEnabled() bool {
	return traceEnable()
}

var traceEnable = func() bool {
	enableKey := os.Getenv(KeyTracingEnabled)
	return strings.ToLower(enableKey) == strconv.FormatBool(true)
}

// WrapHandlerWithTracing returns http.Handler with xray integration
func WrapHandlerWithTracing(config *Config, next http.Handler) (_ http.Handler, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()
	if !TraceEnabled() {
		return next, nil
	}
	isEnabled, err := isEnabled(config)
	if err != nil {
		return nil, err
	}
	if isEnabled {
		return xray.Handler(xray.NewFixedSegmentNamer(config.ServiceName), next), err
	}
	return next, err
}

// HTTPClient returns an instrumented http.Client
func HTTPClient(config *Config, c *http.Client) (_ *http.Client, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()
	if !TraceEnabled() {
		return c, nil
	}

	isEnabled, err := isEnabled(config)
	if err != nil {
		return nil, err
	}
	if isEnabled {
		return xray.Client(c), err
	}
	return c, err
}

func isEnabled(config *Config) (bool, error) {
	if config != nil && config.Enabled {
		switch config.Type {
		case AwsXrayTracingType:
			return true, nil
		default:
			return false, fmt.Errorf("tracing type %s is not a valid option", config.Type)
		}
	}
	return false, nil
}

// AWSClient instruments aws client
func AWSClient(cfg *aws.Config) {
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)
}

// Capture traces synchronous function as subsegment
func Capture(ctx context.Context, segmentName string, fn func(context.Context) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in Capture", r)
		}
	}()
	if !TraceEnabled() {
		return fn(ctx)
	}
	return xray.Capture(ctx, segmentName, fn)
}
