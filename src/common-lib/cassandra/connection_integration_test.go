//go:build integration
// +build integration

package cassandra

import (
	"fmt"
	"testing"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra/cql"

	"github.com/gocql/gocql"
)

var cassandraHosts = []string{"localhost:9042"}

func TestNewDbConnection(t *testing.T) {
	type args struct {
		conf *DbConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Nil Config", wantErr: true, args: args{conf: nil}},
		{name: "Blank Config", wantErr: true, args: args{conf: &DbConfig{}}},
		{name: "Blank Host", wantErr: true, args: args{conf: &DbConfig{Hosts: []string{}}}},
		{name: "Blank Keyspace", wantErr: true, args: args{conf: &DbConfig{Hosts: []string{"Sever"}, Keyspace: ""}}},
		{name: "Invalid Keyspace", wantErr: true, args: args{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "???"}}},
		{name: "Valid Keyspace", wantErr: false, args: args{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "system"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDbConnection(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDbConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_connection_Insert(t *testing.T) {
	type fields struct {
		conf *DbConfig
	}

	type args struct {
		table        string
		id           string
		timestamp    time.Time
		source       string
		updateSource string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantInsertErr bool
		wantSelectErr bool
		wantUpdateErr bool
		wantDeleteErr bool
	}{
		{
			name: "Blank table", wantInsertErr: true, wantUpdateErr: true, wantSelectErr: true, wantDeleteErr: true,
			fields: fields{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "platform_common_test"}},
			args:   args{table: ""},
		},
		{
			name: "Invalid table", wantInsertErr: true, wantUpdateErr: true, wantSelectErr: true, wantDeleteErr: true,
			fields: fields{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "platform_common_test"}},
			args:   args{table: "platform_common_test.invalid", id: "1", timestamp: time.Now().UTC(), source: "First Test"},
		},
		{
			name: "Valid table", wantInsertErr: false, wantUpdateErr: false, wantSelectErr: false, wantDeleteErr: false,
			fields: fields{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "platform_common_test"}},
			args: args{table: "platform_common_test.test", id: "1", timestamp: time.Now().UTC(),
				updateSource: "Update First Source", source: "First Test"},
		},
	}

	validateSelect := func(table string, id string, wantErr bool, want string, c DbConnector, t *testing.T) {
		query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", table)
		data, err := c.Select(query, id)
		if (err != nil || len(data) == 0) != wantErr {
			t.Errorf("connection.Select() error = %v, wantErr %v", err, wantErr)
			return
		}

		if len(data) > 0 && data[0]["source"] != want {
			t.Errorf("connection.Select() Source = %v, want %v", data[0]["source"], want)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewDbConnection(tt.fields.conf)

			if err != nil {
				t.Errorf("NewDbConnection() error = %v", err)
				return
			}

			query := fmt.Sprintf("Insert into %s (id, timestamp_utc, source) values (?, ?, ?)", tt.args.table)
			if err = c.Insert(query, tt.args.id, tt.args.timestamp, tt.args.source); (err != nil) != tt.wantInsertErr {
				t.Errorf("connection.Insert() error = %v, wantErr %v", err, tt.wantInsertErr)
			}

			validateSelect(tt.args.table, tt.args.id, tt.wantSelectErr, tt.args.source, c, t)

			query = fmt.Sprintf("Update %s Set timestamp_utc = ?, source = ? Where id = ?", tt.args.table)
			if err = c.Update(query, tt.args.timestamp, tt.args.updateSource, tt.args.id); (err != nil) != tt.wantUpdateErr {
				t.Errorf("connection.Update() error = %v, wantErr %v", err, tt.wantUpdateErr)
			}

			validateSelect(tt.args.table, tt.args.id, tt.wantSelectErr, tt.args.updateSource, c, t)

			query = fmt.Sprintf("Delete from %s Where id = ?", tt.args.table)
			if err = c.Delete(query, tt.args.id); (err != nil) != tt.wantDeleteErr {
				t.Errorf("connection.Delete() error = %v, wantErr %v", err, tt.wantDeleteErr)
			}

			validateSelect(tt.args.table, tt.args.id, true, "", c, t)

			c.Close()

			if !c.Closed() {
				t.Errorf("connection.Closed() expected true got false")
			}
		})
	}
}

func Test_connection_executeQuery(t *testing.T) {
	conf := &DbConfig{Hosts: cassandraHosts, Keyspace: "platform_common_test"}
	c, _ := newConnection(conf)
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
		wantErr bool
	}{
		{name: "nil Session", wantErr: true},
		{name: "Blank Session", wantErr: true, args: args{query: "", policy: &gocql.SimpleRetryPolicy{}},
			fields: fields{session: c.session}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := connection{
				session: tt.fields.session,
				conf:    tt.fields.conf,
			}
			if err := d.executeQuery(tt.args.policy, tt.args.query, tt.args.value...); (err != nil) != tt.wantErr {
				t.Errorf("connection.executeQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_connection_InsertWithScanCas(t *testing.T) {
	type fields struct {
		conf *DbConfig
	}
	type args struct {
		dest         *map[string]interface{}
		table        string
		id           string
		timestamp    time.Time
		source       string
		updateSource string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		want          bool
		wantInsertErr bool
		wantSelectErr bool
		wantUpdateErr bool
		wantDeleteErr bool
	}{
		{
			name: "Blank table", wantInsertErr: true, wantUpdateErr: true, wantSelectErr: true, wantDeleteErr: true,
			fields: fields{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "platform_common_test"}},
			args:   args{table: ""},
		},
		{
			name: "Invalid table", wantInsertErr: true, wantUpdateErr: true, wantSelectErr: true, wantDeleteErr: true,
			fields: fields{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "platform_common_test"}},
			args:   args{table: "platform_common_test.invalid", id: "3", timestamp: time.Now().UTC(), source: "First Test"},
		},
		{
			name: "Valid table", wantInsertErr: true, wantUpdateErr: true, wantSelectErr: true, wantDeleteErr: true,
			fields: fields{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "platform_common_test"}},
			args: args{table: "platform_common_test.test", id: "3", timestamp: time.Now().UTC(),
				updateSource: "Update First Source", source: "First Test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewDbConnection(tt.fields.conf)

			if err != nil {
				t.Errorf("NewDbConnection() error = %v", err)
				return
			}

			query := fmt.Sprintf("Insert into %s (id, timestamp_utc, source) values (?, ?, ?)", tt.args.table)
			got, err := c.InsertWithScanCas(tt.args.dest, query, tt.args.id, tt.args.timestamp, tt.args.source)
			if (err != nil) != tt.wantInsertErr {
				t.Errorf("connection.InsertWithScanCas() error = %v, wantErr %v", err, tt.wantInsertErr)
			}

			if got != tt.want {
				t.Errorf("connection.InsertWithScanCas() = %v, want %v", got, tt.want)
			}

			query = fmt.Sprintf("Update %s Set timestamp_utc = ?, source = ? Where id = ?", tt.args.table)
			got, err = c.UpdateWithScanCas(tt.args.dest, query, tt.args.timestamp, tt.args.updateSource, tt.args.id)
			if (err != nil) != tt.wantUpdateErr {
				t.Errorf("connection.UpdateWithScanCas() error = %v, wantErr %v", err, tt.wantUpdateErr)
			}

			if got != tt.want {
				t.Errorf("connection.UpdateWithScanCas() = %v, want %v", got, tt.want)
			}

			query = fmt.Sprintf("Delete from %s Where id = ?", tt.args.table)
			got, err = c.DeleteWithScanCas(tt.args.dest, query, tt.args.id)
			if (err != nil) != tt.wantDeleteErr {
				t.Errorf("connection.DeleteWithScanCas() error = %v, wantErr %v", err, tt.wantDeleteErr)
			}

			if got != tt.want {
				t.Errorf("connection.DeleteWithScanCas() = %v, want %v", got, tt.want)
			}

			c.Close()

			if !c.Closed() {
				t.Errorf("connection.Closed() expected true got false")
			}

		})
	}
}

func Test_connection_ExecuteDmlWithRetrial(t *testing.T) {
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
		wantErr bool
	}{
		{name: "nil Policy", wantErr: true},
		{name: "nil Session", wantErr: true, args: args{policy: &gocql.SimpleRetryPolicy{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
