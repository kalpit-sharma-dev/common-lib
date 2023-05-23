package cql

import (
	"testing"

	"github.com/gocql/gocql"
)

func TestCreateSession(t *testing.T) {
	type args struct {
		cluster *gocql.ClusterConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Nil Config", wantErr: true, args: args{cluster: nil}},
		{name: "Blank Config", wantErr: true, args: args{cluster: &gocql.ClusterConfig{}}},
		{name: "Blank Host", wantErr: true, args: args{&gocql.ClusterConfig{Hosts: []string{}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateSession(tt.args.cluster)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
