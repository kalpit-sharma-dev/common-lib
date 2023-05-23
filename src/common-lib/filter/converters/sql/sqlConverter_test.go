package sql

import (
	"reflect"
	"strings"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/command"
)

const (
	filterCompanyName        string = "companyName"
	filterCompanyUniqueField string = "id"
	filterPartnerID          string = "partnerId"
	colCompanyName           string = "name"
	colPartnerID             string = "partner_id"
	testCompany              string = "test_company"
	testCompanies            string = "company1,company2"

	filterCompanyLocation string = "address.country"
	colCountry            string = "address_country"
	testCountry           string = "\"India\""

	partnerID string = "00112233-4455-6677-8899-aabbccddeeff"
	companyID string = "00332222-4455-6677-8899-aabbccddeeff"
)

var testMap = map[string]string{
	filterCompanyUniqueField: filterCompanyUniqueField,
	filterCompanyName:        colCompanyName,
	filterPartnerID:          colPartnerID,
	filterCompanyLocation:    colCountry,
}

func mapper(key string) string {
	return testMap[key]
}

func TestGetConverter(t *testing.T) {
	tests := []struct {
		name string
		want *SQLconverter
	}{
		{
			name: "Success",
			want: &SQLconverter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConverter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSQLconverter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLconverter_DoForCommandWithoutValue(t *testing.T) {

	type args struct {
		c      command.Command
		mapper func(string) string
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
				c:      command.New("", string(command.Or), ""),
				mapper: nil,
			},
			want:    filter.New("or"),
			wantErr: false,
		},
		{
			name: "Success_using_mapper",
			args: args{
				c:      command.New("foo", string(command.Null), ""),
				mapper: strings.ToUpper,
			},
			want:    filter.New("FOO is null"),
			wantErr: false,
		},
		{
			name: "No operator match for UNKNOWN operator in filter",
			args: args{
				c:      command.New("", "UNKNOWN", ""),
				mapper: nil,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SQLconverter{}
			got, err := s.DoForCommandWithoutValue(tt.args.c, tt.args.mapper)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLconverter.DoForCommandWithoutValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLconverter.DoForCommandWithoutValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLconverter_DoForCommandWithValue(t *testing.T) {
	type fields struct {
		columnMapper func(string) string
	}
	type args struct {
		c command.Command
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *filter.Filter
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				columnMapper: mapper,
			},
			args: args{
				c: command.New(filterCompanyName, string(command.Eq), testCompany),
			},
			want:    filter.New("(name) = (%v)", "test_company"),
			wantErr: false,
		},
		{
			name: "Success_Multi_Columns",
			fields: fields{
				columnMapper: mapper,
			},
			args: args{
				c: command.New(strings.Join([]string{filterCompanyName, filterCompanyLocation}, ","), string(command.Gt), testCompany),
			},
			want:    filter.New("(name,address_country) > (%v,%v)", testCompany),
			wantErr: false,
		},
		{
			name: "Success | Like",
			fields: fields{
				columnMapper: mapper,
			},
			args: args{
				c: command.New(filterCompanyName, string(command.Like), testCompany),
			},
			want:    filter.New("name  like  %v", "%test_company%"),
			wantErr: false,
		},
		{
			name: "Success | IN",
			fields: fields{
				columnMapper: mapper,
			},
			args: args{
				c: command.New(filterCompanyName, string(command.In), testCompanies),
			},
			want:    filter.New("name in (%v,%v)", "company1", "company2"),
			wantErr: false,
		},
		{
			name: "Success | Nested field",
			fields: fields{
				columnMapper: mapper,
			},
			args: args{
				c: command.New(filterCompanyLocation, string(command.Eq), testCountry),
			},
			want:    filter.New("(address_country) = (%v)", "India"),
			wantErr: false,
		},
		{
			name: "No property match",
			fields: fields{
				columnMapper: mapper,
			},
			args: args{
				c: command.New("UNKNOWN", string(command.Eq), testCountry),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No operator match",
			fields: fields{
				columnMapper: mapper,
			},
			args: args{
				c: command.New(filterCompanyName, "UNKNOWN", testCountry),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SQLconverter{}
			got, err := s.DoForCommandWithValue(tt.args.c, tt.fields.columnMapper)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLconverter.DoForCommandWithValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLconverter.DoForCommandWithValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLconverter_AND(t *testing.T) {
	s := &SQLconverter{}
	tests := []struct {
		name     string
		operands []*filter.Filter
		want     *filter.Filter
	}{
		{
			name: "success_with_two_operands",
			operands: []*filter.Filter{
				filter.New("hi", 1),
				filter.New("bye", 2, 3),
			},
			want: filter.New(
				"hi AND bye",
				1, 2, 3,
			),
		},
		{
			name: "success_with_three_operands",
			operands: []*filter.Filter{
				filter.New("hi", 4),
				filter.New("bye"),
				filter.New("why", 5),
			},
			want: filter.New(
				"hi AND bye AND why",
				4, 5,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.AND(tt.operands...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLconverter.AND(...) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLconverter_OR(t *testing.T) {
	s := &SQLconverter{}
	tests := []struct {
		name     string
		operands []*filter.Filter
		want     *filter.Filter
	}{
		{
			name: "success_with_two_operands",
			operands: []*filter.Filter{
				filter.New("hi", 1),
				filter.New("bye", 2, 3),
			},
			want: filter.New(
				"hi OR bye",
				1, 2, 3,
			),
		},
		{
			name: "success_with_three_operands",
			operands: []*filter.Filter{
				filter.New("hi", 4),
				filter.New("bye"),
				filter.New("why", 5),
			},
			want: filter.New(
				"hi OR bye OR why",
				4, 5,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.OR(tt.operands...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLconverter.OR(...) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppendFilterToWhereClause(t *testing.T) {

	type args struct {
		query  string
		filter *filter.Filter
		args   []interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []interface{}
	}{
		{
			name: "success",
			args: args{
				query:  "select * from company where partner_id=$1;",
				filter: filter.New("company_id=%v", companyID),
				args:   []interface{}{partnerID},
			},
			want:  "select * from company where partner_id=$1 AND company_id=$2;",
			want1: []interface{}{partnerID, companyID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := AppendFilterToWhereClause(tt.args.query, tt.args.filter, tt.args.args)
			if got != tt.want {
				t.Errorf("AppendFilterToWhereClause() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("AppendFilterToWhereClause() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSQLconverter_GetLimitFilter(t *testing.T) {
	type args struct {
		limit int
	}
	tests := []struct {
		name    string
		s       *SQLconverter
		args    args
		want    *filter.Filter
		wantErr bool
	}{
		{
			name: "success",
			s:    &SQLconverter{},
			args: args{
				limit: 10,
			},
			want: func() *filter.Filter {
				f := filter.New("limit %v", 10)
				f.ShouldAnd = false
				return f
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SQLconverter{}
			got, err := s.GetLimitFilter(tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLconverter.GetLimitFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.GetQuery(), tt.want.GetQuery()) {
				t.Errorf("SQLconverter.GetLimitFilter() = %v, want %v", got, tt.want)
			}
			// TODO: Need to check for filter value equality as well
		})
	}
}

func TestSQLconverter_GetOrderByFilter(t *testing.T) {
	type args struct {
		field  string
		mapper func(string) string
	}
	tests := []struct {
		name    string
		s       *SQLconverter
		args    args
		want    *filter.Filter
		wantErr bool
	}{
		{
			name: "success",
			s:    &SQLconverter{},
			args: args{
				field:  filterCompanyName,
				mapper: mapper,
			},
			want:    filter.New("order by " + colCompanyName),
			wantErr: false,
		},
		{
			name: "success_with_sortOrder",
			s:    &SQLconverter{},
			args: args{
				field:  filterCompanyName + " asc",
				mapper: mapper,
			},
			want:    filter.New("order by " + colCompanyName + " asc"),
			wantErr: false,
		},
		{
			name: "success_with_uniqueField",
			s:    &SQLconverter{},
			args: args{
				field:  filterCompanyName + "," + filterCompanyUniqueField,
				mapper: mapper,
			},
			want:    filter.New("order by " + colCompanyName + ", " + filterCompanyUniqueField),
			wantErr: false,
		},
		{
			name: "success_with_sortOrder_and_uniqueField",
			s:    &SQLconverter{},
			args: args{
				field:  filterCompanyName + "," + filterCompanyUniqueField + " asc",
				mapper: mapper,
			},
			want:    filter.New("order by " + colCompanyName + " asc, " + filterCompanyUniqueField + " asc"),
			wantErr: false,
		},
		{
			name: "err_when_providing_invalid_field",
			s:    &SQLconverter{},
			args: args{
				field:  "unexpectedField," + filterCompanyUniqueField,
				mapper: mapper,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SQLconverter{}
			got, err := s.GetOrderByFilter(tt.args.field, tt.args.mapper)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLconverter.GetOrderByFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && !reflect.DeepEqual(got.GetQuery(), tt.want.GetQuery()) {
				t.Errorf("SQLconverter.GetOrderByFilter() = %v, want %v", got, tt.want)
			}
			// TODO: Need to ceck for filter value equality as well
		})
	}
}
