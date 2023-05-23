package filter

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/converters/sql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/filter/strategies/tokenize"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/command"
)

const (
	query1 string = "parnerId=%v"
	val1   string = "f28f9d07-9853-43d2-a701-cb0764d8524b"
)

func TestMiddleware(t *testing.T) {
	type args struct {
		w        http.ResponseWriter
		r        *http.Request
		st       Strategy
		cnv      command.Converter
		callback func(string) string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantFilter *filter.Filter
	}{
		{
			name: "Success_Like",
			args: args{
				w: nil,
				r: func() *http.Request {
					r, _ := http.NewRequest(http.MethodGet,
						`https://ticket-service/v1/tickets?filter=(summary : "critical" OR summary : "emergency") AND (status = "New" OR status = "InProgress")`,
						nil)
					return r
				}(),
				st:  tokenize.GetStrategy(),
				cnv: sql.GetConverter(),
				callback: func(key string) string {
					fieldMap := map[string]string{
						"summary": "summary",
						"status":  "status",
					}
					return fieldMap[key]
				},
			},
			wantErr: false,
			wantFilter: filter.New(
				" ( ( summary  like  %v or summary  like  %v ) and ( (status) = (%v) or (status) = (%v) ) )",
				"%critical%", "%emergency%", "New", "InProgress",
			),
		},
		{
			name: "Success_In",
			args: args{
				w: nil,
				r: func() *http.Request {
					r, _ := http.NewRequest(http.MethodGet,
						`https://ticket-service/v1/tickets?filter=(summary IN "critical,emergency")`,
						nil)
					return r
				}(),
				st:  tokenize.GetStrategy(),
				cnv: sql.GetConverter(),
				callback: func(key string) string {
					fieldMap := map[string]string{
						"summary": "summary",
						"status":  "status",
					}
					return fieldMap[key]
				},
			},
			wantErr: false,
			wantFilter: filter.New(
				" ( ( summary in (%v,%v) ) )",
				"critical", "emergency",
			),
		},
		{
			name: "Invalid field UNKNOWN",
			args: args{
				w: nil,
				r: func() *http.Request {
					r, _ := http.NewRequest(http.MethodGet,
						`https://ticket-service/v1/tickets?filter=(UNKNOWN : "critical" OR summary : "emergency") AND (status = "New" OR status = "InProgress")`,
						nil)
					return r
				}(),
				st:  tokenize.GetStrategy(),
				cnv: sql.GetConverter(),
				callback: func(key string) string {
					fieldMap := map[string]string{
						"summary": "summary",
						"status":  "status",
					}
					return fieldMap[key]
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid operator UNKNOWN",
			args: args{
				w: nil,
				r: func() *http.Request {
					r, _ := http.NewRequest(http.MethodGet,
						`https://ticket-service/v1/tickets?filter=(summary UNKNOWN "critical" OR summary : "emergency") AND (status = "New" OR status = "InProgress")`,
						nil)
					return r
				}(),
				st:  tokenize.GetStrategy(),
				cnv: sql.GetConverter(),
				callback: func(key string) string {
					fieldMap := map[string]string{
						"summary": "summary",
						"status":  "status",
					}
					return fieldMap[key]
				},
			},
			wantErr: true,
		},
		{
			name: "Brackets mismatch",
			args: args{
				w: nil,
				r: func() *http.Request {
					r, _ := http.NewRequest(http.MethodGet,
						`https://ticket-service/v1/tickets?filter=(summary UNKNOWN "critical" OR summary : "emergency" AND (status = "New" OR status = "InProgress")`,
						nil)
					return r
				}(),
				st:  tokenize.GetStrategy(),
				cnv: sql.GetConverter(),
				callback: func(key string) string {
					fieldMap := map[string]string{
						"summary": "summary",
						"status":  "status",
					}
					return fieldMap[key]
				},
			},
			wantErr: true,
		},
		{
			name: "Space missing between field and operator",
			args: args{
				w: nil,
				r: func() *http.Request {
					r, _ := http.NewRequest(http.MethodGet,
						`https://ticket-service/v1/tickets?filter=(summary: "critical" OR summary : "emergency") AND (status = "New" OR status = "InProgress")`,
						nil)
					return r
				}(),
				st:  tokenize.GetStrategy(),
				cnv: sql.GetConverter(),
				callback: func(key string) string {
					fieldMap := map[string]string{
						"summary": "summary",
						"status":  "status",
					}
					return fieldMap[key]
				},
			},
			wantErr: true,
		},
		{
			name: "Space missing between operator and value",
			args: args{
				w: nil,
				r: func() *http.Request {
					r, _ := http.NewRequest(http.MethodGet,
						`https://ticket-service/v1/tickets?filter=(summary :"critical" OR summary : "emergency") AND (status = "New" OR status = "InProgress")`,
						nil)
					return r
				}(),
				st:  tokenize.GetStrategy(),
				cnv: sql.GetConverter(),
				callback: func(key string) string {
					fieldMap := map[string]string{
						"summary": "summary",
						"status":  "status",
					}
					return fieldMap[key]
				},
			},
			wantErr: true,
		},
		{
			name: "Error when operator in lower case",
			args: args{
				w: nil,
				r: func() *http.Request {
					r, _ := http.NewRequest(http.MethodGet,
						`https://ticket-service/v1/tickets?filter=(summary :"critical" or summary : "emergency") AND (status = "New" OR status = "InProgress")`,
						nil)
					return r
				}(),
				st:  tokenize.GetStrategy(),
				cnv: sql.GetConverter(),
				callback: func(key string) string {
					fieldMap := map[string]string{
						"summary": "summary",
						"status":  "status",
					}
					return fieldMap[key]
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var err error
			if tt.args.r, err = Middleware(tt.args.w, tt.args.r, tt.args.st, tt.args.cnv, tt.args.callback); (err != nil) != tt.wantErr {
				t.Errorf("Middleware() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == false {
				f, err := GetFilter(tt.args.r)
				if err != nil {
					t.Errorf("Middleware() GetFilter error = %v, wantErr %v", err, nil)
				}

				if !reflect.DeepEqual(f, tt.wantFilter) {
					t.Errorf("Middleware() f = %v, wantFilter %v", f, tt.wantFilter)
				}
			}
		})
	}
}

func TestFilter_SetRequestContext(t *testing.T) {

	testRequest, _ := http.NewRequest(http.MethodGet, "http://dummy.dummy", nil)

	type fields struct {
		query  string
		values []interface{}
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *http.Request
	}{
		{
			name: "Success",
			fields: fields{
				query:  query1,
				values: []interface{}{val1},
			},
			args: args{
				req: testRequest,
			},
			want: func() *http.Request {
				expectedFilter := filter.New(query1, val1)
				ctx := context.WithValue(testRequest.Context(), "filter", *expectedFilter)
				testRequest = testRequest.WithContext(ctx)
				return testRequest
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := filter.New(tt.fields.query, tt.fields.values...)
			got := SetRequestContext(f, tt.args.req)
			gotFilter := got.Context().Value(filterCtxKey)

			if !reflect.DeepEqual(gotFilter, f) {
				t.Errorf("Output.SetRequestContext() gotFilter = %v, wantFilter %v", gotFilter, f)
			}
		})
	}
}

func TestGetFilter(t *testing.T) {
	f := filter.New(query1, val1)
	f.ShouldAnd = true

	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *filter.Filter
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				req: func() *http.Request {
					r, _ := http.NewRequest(http.MethodGet, "http://dummy.dummy", nil)
					r = SetRequestContext(f, r)
					return r
				}(),
			},
			want: f,
		},
		{
			name: "Nil on missing header",
			args: args{
				req: func() *http.Request {
					r, _ := http.NewRequest(http.MethodGet, "http://dummy.dummy", nil)
					return r
				}(),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFilter(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
