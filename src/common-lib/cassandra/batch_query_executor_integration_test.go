// +build integration

package cassandra

import (
	"testing"
)

func TestGetBatchQueryExecutor(t *testing.T) {
	type args struct {
		conf *DbConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Nil Config", wantErr: true},
		{name: "Blank Config", wantErr: true, args: args{conf: &DbConfig{}}},
		{name: "Blank Host", wantErr: true, args: args{conf: &DbConfig{Hosts: []string{}}}},
		{name: "Blank Keyspace", wantErr: true, args: args{conf: &DbConfig{Hosts: []string{"Sever"}, Keyspace: ""}}},
		{name: "Invalid Keyspace", wantErr: true, args: args{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "???"}}},
		{name: "Valid Keyspace", wantErr: false, args: args{conf: &DbConfig{Hosts: cassandraHosts, Keyspace: "system"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetBatchQueryExecutor(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBatchQueryExecutor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
