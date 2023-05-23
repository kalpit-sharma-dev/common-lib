package metric

import (
	"reflect"
	"testing"
)

func TestCreateHistogram(t *testing.T) {
	type args struct {
		name        string
		description string
		values      []float64
	}
	tests := []struct {
		name string
		args args
		want *Histogram
	}{
		{
			name: "1 Histogram-1",
			args: args{name: "Histogram", description: "description", values: []float64{1}},
			want: &Histogram{Name: "Histogram", Description: "description", Values: []float64{1}, Properties: map[string]string{}},
		},
		{
			name: "2 H-2",
			args: args{name: "H", description: "desc", values: []float64{2, 3}},
			want: &Histogram{Name: "H", Description: "desc", Values: []float64{2, 3}, Properties: map[string]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateHistogram(tt.args.name, tt.args.description, tt.args.values); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateHistogram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHistogram_Snapshot(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Values      []float64
	}
	tests := []struct {
		name   string
		fields fields
		want   []float64
	}{
		{name: "Value 1", fields: fields{Values: []float64{2, 3}}, want: []float64{2, 3}},
		{name: "Value 2", fields: fields{Values: []float64{4, 5}}, want: []float64{4, 5}},
		{name: "Value 3", fields: fields{Values: []float64{6, 6}}, want: []float64{6, 6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Histogram{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Values:      tt.fields.Values,
			}
			if got := h.Snapshot(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Histogram.Snapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHistogram_Update(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Values      []float64
	}
	type args struct {
		values []float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []float64
	}{
		{name: "Value 1", fields: fields{Values: []float64{2, 3}}, args: args{values: []float64{3, 4}}, want: []float64{3, 4}},
		{name: "Value 2", fields: fields{Values: []float64{4, 5}}, args: args{values: []float64{3, 4}}, want: []float64{3, 4}},
		{name: "Value 3", fields: fields{Values: []float64{6, 6}}, args: args{values: []float64{5, 4}}, want: []float64{5, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Histogram{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Values:      tt.fields.Values,
			}
			h.Update(tt.args.values)
			if got := h.Snapshot(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Histogram.Snapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHistogram_Clear(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Values      []float64
	}
	tests := []struct {
		name   string
		fields fields
		want   []float64
	}{
		{name: "Value 1", fields: fields{Values: []float64{2, 3}}, want: []float64{}},
		{name: "Value 2", fields: fields{Values: []float64{4, 5}}, want: []float64{}},
		{name: "Value 3", fields: fields{Values: []float64{6, 6}}, want: []float64{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Histogram{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Values:      tt.fields.Values,
			}
			h.Clear()
			if got := h.Snapshot(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Histogram.Snapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHistogram_MetricType(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "1 Histogram 1", want: "Histogram"},
		{name: "2 Histogram 2", want: "Histogram"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Histogram{}
			if got := h.MetricType(); got != tt.want {
				t.Errorf("Histogram.MetricType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHistogram_AddProperty(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Values      []float64
		Properties  map[string]string
	}
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{name: "1 Blank", want: map[string]string{"": ""}, fields: fields{Properties: map[string]string{}}},
		{name: "2 Key", want: map[string]string{"Key": ""}, args: args{key: "Key"}, fields: fields{Properties: map[string]string{}}},
		{name: "3 Key Value", want: map[string]string{"Key": "Value"}, args: args{key: "Key", value: "Value"}, fields: fields{Properties: map[string]string{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Histogram{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Values:      tt.fields.Values,
				Properties:  tt.fields.Properties,
			}
			h.AddProperty(tt.args.key, tt.args.value)
			if got := h.Properties; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Histogram.AddProperty(key, value) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHistogram_RemoveProperty(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Values      []float64
		Properties  map[string]string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{name: "1 Blank", want: map[string]string{}, fields: fields{Properties: map[string]string{}}},
		{name: "2 Key", want: map[string]string{}, args: args{key: "Key"}, fields: fields{Properties: map[string]string{"Key": ""}}},
		{name: "3 Key Value", want: map[string]string{"Key": "Value"}, args: args{key: ""}, fields: fields{Properties: map[string]string{"Key": "Value"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Histogram{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Values:      tt.fields.Values,
				Properties:  tt.fields.Properties,
			}
			h.RemoveProperty(tt.args.key)
			if got := h.Properties; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Histogram.RemoveProperty(key) = %v, want %v", got, tt.want)
			}
		})
	}
}
