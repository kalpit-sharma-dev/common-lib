package timeutil

import (
	"reflect"
	"testing"
	"time"
)

func TestToHourLong(t *testing.T) {
	fromTm1, _ := time.Parse(time.RFC3339, "2018-02-05T11:24:53.489110939Z")
	toTm1, _ := time.Parse(time.RFC3339, "2018-02-05T12:24:53.489110939Z")

	fromTm2, _ := time.Parse(time.RFC3339, "2018-02-05T11:24:53.489110939Z")
	toTm2, _ := time.Parse(time.RFC3339, "2018-02-05T20:44:53.489110939Z")

	fromTm3, _ := time.Parse(time.RFC3339, "2018-02-05T23:24:53.489110939Z")
	toTm3, _ := time.Parse(time.RFC3339, "2018-02-06T01:24:53.489110939Z")

	fromTm4, _ := time.Parse(time.RFC3339, "2018-02-04T23:24:53.489110939Z")
	toTm4, _ := time.Parse(time.RFC3339, "2018-02-06T05:24:53.489110939Z")

	type args struct {
		fromTime time.Time
		toTime   time.Time
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "Test1",
			args: args{fromTime: fromTm1, toTime: toTm1},
			want: []int{20180205},
		},
		{
			name: "Test2",
			args: args{fromTime: fromTm2, toTime: toTm2},
			want: []int{20180205},
		},
		{
			name: "Test3",
			args: args{fromTime: fromTm3, toTime: toTm3},
			want: []int{20180205, 20180206},
		},
		{
			name: "Test4",
			args: args{fromTime: fromTm4, toTime: toTm4},
			want: []int{20180204, 20180205, 20180206},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToHourLong(tt.args.fromTime, tt.args.toTime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToHourLong() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCurrentTime(t *testing.T) {
	type args struct {
		locationName string
	}

	//InputAgruments
	//Location Asia/Cal does not exists
	errArg := args{locationName: "Asia/Cal"}
	//Location Asia/Calcutta does  exists
	successArg := args{locationName: "Asia/Calcutta"}

	//want
	now := time.Now()
	location, _ := time.LoadLocation("Asia/Calcutta")
	specifiedZoneTime := now.In(location)
	currTime := &CurrentTime{
		Time: specifiedZoneTime.Format(timeFormat),
		Date: specifiedZoneTime.Format(dateFormat),
		Day:  specifiedZoneTime.Format(dayFormat),
	}

	tests := []struct {
		name    string
		args    args
		want    *CurrentTime
		wantErr bool
	}{

		{name: "Error (wrong timezone Name)", args: errArg, want: nil, wantErr: true},
		{name: "Success (timezone Asia/Calcutta)", args: successArg, want: currTime, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCurrentTime(tt.args.locationName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCurrentTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToLongYYYYMMDDHH(t *testing.T) {
	type args struct {
		tm *time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Test1",
			args:    args{&time.Time{}},
			want:    1010100,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToLongYYYYMMDDHH(tt.args.tm)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToLongYYYYMMDDHH() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToLongYYYYMMDDHH() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestGetTimeZoneOffsetFromLocation(t *testing.T) {
	type args struct {
		locationName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test1 : Valid Location Positive",
			args: args{
				locationName: "Asia/Calcutta",
			},
			want:    "+0530",
			wantErr: false,
		},
		{
			name: "Test2 : Valid Location Negative",
			args: args{
				locationName: "Pacific/Samoa",
			},
			want:    "-1100",
			wantErr: false,
		},
		{
			name: "Test1 : InValid Location",
			args: args{
				locationName: "AA",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTimeZoneOffsetFromLocation(tt.args.locationName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTimeZoneOffsetFromLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTimeZoneOffsetFromLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}
