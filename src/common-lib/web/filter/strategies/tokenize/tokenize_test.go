package tokenize

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/command"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/converters/sql"
)

const (
	testQuery1 = "partner_id=%v"
	testVal1   = "ABCD"
)

const (
	filterCompanyName string = "companyName"
	filterPartnerID   string = "partnerId"
	colCompanyName    string = "name"
	colPartnerID      string = "partner_id"
	testCompany       string = "test_company"

	filterCompanyLocation string = "address.country"
	tableAddress                 = "address"
	colCountry                   = "country"
	filterAddress                = "address"
	filterCountry                = "country"
	testCountry                  = "\"India\""
)

var testMap = map[string]string{
	filterCompanyName: colCompanyName,
	filterPartnerID:   colPartnerID,
	filterAddress:     tableAddress,
	filterCountry:     colCountry,
}

func mapper(key string) string {
	return testMap[key]
}

var (
	ctrl *gomock.Controller
)

func setup(t *testing.T) {
	ctrl = gomock.NewController(t)
	defer ctrl.Finish()
}

func TestGetTokenStrategy(t *testing.T) {
	tests := []struct {
		name string
		want *Strategy
	}{
		{
			name: "Success",
			want: &Strategy{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStrategy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenStrategy_Parse(t *testing.T) {

	setup(t)

	type args struct {
		cnv    command.Converter
		filter string
	}
	tests := []struct {
		name    string
		args    args
		want    *filter.Filter
		wantErr bool
	}{
		{
			name: "Success_IN_LIKE",
			args: args{
				cnv:    sql.GetConverter(),
				filter: "(partnerId IN 1000,2000,3000,4000 OR partnerId = 5000 AND partnerId : 6000)",
			},
			want: filter.New(
				" ( ( partner_id in (%v,%v,%v,%v) or (partner_id) = (%v) and partner_id  like  %v ) )",
				"1000", "2000", "3000", "4000", "5000", "%6000%",
			),
			wantErr: false,
		},
		{
			name: "Success_AND_OR_Equal_NotEqual",
			args: args{
				cnv:    sql.GetConverter(),
				filter: "(partnerId = \"ABCD\" OR partnerId = \"XYZ\") AND (companyName = \"ABCD\" OR companyName != \"XYZ\")",
			},
			want: filter.New(
				" ( ( (partner_id) = (%v) or (partner_id) = (%v) ) and ( (name) = (%v) or (name) != (%v) ) )",
				"ABCD", "XYZ", "ABCD", "XYZ",
			),
			wantErr: false,
		},
		{
			name: "Operator case sensitivity",
			args: args{
				cnv:    sql.GetConverter(),
				filter: "(partnerId = \"ABCD\" OR partnerId = \"XYZ\") and (companyName = \"ABCD\" OR companyName != \"XYZ\")",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No LHS present for RHS",
			args: args{
				cnv:    sql.GetConverter(),
				filter: "partnerId = \"ABCD\")",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error | Extra bracket present",
			args: args{
				cnv:    sql.GetConverter(),
				filter: "(partnerId = \"ABCD\")(",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error | Unknown field",
			args: args{
				cnv:    sql.GetConverter(),
				filter: "(UNKNOWN = \"ABCD\")(",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error | Unknown operator",
			args: args{
				cnv:    sql.GetConverter(),
				filter: "(partnerId UNKNOWN \"ABCD\")(",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := Strategy{}
			got, err := tr.Parse(tt.args.cnv, tt.args.filter, mapper)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenStrategy.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got.GetQuery(), tt.want.GetQuery()) {
				t.Errorf("TokenStrategy.Parse() = %v, want %v", got, tt.want)
			}
			if err == nil && !reflect.DeepEqual(got.GetValues(), tt.want.GetValues()) {
				t.Errorf("TokenStrategy.Parse() = %v, want %v", got.GetValues(), tt.want.GetValues())
			}
		})
	}
}
