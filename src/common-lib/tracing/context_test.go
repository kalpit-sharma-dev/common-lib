package tracing

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-xray-sdk-go/xray"
)

func TestContext_AddAnnotation(t *testing.T) {
	os.Setenv("AWS_XRAY_SDK_DISABLED", "true")
	tests := []struct {
		name string
		arg  func() error
		want error
	}{
		{
			name: "default",
			arg: func() error {
				ctx, _ := BeginSegment(context.Background(), "default")
				return AddAnnotation(ctx, "key", "value")
			},
			want: nil,
		},
		{
			name: "error",
			arg: func() error {
				return AddAnnotation(context.Background(), "key", "value")
			},
			want: xray.ErrRetrieveSegment,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.arg(); got != tt.want {
				t.Errorf("AddAnnotation = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContext_AddMetadata(t *testing.T) {
	os.Setenv("AWS_XRAY_SDK_DISABLED", "true")
	tests := []struct {
		name string
		arg  func() error
		want error
	}{
		{
			name: "default",
			arg: func() error {
				ctx, _ := BeginSegment(context.Background(), "default")
				return AddMetadata(ctx, "key", "value")
			},
			want: nil,
		},
		{
			name: "error",
			arg: func() error {
				return AddMetadata(context.Background(), "key", "value")
			},
			want: xray.ErrRetrieveSegment,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.arg(); got != tt.want {
				t.Errorf("AddMetadata = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContext_AddError(t *testing.T) {
	tests := []struct {
		name string
		arg  func() error
		want error
	}{
		{
			name: "default",
			arg: func() error {
				// If SDK is disabled errors are not added.
				os.Setenv("AWS_XRAY_SDK_DISABLED", "true")
				ctx, _ := BeginSegment(context.Background(), "default")
				return AddError(ctx, errors.New("default error"))
			},
			want: nil,
		},
		{
			name: "default",
			arg: func() error {
				// Errors are added to segment's exception and AddError function returns nil.
				os.Setenv("AWS_XRAY_SDK_DISABLED", "false")
				ctx, _ := BeginSegment(context.Background(), "default")
				AddError(ctx, errors.New("default error"))
				errorMessage := (xray.GetSegment(ctx).GetCause()).Exceptions[0].Message
				return errors.New(errorMessage)
			},
			want: errors.New("default error"),
		},
		{
			name: "error",
			arg: func() error {
				return AddError(context.Background(), errors.New("error"))
			},
			want: xray.ErrRetrieveSegment,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.arg()
			if tt.want != nil && tt.want.Error() == "default error" {
				if got == nil {
					t.Errorf("AddError = %v, want %v", got, tt.want)
				}
			} else if got != tt.want {
				t.Errorf("AddError = %v, want %v", got, tt.want)
			}
		})
	}
}
