package tracing

import (
	"context"
	"net/http"

	"github.com/aws/aws-xray-sdk-go/xray"
)

// Segment represents a tracing segment
type Segment struct {
	s *xray.Segment
}

// BeginSegment creates a trace segment with a given name
func BeginSegment(ctx context.Context, name string) (context.Context, *Segment) {
	c, s := xray.BeginSegment(ctx, name)
	return c, newSegment(s)
}

// Close a segment
func (s *Segment) Close(err error) {
	s.s.Close(err)
}

// GetStartTime  gives a segment start time
func (s *Segment) GetStartTime() float64 {
	return s.s.StartTime
}

// SetStartTime  set a segment start time
func (s *Segment) SetStartTime(startTime float64) {
	s.s.StartTime = startTime
}

// GetEndTime gives a segment end time
func (s *Segment) GetEndTime() float64 {
	return s.s.EndTime
}

// SetEndTime set a segment end time
func (s *Segment) SetEndTime(endTime float64) {
	s.s.EndTime = endTime
}

// BeginSubSegment creates a trace subsegment with a given name
func BeginSubSegment(ctx context.Context, name string) (context.Context, *Segment) {
	c, s := xray.BeginSubsegment(ctx, name)
	return c, newSegment(s)
}

func newSegment(s *xray.Segment) *Segment {
	return &Segment{s}
}

// NewSegmentFromParent creates a segment for downstream call and add information to the segment that gets from HTTP header.
func NewSegmentFromParent(ctx context.Context, name string, r *http.Request, segParent *Segment) (context.Context, *Segment) {
	ctx, seg := xray.NewSegmentFromHeader(ctx, name, r, segParent.s.DownstreamHeader())
	return ctx, newSegment(seg)
}

// GetSegment returns a pointer to the segment or subsegment provided
// in ctx, or nil if no segment or subsegment is found.
func GetSegment(ctx context.Context) *Segment {
	seg := xray.GetSegment(ctx)
	if seg == nil {
		return nil
	}
	return &Segment{seg}
}
