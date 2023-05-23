package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

// ObserveQuery observe cassandra query
func ObserveQuery(ctx context.Context, ob gocql.ObservedQuery) {
	if !TraceEnabled() {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in ObserveQuery Cassandra", r)
		}
	}()
	ctx = checkForParentSegment(ctx)
	ctx, s := BeginSubSegment(ctx, KeyCassandra)
	s.SetStartTime(convertTime(ob.Start))
	s.SetEndTime(convertTime(ob.End))
	AddMetadata(ctx, KeyCassandraStartTime, ob.Start)
	AddMetadata(ctx, KeyCassandraEndTime, ob.End)
	AddAnnotation(ctx, KeyCassandraKeySpace, ob.Keyspace)
	AddMetadata(ctx, KeyCassandraKeySpace, ob.Keyspace)
	AddMetadata(ctx, KeyCassandraType, KeyCassandraTypeNormalQuery)
	AddMetadata(ctx, KeyCassandraStatement, ob.Statement)
	if ob.Host != nil {
		AddAnnotation(ctx, KeyCassandraHost, ob.Host.HostnameAndPort())
		AddMetadata(ctx, KeyCassandraHostInfo, ob.Host.String())
	}
	if ob.Metrics != nil {
		AddMetadata(ctx, KeyCassandraAttempts, ob.Metrics.Attempts)
		AddMetadata(ctx, KeyCassandraTotalLatency, ob.Metrics.TotalLatency)
	}
	if ob.Err != nil {
		AddError(ctx, ob.Err)
	}
	s.Close(ob.Err)
}

// ObserveBatch observe cassandra batch query
func ObserveBatch(ctx context.Context, ob gocql.ObservedBatch) {
	if !TraceEnabled() {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in ObserveBatch Cassandra", r)
		}
	}()
	ctx = checkForParentSegment(ctx)
	ctx, s := BeginSubSegment(ctx, KeyCassandra)
	s.SetStartTime(convertTime(ob.Start))
	s.SetEndTime(convertTime(ob.End))
	AddMetadata(ctx, KeyCassandraStartTime, ob.Start)
	AddMetadata(ctx, KeyCassandraEndTime, ob.End)
	AddAnnotation(ctx, KeyCassandraKeySpace, ob.Keyspace)
	AddMetadata(ctx, KeyCassandraKeySpace, ob.Keyspace)
	AddMetadata(ctx, KeyCassandraType, KeyCassandraTypeBatchQuery)
	AddMetadata(ctx, KeyCassandraStatement, ob.Statements)
	if ob.Host != nil {
		AddAnnotation(ctx, KeyCassandraHost, ob.Host.HostnameAndPort())
		AddMetadata(ctx, KeyCassandraHostInfo, ob.Host.String())
	}
	if ob.Metrics != nil {
		AddMetadata(ctx, KeyCassandraAttempts, ob.Metrics.Attempts)
		AddMetadata(ctx, KeyCassandraTotalLatency, ob.Metrics.TotalLatency)
	}
	if ob.Err != nil {
		AddError(ctx, ob.Err)
	}
	s.Close(ob.Err)
}

//If no segment associated start new Segment
func checkForParentSegment(ctx context.Context) context.Context {
	seg := GetSegment(ctx)
	if seg == nil && ctx != nil {
		ctx, _ = BeginSegment(ctx, NewConfig().ServiceName)
	}
	return ctx
}

func convertTime(t time.Time) float64 {
	return float64(t.UnixNano()) / float64(time.Second)
}
