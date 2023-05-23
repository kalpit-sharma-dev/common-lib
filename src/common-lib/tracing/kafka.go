package tracing

import (
	"context"
)

// CaptureKafkaProduce traces messages published by producer
func CaptureKafkaProduce(ctx context.Context,
	fn func(ctx context.Context) error) error {
	return Capture(ctx, KeyProducerSubSegmentName, fn)
}

// CaptureKafkaConsumer traces messages consumed by consumer
func CaptureKafkaConsumer(ctx context.Context,
	fn func(ctx context.Context) error) error {
	return Capture(ctx, KeyConsumerSubSegmentName, fn)
}
