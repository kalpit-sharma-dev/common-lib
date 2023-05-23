package webClient

import (
	"reflect"
	"testing"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

func TestHealth(t *testing.T) {
	type args struct {
		cbList []CircuitBreakerConfig
	}
	tests := []struct {
		name    string
		args    args
		want    []rest.Statuser
		wantErr bool
	}{
		{
			name: "Valid baseURL",
			args: args{
				[]CircuitBreakerConfig{{
					BaseURL: "http://localhost:9090/hello",
				},
					{
						BaseURL: "http://localhost:8080/helloworld",
					},
				},
			},
			want:    []rest.Statuser{status{"localhost:9090", "http://localhost:9090/hello"}, status{"localhost:8080", "http://localhost:8080/helloworld"}},
			wantErr: false,
		},
		{
			name: "baseURL parsing failure",
			args: args{
				[]CircuitBreakerConfig{{
					BaseURL: "%gh&%ij",
				}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Health(tt.args.cbList)
			if (err != nil) != tt.wantErr {
				t.Errorf("Health() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Health() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_status_Status(t *testing.T) {
	tm := time.Now()

	RegisterCircuitBreaker([]CircuitBreakerConfig{{
		BaseURL: "http://www.google.com",
	}})

	connectionStatus := rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "1", Name: "1"}

	type fields struct {
		commandName string
		baseURL     string
	}
	type args struct {
		conn rest.OutboundConnectionStatus
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *rest.OutboundConnectionStatus
	}{
		{
			name: "UninitializedState", args: args{conn: connectionStatus},
			fields: fields{commandName: "www.amazon.com", baseURL: "http://www.amazon.com"},
			want: &rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "1", Name: "1", ConnectionType: "http://www.amazon.com",
				ConnectionURLs: []string{"http://www.amazon.com"}, ConnectionStatus: rest.ConnectionStatusUnavailable},
		},
		{
			name: "ClosedState", args: args{conn: connectionStatus},
			fields: fields{commandName: "www.google.com", baseURL: "http://www.google.com"},
			want: &rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "1", Name: "1", ConnectionType: "http://www.google.com",
				ConnectionURLs: []string{"http://www.google.com"}, ConnectionStatus: rest.ConnectionStatusActive},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := status{
				commandName: tt.fields.commandName,
				baseURL:     tt.fields.baseURL,
			}
			if got := k.Status(tt.args.conn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("status.Status() = %v, want %v", got, tt.want)
			}
		})
	}
}
