package with

import (
	"context"
	"fmt"
	"runtime/debug"
)

// Context - Helper function to Execute function with context, so that GO-routine creation can be avoided in MS
func Context(ctx context.Context, name string, transaction string, fn func() error) error {
	errChan := make(chan error, 1)
	go func(errChan chan error) {
		defer func() {
			if r := recover(); r != nil {
				log().Error(transaction, fmt.Sprintf("%s-contextRecovered", name), "%s-Recovered:%v Stack Trace :: %s", name, r, debug.Stack())
				errChan <- fmt.Errorf("%s-contextRecovered", name)
			}
		}()
		errChan <- fn()
	}(errChan)

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
