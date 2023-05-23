// +build integration

package cassandra

import (
	"fmt"
	"testing"
	"time"
)

func TestNewBatchDbConnection(t *testing.T) {
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
			_, err := NewBatchDbConnection(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDbConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_connection_BatchExecution(t *testing.T) {
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
			args:   args{table: "platform_common_test.invalid", id: "2", timestamp: time.Now().UTC(), source: "First Test"},
		},
		{
			name: "Valid table", wantInsertErr: false, wantUpdateErr: false, wantSelectErr: false, wantDeleteErr: false,
			fields: fields{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "platform_common_test"}},
			args: args{table: "platform_common_test.test", id: "2", timestamp: time.Now().UTC(),
				updateSource: "Update Batch First Source", source: "Batch First Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewBatchDbConnection(tt.fields.conf)

			if err != nil {
				t.Errorf("NewDbConnection() error = %v", err)
				return
			}

			query := fmt.Sprintf("Insert into %s (id, timestamp_utc, source) values (?, ?, ?)", tt.args.table)
			if err = c.BatchExecution(query, [][]interface{}{[]interface{}{tt.args.id, tt.args.timestamp, tt.args.source}}); (err != nil) != tt.wantInsertErr {
				t.Errorf("connection.Insert() error = %v, wantErr %v", err, tt.wantInsertErr)
			}

			query = fmt.Sprintf("Delete from %s Where id = ?", tt.args.table)
			if err = c.BatchExecution(query, [][]interface{}{[]interface{}{tt.args.id}}); (err != nil) != tt.wantDeleteErr {
				t.Errorf("connection.Delete() error = %v, wantErr %v", err, tt.wantDeleteErr)
			}

			c.Close()

			if !c.Closed() {
				t.Errorf("connection.Closed() expected true got false")
			}
		})
	}
}
