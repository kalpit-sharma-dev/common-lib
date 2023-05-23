package cql

import (
	"context"

	"github.com/gocql/gocql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/tracing"
)

// Observer class for Cassandra notification
type Observer struct {
}

// ObserveQuery observe query
func (q Observer) ObserveQuery(ctx context.Context, ob gocql.ObservedQuery) {
	tracing.ObserveQuery(ctx, ob)
}

// ObserveBatch bbserve batch query
func (q Observer) ObserveBatch(ctx context.Context, ob gocql.ObservedBatch) {
	tracing.ObserveBatch(ctx, ob)
}
