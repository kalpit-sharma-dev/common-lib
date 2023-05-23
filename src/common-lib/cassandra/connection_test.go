//go:build !integration
// +build !integration

package cassandra

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra/cql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra/cql/mock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/validate/is"
)

func TestNewDbConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	session := mock.NewMockSession(ctrl)

	type args struct {
		conf *DbConfig
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    DbConnector
		wantErr bool
	}{
		{
			name: "Nil Config", wantErr: true, want: nil, args: args{conf: nil}, setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return session, errors.New("Error") }
			},
		},
		{
			name: "Blank Config", wantErr: true, want: nil, args: args{conf: &DbConfig{}}, setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return session, errors.New("Error") }
			},
		},
		{
			name: "Blank Host", wantErr: true, want: nil, args: args{conf: &DbConfig{Hosts: []string{}}}, setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return session, errors.New("Error") }
			},
		},
		{
			name: "Blank Keyspace", wantErr: true, want: nil, args: args{conf: &DbConfig{Hosts: []string{"Sever"}, Keyspace: ""}}, setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return session, errors.New("Error") }
			},
		},
		{
			name: "Session Creation Error", wantErr: true, want: nil,
			args: args{conf: &DbConfig{Hosts: []string{"Sever"}, Keyspace: "test"}}, setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return session, errors.New("Error") }
			},
		},
		{
			name: "Session Creation Success", wantErr: false,
			want: &connection{conf: &DbConfig{Hosts: []string{"Server"}, Keyspace: "test", Consistency: gocql.Quorum,
				NumConn: 10, TimeoutMillisecond: time.Millisecond * 10, DisableInitialHostLookup: false, ConnectTimeout: time.Millisecond * 600, CircuitBreaker: &circuit.Config{
					Enabled: true, TimeoutInSecond: 1, MaxConcurrentRequests: 15000,
					ErrorPercentThreshold: 25, RequestVolumeThreshold: 500, SleepWindowInSecond: 10,
				}, CommandName: "Database-Command1", ValidErrors: []string{"Error1"}}, session: session},
			args: args{conf: &DbConfig{Hosts: []string{"Server"}, Keyspace: "test",
				NumConn: 10, TimeoutMillisecond: time.Millisecond * 10, CircuitBreaker: &circuit.Config{
					Enabled: true, TimeoutInSecond: 1, MaxConcurrentRequests: 15000,
					ErrorPercentThreshold: 25, RequestVolumeThreshold: 500, SleepWindowInSecond: 10,
				}, CommandName: "Database-Command1", ValidErrors: []string{"Error1"}}},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return session, nil }
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := NewDbConnection(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDbConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDbConnection() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_connection_Insert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	session := mock.NewMockSession(ctrl)
	query := mock.NewMockQuery(ctrl)
	type fields struct {
		session cql.Session
		conf    *DbConfig
	}
	type args struct {
		query string
		value []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "Nil Session", wantErr: true,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: nil},
			setup:  func() {},
		},
		{
			name: "Query Execute Error", wantErr: true,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().Release()
				query.EXPECT().Exec().Return(errors.New("Error"))
			},
		},
		{
			name: "Query Execute Success", wantErr: false,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().Release()
				query.EXPECT().Exec().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			if err := d.Insert(tt.args.query, tt.args.value...); (err != nil) != tt.wantErr {
				t.Errorf("connection.Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_connection_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	session := mock.NewMockSession(ctrl)
	query := mock.NewMockQuery(ctrl)

	type fields struct {
		session cql.Session
		conf    *DbConfig
	}
	type args struct {
		query string
		value []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "Nil Session", wantErr: true,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: nil},
			setup:  func() {},
		},
		{
			name: "Query Execute Error", wantErr: true,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().Release()
				query.EXPECT().Exec().Return(errors.New("Error"))
			},
		},
		{
			name: "Query Execute Success", wantErr: false,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().Release()
				query.EXPECT().Exec().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			if err := d.Update(tt.args.query, tt.args.value...); (err != nil) != tt.wantErr {
				t.Errorf("connection.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_connection_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	session := mock.NewMockSession(ctrl)
	query := mock.NewMockQuery(ctrl)

	type fields struct {
		session cql.Session
		conf    *DbConfig
	}
	type args struct {
		query string
		value []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "Nil Session", wantErr: true,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: nil},
			setup:  func() {},
		},
		{
			name: "Query Execute Error", wantErr: true,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().Release()
				query.EXPECT().Exec().Return(errors.New("Error"))
			},
		},
		{
			name: "Query Execute Success", wantErr: false,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().Release()
				query.EXPECT().Exec().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			if err := d.Delete(tt.args.query, tt.args.value...); (err != nil) != tt.wantErr {
				t.Errorf("connection.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_connection_ExecuteDmlWithRetrial(t *testing.T) {
	ctrl := gomock.NewController(t)
	session := mock.NewMockSession(ctrl)
	query := mock.NewMockQuery(ctrl)

	type fields struct {
		session cql.Session
		conf    *DbConfig
	}
	type args struct {
		policy gocql.RetryPolicy
		query  string
		value  []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "Nil Policy", wantErr: true,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: nil},
			setup:  func() {},
		},
		{
			name: "Nil Session", wantErr: true,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: nil},
			args:   args{policy: &gocql.SimpleRetryPolicy{}},
			setup:  func() {},
		},
		{
			name: "Query Execute Error", wantErr: true,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session}, args: args{policy: &gocql.SimpleRetryPolicy{}},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().RetryPolicy(gomock.Any())
				query.EXPECT().Release()
				query.EXPECT().Exec().Return(errors.New("Error"))
			},
		},
		{
			name: "Query Execute Success", wantErr: false,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session}, args: args{policy: &gocql.SimpleRetryPolicy{}},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().RetryPolicy(gomock.Any())
				query.EXPECT().Release()
				query.EXPECT().Exec().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			if err := d.ExecuteDmlWithRetrial(tt.args.policy, tt.args.query, tt.args.value...); (err != nil) != tt.wantErr {
				t.Errorf("connection.ExecuteDmlWithRetrial() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_connection_Select(t *testing.T) {
	ctrl := gomock.NewController(t)
	session := mock.NewMockSession(ctrl)
	query := mock.NewMockQuery(ctrl)

	type fields struct {
		session cql.Session
		conf    *DbConfig
	}
	type args struct {
		query string
		value []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		want    []map[string]interface{}
		wantErr bool
	}{
		{
			name: "Slice Map Success", want: []map[string]interface{}{},
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session}, args: args{query: "q"},
			setup: func() {
				session.EXPECT().Query(gomock.Any()).Return(query)
				query.EXPECT().Iter().Return(&gocql.Iter{})
				query.EXPECT().Release()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			got, err := d.Select(tt.args.query, tt.args.value...)
			if (err != nil) != tt.wantErr {
				t.Errorf("connection.Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("connection.Select() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_connection_SelectWithRetrial(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	session := mock.NewMockSession(ctrl)
	query := mock.NewMockQuery(ctrl)
	type fields struct {
		session cql.Session
		conf    *DbConfig
	}
	type args struct {
		policy gocql.RetryPolicy
		query  string
		value  []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		want    []map[string]interface{}
		wantErr bool
	}{
		{
			name: "Nil Policy", wantErr: true, want: nil,
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session}, args: args{query: "q"},
			setup: func() {
			},
		},
		{
			name: "Simple Policy", wantErr: false, want: []map[string]interface{}{},
			fields: fields{conf: &DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}, session: session}, args: args{query: "q", policy: &gocql.SimpleRetryPolicy{}},
			setup: func() {
				session.EXPECT().Query(gomock.Any()).Return(query)
				query.EXPECT().RetryPolicy(gomock.Any())
				query.EXPECT().Iter().Return(&gocql.Iter{})
				query.EXPECT().Release()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			got, err := d.SelectWithRetrial(tt.args.policy, tt.args.query, tt.args.value...)
			if (err != nil) != tt.wantErr {
				t.Errorf("connection.SelectWithRetrial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("connection.SelectWithRetrial() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_connection_GetRandomUUID(t *testing.T) {
	t.Run("Generate UUID", func(t *testing.T) {
		c := connection{}
		got, err := c.GetRandomUUID()
		if err != nil {
			t.Errorf("connection.GetRandomUUID() error = %v, want nil", err)
			return
		}

		if !is.UUID(got) {
			t.Errorf("connection.GetRandomUUID() = %v, want UUID", got)
		}
	})
}

func Test_connection_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	session := mock.NewMockSession(ctrl)
	type fields struct {
		session cql.Session
		conf    *DbConfig
	}
	tests := []struct {
		name   string
		fields fields
		setup  func()
	}{
		{
			name:   "Nil Session",
			fields: fields{conf: &DbConfig{}, session: nil},
			setup:  func() {},
		},
		{
			name:   "Session",
			fields: fields{conf: &DbConfig{}, session: session},
			setup:  func() { session.EXPECT().Close() },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			d.Close()
		})
	}
}

func Test_connection_Closed(t *testing.T) {
	ctrl := gomock.NewController(t)
	session := mock.NewMockSession(ctrl)
	type fields struct {
		session cql.Session
		conf    *DbConfig
	}
	tests := []struct {
		name   string
		fields fields
		setup  func()
		want   bool
	}{
		{
			name: "Nil Session", want: true,
			fields: fields{conf: &DbConfig{}, session: nil},
			setup:  func() {},
		},
		{
			name: "Session Closed", want: true,
			fields: fields{conf: &DbConfig{}, session: session},
			setup:  func() { session.EXPECT().Closed().Return(true) },
		},
		{
			name: "Session Open", want: false,
			fields: fields{conf: &DbConfig{}, session: session},
			setup:  func() { session.EXPECT().Closed().Return(false) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			if got := d.Closed(); got != tt.want {
				t.Errorf("connection.Closed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_connection_InsertWithScanCas(t *testing.T) {
	ctrl := gomock.NewController(t)
	session := mock.NewMockSession(ctrl)
	query := mock.NewMockQuery(ctrl)
	config := DbConfig{CommandName: "DB", CircuitBreaker: circuit.New()}
	type fields struct {
		session cql.Session
		conf    *DbConfig
	}
	type args struct {
		dest  *map[string]interface{}
		query string
		value []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		want    bool
		wantErr bool
	}{
		{
			name: "Nil Session", wantErr: true,
			fields: fields{conf: &DbConfig{}, session: nil},
			setup:  func() {},
		},
		{
			name: "Query Execute Error", wantErr: true, want: false,
			fields: fields{conf: &config, session: session},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().MapScanCAS(gomock.Any()).Return(true, errors.New("Error"))
				query.EXPECT().Release()
			},
		},
		{
			name: "Query Execute Success Applied", wantErr: false, want: true,
			fields: fields{conf: &config, session: session},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().MapScanCAS(gomock.Any()).Return(true, nil)
				query.EXPECT().Release()
			},
		},
		{
			name: "Query Execute Success Not Applied", wantErr: false, want: false,
			fields: fields{conf: &config, session: session}, args: args{dest: &map[string]interface{}{}},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().MapScanCAS(gomock.Any()).Return(false, nil)
				query.EXPECT().Release()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			got, err := d.InsertWithScanCas(tt.args.dest, tt.args.query, tt.args.value...)
			if (err != nil) != tt.wantErr {
				t.Errorf("connection.InsertWithScanCas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("connection.InsertWithScanCas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_connection_SelectWithPaging(t *testing.T) {
	ctrl := gomock.NewController(t)
	session := mock.NewMockSession(ctrl)
	query := mock.NewMockQuery(ctrl)

	type fields struct {
		session cql.Session
		conf    *DbConfig
	}
	type args struct {
		page     int
		callback ProcessRow
		query    string
		value    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name:   "test",
			fields: fields{conf: &DbConfig{}, session: session},
			setup: func() {
				session.EXPECT().Query(gomock.Any(), gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any())
				query.EXPECT().Iter().Return(&gocql.Iter{})
				query.EXPECT().Release()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			if err := d.SelectWithPaging(tt.args.page, tt.args.callback, tt.args.query, tt.args.value...); (err != nil) != tt.wantErr {
				t.Errorf("connection.SelectWithPaging() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
