package consumer

import (
	"reflect"
	"testing"
)

func Test_offsetStatus_String(t *testing.T) {
	type fields struct {
		offset int64
		status int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "1 In-Progress", want: "In-Progress/0", fields: fields{offset: 0, status: inProgress}},
		{name: "2 Completed", want: "Completed/0", fields: fields{offset: 0, status: completed}},

		{name: "1 In-Progress", want: "In-Progress/-1", fields: fields{offset: -1, status: inProgress}},
		{name: "2 Completed", want: "Completed/-1", fields: fields{offset: -1, status: completed}},

		{name: "1 In-Progress", want: "In-Progress/100", fields: fields{offset: 100, status: inProgress}},
		{name: "2 Completed", want: "Completed/100", fields: fields{offset: 100, status: completed}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &offsetStatus{
				offset: tt.fields.offset,
				status: tt.fields.status,
			}
			if got := o.String(); got != tt.want {
				t.Errorf("offsetStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_partitionState_Len(t *testing.T) {
	type fields struct {
		offset             []*offsetStatus
		lastCommitedOffset int64
		dirty              bool
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{name: "1 No value", want: 0},
		{name: "2 Blank", want: 0, fields: fields{offset: []*offsetStatus{}}},
		{name: "3 Size-1", want: 1, fields: fields{offset: []*offsetStatus{{}}}},
		{name: "4 Size-3", want: 3, fields: fields{offset: []*offsetStatus{{}, {}, {}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &partitionState{
				offset:             tt.fields.offset,
				lastCommitedOffset: tt.fields.lastCommitedOffset,
				dirty:              tt.fields.dirty,
			}
			if got := p.Len(); got != tt.want {
				t.Errorf("partitionState.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_partitionState_getCommitOffset(t *testing.T) {
	type fields struct {
		offset             []*offsetStatus
		lastCommitedOffset int64
		dirty              bool
	}
	type args struct {
		offset int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
		want1  bool
	}{
		{
			name: "1 Blank Array", want: -1, want1: false, args: args{offset: 20},
			fields: fields{offset: []*offsetStatus{}, lastCommitedOffset: -1, dirty: false},
		},
		{
			name: "2 Single Message", want: 20, want1: true, args: args{offset: 20},
			fields: fields{offset: []*offsetStatus{{offset: 20, status: inProgress}}, lastCommitedOffset: -1, dirty: false},
		},
		{
			name: "3 Offset Commit having next completed", want: 20, want1: true, args: args{offset: 19},
			fields: fields{offset: []*offsetStatus{{offset: 19, status: inProgress},
				{offset: 20, status: completed}}, lastCommitedOffset: -1, dirty: false},
		},
		{
			name: "4 Offset Commit having previous inprogress", want: -1, want1: false, args: args{offset: 20},
			fields: fields{offset: []*offsetStatus{{offset: 19, status: inProgress},
				{offset: 20, status: inProgress}}, lastCommitedOffset: -1, dirty: false},
		},
		{
			name: "5 Offset Commit intermediate having previous inprogress", want: -1, want1: false, args: args{offset: 21},
			fields: fields{offset: []*offsetStatus{{offset: 19, status: inProgress},
				{offset: 22, status: inProgress}, {offset: 20, status: completed},
				{offset: 21, status: inProgress}}, lastCommitedOffset: -1, dirty: false},
		},
		{
			name: "6 Offset Commit having next more inprogress", want: 20, want1: true, args: args{offset: 19},
			fields: fields{offset: []*offsetStatus{{offset: 19, status: inProgress},
				{offset: 22, status: inProgress}, {offset: 20, status: completed},
				{offset: 21, status: inProgress}}, lastCommitedOffset: -1, dirty: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &partitionState{
				offset:             tt.fields.offset,
				lastCommitedOffset: tt.fields.lastCommitedOffset,
				dirty:              tt.fields.dirty,
			}
			got, got1 := p.getCommitOffset(tt.args.offset)
			if got != tt.want {
				t.Errorf("partitionState.getCommitOffset() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("partitionState.getCommitOffset() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_partitionState_updateStatus(t *testing.T) {
	type fields struct {
		offset             []*offsetStatus
		lastCommitedOffset int64
		dirty              bool
	}
	type args struct {
		offset int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*offsetStatus
	}{
		{name: "1 Offset 0 Blank list", args: args{offset: 0}, fields: fields{}},
		{
			name: "2 Offset 0 / 21", args: args{offset: 0}, want: []*offsetStatus{{offset: 21, status: inProgress}},
			fields: fields{offset: []*offsetStatus{{offset: 21, status: inProgress}}, lastCommitedOffset: -1, dirty: false},
		},
		{
			name: "3 Offset 21 / 21", args: args{offset: 21}, want: []*offsetStatus{{offset: 21, status: completed}},
			fields: fields{offset: []*offsetStatus{{offset: 21, status: inProgress}}, lastCommitedOffset: -1, dirty: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &partitionState{
				offset:             tt.fields.offset,
				lastCommitedOffset: tt.fields.lastCommitedOffset,
				dirty:              tt.fields.dirty,
			}
			p.updateStatus(tt.args.offset)

			got := p.offset

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("partitionState.updateStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_partitionState_setOffset(t *testing.T) {
	type fields struct {
		offset             []*offsetStatus
		lastCommitedOffset int64
		dirty              bool
	}
	type args struct {
		offset []*offsetStatus
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*offsetStatus
	}{
		{name: "1 Nil", args: args{}, fields: fields{}},
		{name: "2 Blank", args: args{}, fields: fields{}},
		{
			name: "3 Assignement-1", args: args{offset: []*offsetStatus{{offset: 21, status: inProgress}}},
			fields: fields{offset: []*offsetStatus{{offset: 20, status: inProgress}}},
			want:   []*offsetStatus{{offset: 21, status: inProgress}},
		},
		{
			name: "4 Assignement-2", args: args{offset: []*offsetStatus{{offset: 22, status: completed}}},
			fields: fields{offset: []*offsetStatus{{offset: 21, status: completed}}},
			want:   []*offsetStatus{{offset: 22, status: completed}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &partitionState{
				offset:             tt.fields.offset,
				lastCommitedOffset: tt.fields.lastCommitedOffset,
				dirty:              tt.fields.dirty,
			}
			p.setOffset(tt.args.offset)
			got := p.offset

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("partitionState.setOffset() got = %v, want %v", got, tt.want)
			}
		})
	}
}
