package metric

import (
	"reflect"
	"testing"
)

func TestNDIMCounter_Snapshot(t *testing.T) {
	tests := []struct {
		name string
		n    *NDIMCounter
		want map[string]int64
	}{
		// TODO: Add test cases.
		{
			name: "default",
			n: &NDIMCounter{
				Name:        "metricName",
				Description: "metric Description",
				DimCounters: map[string]int64{
					"counter1": 1,
					"counter2": 2,
				},
			},
			want: map[string]int64{
				"counter1": 1,
				"counter2": 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.Snapshot()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NDIMCounter.Snapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNDIMCounter_Clear(t *testing.T) {
	tests := []struct {
		name string
		n    *NDIMCounter
	}{
		// TODO: Add test cases.
		{
			name: "default",
			n: &NDIMCounter{
				Name:        "",
				Description: "",
				Properties:  map[string]string{},
				DimCounters: map[string]int64{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.Clear()
		})
	}
}

func TestNDIMCounter_Inc(t *testing.T) {
	dim1 := "dim1"
	type args struct {
		dimension string
		value     int64
	}
	tests := []struct {
		name string
		n    *NDIMCounter
		args args
		want int64
	}{
		// TODO: Add test cases.
		{
			name: "metric1",
			n: &NDIMCounter{
				Name:        "",
				Description: "",
				Properties: map[string]string{
					"": "",
				},
				DimCounters: map[string]int64{
					dim1: 0,
				},
			},
			args: args{
				dimension: "dim1",
				value:     1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.Inc(tt.args.dimension, tt.args.value)
			got := tt.n.GetDimCounters()[dim1]
			if got != tt.want {
				t.Errorf("want: %d, got: %d", tt.want, got)
			}

		})
	}
}

func TestNDIMCounter_MetricType(t *testing.T) {
	tests := []struct {
		name string
		n    *NDIMCounter
		want string
	}{
		// TODO: Add test cases.
		{
			name: "default",
			n:    &NDIMCounter{},
			want: NDIMCounterType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.MetricType(); got != tt.want {
				t.Errorf("NDIMCounter.MetricType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNDIMCounter_AddProperty(t *testing.T) {
	propertyName := "prop1"
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		n    *NDIMCounter
		args args
	}{
		// TODO: Add test cases.
		{
			name: "",
			n: &NDIMCounter{
				Name:        propertyName,
				Description: "",
				Properties:  map[string]string{},
				DimCounters: map[string]int64{},
			},
			args: args{
				key:   propertyName,
				value: "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.AddProperty(tt.args.key, tt.args.value)
			if tt.n.Properties[propertyName] != tt.args.value {
				t.Errorf("Property added not found")
			}
		})
	}
}

func TestNDIMCounter_RemoveProperty(t *testing.T) {
	propertyName := "prop1"
	type args struct {
		key string
	}
	tests := []struct {
		name string
		n    *NDIMCounter
		args args
	}{
		// TODO: Add test cases.
		{
			name: "",
			n: &NDIMCounter{
				Name:        "",
				Description: "",
				Properties: map[string]string{
					propertyName: "0",
				},
			},
			args: args{
				key: propertyName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.RemoveProperty(tt.args.key)
			if len(tt.n.Properties) != 0 {
				t.Errorf("Property not removed")
			}
		})
	}
}

func TestNDIMCounter_AddDimension(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Properties  map[string]string
		DimCounters map[string]int64
	}
	type args struct {
		dimkey string
		value  int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "default",
			fields: fields{
				Name:        "metric1",
				Description: "metric desc",
				DimCounters: map[string]int64{},
			},
			args: args{
				dimkey: "dim1",
				value:  1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := &NDIMCounter{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				DimCounters: tt.fields.DimCounters,
			}
			n.AddDimension(tt.args.dimkey, tt.args.value)

			if tt.fields.DimCounters[tt.args.dimkey] != tt.args.value {
				t.Errorf("Dimension not added")
			}
		})
	}
}

func TestNDIMCounter_RemoveDimension(t *testing.T) {
	dimkey := "dim1"
	type fields struct {
		Name        string
		Description string
		Properties  map[string]string
		DimCounters map[string]int64
	}
	type args struct {
		dimkey string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "default",
			fields: fields{
				Name:        "metric1",
				Description: "desc",
				Properties:  map[string]string{},
				DimCounters: map[string]int64{
					dimkey: 1,
				},
			},
			args: args{
				dimkey: dimkey,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NDIMCounter{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Properties:  tt.fields.Properties,
				DimCounters: tt.fields.DimCounters,
			}
			n.RemoveDimension(tt.args.dimkey)
			if tt.fields.DimCounters[tt.args.dimkey] != 0 {
				t.Errorf("Dimension not removed")
			}
		})
	}
}
