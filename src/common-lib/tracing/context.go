package tracing

import (
	"context"
	"fmt"

	"github.com/aws/aws-xray-sdk-go/xray"
)

// AddAnnotation adds an annotation to the provided segment or subsegment in ctx
func AddAnnotation(ctx context.Context, key string, value interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()
	err = xray.AddAnnotation(ctx, key, value)
	return err
}

// AddMetadata adds a metadata to the provided segment or subsegment in ctx.
func AddMetadata(ctx context.Context, key string, value interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()
	err = xray.AddMetadata(ctx, key, value)
	return err
}

// AddError adds an error to the provided segment or subsegment in ctx.
func AddError(ctx context.Context, iErr error) (err error) {
	// Panic no longer occurs as error gets appended to cause.Exception.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()
	return xray.AddError(ctx, iErr)
}
