package filter

import (
	"reflect"
	"testing"
)

const (
	query1 string = "parnerId=%v"
	val1   string = "f28f9d07-9853-43d2-a701-cb0764d8524b"

	query2 string = "companyName=%v"
	val2   string = "ABCD"

	queryAND string = "AND"
	queryOR  string = "OR"

	queryLimit string = "limit %v"
	limitVal   int    = 10

	queryOrderBy string = "order by %v"
	orderByVal   string = "name"
)

func TestNew(t *testing.T) {
	type args struct {
		q string
		v []interface{}
	}
	tests := []struct {
		name string
		args args
		want *Filter
	}{
		{
			name: "Success",
			args: args{
				q: query1,
				v: []interface{}{val1},
			},
			want: &Filter{
				query:     query1,
				values:    []interface{}{val1},
				ShouldAnd: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.q, tt.args.v...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_Add(t *testing.T) {
	type fields struct {
		query  string
		values []interface{}
	}
	type args struct {
		f *Filter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Filter
	}{
		{
			name: "Success",
			fields: fields{
				query:  query1,
				values: []interface{}{val1},
			},
			args: args{
				f: &Filter{
					query:  query2,
					values: []interface{}{val2},
				},
			},
			want: Filter{
				query:  query1 + " " + query2,
				values: []interface{}{val1, val2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filter{
				query:  tt.fields.query,
				values: tt.fields.values,
			}
			if got := f.Add(tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyWithNewVals(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		initialVals []interface{}
		newVals     []interface{}
	}{
		{
			name: "Empty",
		},
		{
			name:        "One value",
			query:       "hi = ?",
			initialVals: []interface{}{1},
			newVals:     []interface{}{2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialF := New(tt.query, tt.initialVals...)
			newFFromInitial := initialF.CopyWithNewVals(tt.newVals...)
			newF := New(tt.query, tt.newVals...)
			if !reflect.DeepEqual(newF, newFFromInitial) {
				t.Errorf("Copying did not produce the same filter as just creating a filter. New produced %v, but Copy produced %v", newF, newFFromInitial)
			}
			if initialF.GetQuery() != tt.query {
				t.Errorf("Copying did not copy the initial filter's query! Query is now %v, but should have been %v", newF.GetQuery(), tt.query)
			}
			if !reflect.DeepEqual(initialF.GetValues(), tt.initialVals) {
				t.Errorf("Copying changed the initial filter's values! Values are now %v, but should have remained %v", initialF.GetValues(), tt.initialVals)
			}
			if initialF.GetQuery() != tt.query {
				t.Errorf("Copying changed the initial filter's query! Query is now %v, but should have remained %v", initialF.GetQuery(), tt.query)
			}
		})
	}
}

func TestFilter_Limit(t *testing.T) {
	type fields struct {
		query     string
		values    []interface{}
		ShouldAnd bool
	}
	type args struct {
		limitFilter *Filter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Filter
	}{
		{
			name: "Success",
			fields: fields{
				query:  query1,
				values: []interface{}{val1},
			},
			args: args{
				limitFilter: &Filter{
					query:  queryLimit,
					values: []interface{}{limitVal},
				},
			},
			want: Filter{
				query:  query1 + " " + queryLimit,
				values: []interface{}{val1, limitVal},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filter{
				query:     tt.fields.query,
				values:    tt.fields.values,
				ShouldAnd: tt.fields.ShouldAnd,
			}
			if got := f.Limit(tt.args.limitFilter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter.Limit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_OrderBy(t *testing.T) {
	type fields struct {
		query     string
		values    []interface{}
		ShouldAnd bool
	}
	type args struct {
		orderBy *Filter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Filter
	}{
		{
			name: "Success",
			fields: fields{
				query:  query1,
				values: []interface{}{val1},
			},
			args: args{
				orderBy: &Filter{
					query:  queryOrderBy,
					values: []interface{}{orderByVal},
				},
			},
			want: Filter{
				query:  query1 + " " + queryOrderBy,
				values: []interface{}{val1, orderByVal},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filter{
				query:     tt.fields.query,
				values:    tt.fields.values,
				ShouldAnd: tt.fields.ShouldAnd,
			}
			if got := f.OrderBy(tt.args.orderBy); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter.OrderBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
