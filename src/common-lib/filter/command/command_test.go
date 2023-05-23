package command

import (
	"reflect"
	"testing"

	"github.com/golang-collections/collections/stack"
	"github.com/golang/mock/gomock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter"
)

const (
	filterCompanyName string = "companyName"
	filterPartnerID   string = "partnerId"

	colCompanyName string = "name"
	colPartnerID   string = "partner_id"

	testCompany string = "test_company"

	testQuery1 = "name=%v"
	testVal1   = "ABCD"

	testQuery2 = "or"
)

var testMap = map[string]string{
	filterCompanyName: colCompanyName,
	filterPartnerID:   colPartnerID,
}

func mapper(key string) string {
	return testMap[key]
}

var (
	ctrl       *gomock.Controller
	mconverter *MockConverter
)

func setup(t *testing.T) {
	ctrl = gomock.NewController(t)
	defer ctrl.Finish()
}

func TestNew(t *testing.T) {
	type args struct {
		property string
		operator string
		value    string
	}
	tests := []struct {
		name string
		args args
		want Command
	}{
		{
			name: "Success",
			args: args{
				property: filterCompanyName,
				operator: string(Eq),
				value:    testCompany,
			},
			want: Command{
				property: filterCompanyName,
				operator: string(Eq),
				value:    testCompany,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.property, tt.args.operator, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommand_Accept(t *testing.T) {
	setup(t)

	type fields struct {
		property string
		operator string
		value    string
	}
	type args struct {
		converter Converter
		mapper    func(string) string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *filter.Filter
		wantErr bool
	}{
		{
			name: "Success | Command with value",
			fields: fields{
				property: filterCompanyName,
				operator: string(Eq),
				value:    testCompany,
			},
			args: args{
				converter: func() Converter {
					mc := NewMockConverter(ctrl)
					f := filter.New(testQuery1, testVal1)
					mc.EXPECT().DoForCommandWithValue(gomock.Any(), gomock.Any()).Return(f, nil)
					mc.EXPECT().DoForCommandWithoutValue(gomock.Any(), gomock.Any()).Return(nil, nil)
					return mc
				}(),
			},
			want: filter.New(testQuery1, testVal1),
		},
		{
			name: "Success | Command Without Value",
			fields: fields{
				property: filterCompanyName,
				operator: string(Eq),
				value:    testCompany,
			},
			args: args{
				converter: func() Converter {
					mc := NewMockConverter(ctrl)
					f := filter.New(testQuery2, nil)
					mc.EXPECT().DoForCommandWithoutValue(gomock.Any(), gomock.Any()).Return(f, nil)
					return mc
				}(),
				mapper: mapper,
			},
			want: filter.New(testQuery2, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Command{
				property: tt.fields.property,
				operator: tt.fields.operator,
				value:    tt.fields.value,
			}
			got, err := c.Accept(tt.args.converter, tt.args.mapper)
			if (err != nil) != tt.wantErr {
				t.Errorf("Command.Accept() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Command.Accept() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCommandWrapperWithValidation(t *testing.T) {
	var brackets *stack.Stack
	type args struct {
		words []string
	}
	tests := []struct {
		name    string
		args    args
		mocker  func()
		want    *Command
		want1   int
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				words: []string{"partnerId", "eq", "ABCD", ")"},
			},
			want: func() *Command {
				cmd := New("partnerId", "eq", "ABCD")
				return &cmd
			}(),
			want1:   3,
			wantErr: false,
		},
		{
			name: "Success | value with spaces",
			args: args{
				words: []string{"partnerId", "eq", "ABCD PQRST", ")"},
			},
			want: func() *Command {
				cmd := New("partnerId", "eq", "ABCD PQRST")
				return &cmd
			}(),
			want1:   3,
			wantErr: false,
		},
		{
			name: "Success | filter with just space",
			args: args{
				words: []string{blank},
			},
			want:    nil,
			want1:   1,
			wantErr: false,
		},

		{
			name: "LHS handling",
			args: args{
				words: []string{"(", "partnerId", "eq", "ABCD", ")"},
			},
			want: func() *Command {
				cmd := New("", "(", "")
				return &cmd
			}(),
			want1:   1,
			wantErr: false,
		},
		{
			name: "Missing brackets - RHS without LHS",
			args: args{
				words: []string{")"},
			},
			want:    nil,
			want1:   0,
			wantErr: true,
		},
		{
			name: "Error due to improper termination of filter query",
			args: args{
				words: []string{"partnerId", "eq", "ABCD"},
			},
			want:    nil,
			want1:   0,
			wantErr: true,
		},
		{
			name: "HandleBrackets | check LHS and RHS parity",
			args: args{
				words: []string{")"},
			},
			want: func() *Command {
				cmd := New("", ")", "")
				return &cmd
			}(),
			want1:   1,
			wantErr: false,
			mocker: func() {
				GetCommandWrapperWithValidation([]string{"("}, brackets)
			},
		},
		{
			name: "Success | filter with spaces",
			args: args{
				words: []string{"partnerId", blank, "eq", blank, "ABCD", ")"},
			},
			want: func() *Command {
				cmd := New("partnerId", "eq", "ABCD")
				return &cmd
			}(),
			want1:   5,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brackets = stack.New()
			if tt.mocker != nil {
				tt.mocker()
			}
			got, got1, err := GetCommandWrapperWithValidation(tt.args.words, brackets)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommandWrapperWithValidation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCommandWrapperWithValidation() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetCommandWrapperWithValidation() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
