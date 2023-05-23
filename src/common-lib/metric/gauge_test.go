package metric

import (
	"reflect"
	"testing"
)

func TestCreateGauge(t *testing.T) {
	type args struct {
		name        string
		description string
		value       int64
	}
	tests := []struct {
		name string
		args args
		want *Gauge
	}{
		{
			name: "1 Gauge-1",
			args: args{name: "Gauge", description: "description", value: 1},
			want: &Gauge{Name: "Gauge", Description: "description", Value: 1, Properties: map[string]string{}},
		},
		{
			name: "2 G-2",
			args: args{name: "G", description: "desc", value: 2},
			want: &Gauge{Name: "G", Description: "desc", Value: 2, Properties: map[string]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateGauge(tt.args.name, tt.args.description, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateGauge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGauge_Snapshot(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Value       int64
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{name: "Value 1", fields: fields{Value: 1}, want: 1},
		{name: "Value 2", fields: fields{Value: 2}, want: 2},
		{name: "Value 3", fields: fields{Value: 3}, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Value:       tt.fields.Value,
			}
			if got := g.Snapshot(); got != tt.want {
				t.Errorf("Gauge.Snapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGauge_Clear(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Value       int64
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{name: "Value 1", fields: fields{Value: 1}, want: 0},
		{name: "Value 2", fields: fields{Value: 2}, want: 0},
		{name: "Value 3", fields: fields{Value: 3}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Value:       tt.fields.Value,
			}
			g.Clear()
			if got := g.Snapshot(); got != tt.want {
				t.Errorf("Gauge.Clear() - Gauge.Snapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGauge_Inc(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Value       int64
	}
	type args struct {
		value int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{name: "Value 1", fields: fields{Value: 1}, args: args{value: 1}, want: 2},
		{name: "Value 2", fields: fields{Value: 2}, args: args{value: 2}, want: 4},
		{name: "Value 3", fields: fields{Value: 3}, args: args{value: 4}, want: 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Value:       tt.fields.Value,
			}
			g.Inc(tt.args.value)
			if got := g.Snapshot(); got != tt.want {
				t.Errorf("Gauge.Inc(value) - Gauge.Snapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGauge_Dec(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Value       int64
	}
	type args struct {
		value int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{name: "Value 1", fields: fields{Value: 3}, args: args{value: 1}, want: 2},
		{name: "Value 2", fields: fields{Value: 6}, args: args{value: 2}, want: 4},
		{name: "Value 3", fields: fields{Value: 11}, args: args{value: 4}, want: 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Value:       tt.fields.Value,
			}
			g.Dec(tt.args.value)
			if got := g.Snapshot(); got != tt.want {
				t.Errorf("Gauge.Dec(value) - Gauge.Snapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGauge_MetricType(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "1 Gauge 1", want: "Gauge"},
		{name: "2 Gauge 2", want: "Gauge"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{}
			if got := g.MetricType(); got != tt.want {
				t.Errorf("Gauge.MetricType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGauge_AddProperty(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Value       int64
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
			h := &Gauge{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Value:       tt.fields.Value,
				Properties:  tt.fields.Properties,
			}
			h.AddProperty(tt.args.key, tt.args.value)
			if got := h.Properties; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Gauge.AddProperty(key, value) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGauge_RemoveProperty(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		Value       int64
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
			h := &Gauge{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Value:       tt.fields.Value,
				Properties:  tt.fields.Properties,
			}
			h.RemoveProperty(tt.args.key)
			if got := h.Properties; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Gauge.RemoveProperty(key) = %v, want %v", got, tt.want)
			}
		})
	}
}
