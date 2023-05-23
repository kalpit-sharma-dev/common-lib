package checksum

import (
	"reflect"
	"testing"
)

func TestGetType(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name string
		args args
		want Type
	}{
		{name: "blank", want: NONE, args: args{data: ""}},
		{name: "none", want: NONE, args: args{data: "none"}},
		{name: "md5", want: MD5, args: args{data: "MD5"}},
		{name: "SHA1", want: SHA1, args: args{data: "SHA1"}},
		{name: "SHA256", want: SHA256, args: args{data: "SHA256"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetType(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestType_UnmarshalJSON(t *testing.T) {
	type fields struct {
		order int
		Name  string
		value string
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    Type
	}{
		{name: "Default-Name", fields: fields{}, args: args{[]byte("")}, wantErr: false, want: NONE},
		{name: "Given Name - NONE", fields: fields{}, args: args{[]byte("\"NONE\"")}, wantErr: false, want: NONE},
		{name: "Given Name - NONE", fields: fields{}, args: args{[]byte("\"none\"")}, wantErr: false, want: NONE},
		{name: "Given Name - MD5", fields: fields{}, args: args{[]byte("\"MD5\"")}, wantErr: false, want: MD5},
		{name: "Given Name - MD5", fields: fields{}, args: args{[]byte("\"md5\"")}, wantErr: false, want: MD5},
		{name: "Given Name - SHA1", fields: fields{}, args: args{[]byte("\"SHA1\"")}, wantErr: false, want: SHA1},
		{name: "Given Name - SHA1", fields: fields{}, args: args{[]byte("\"sha1\"")}, wantErr: false, want: SHA1},
		{name: "Given Name - SHA256", fields: fields{}, args: args{[]byte("\"SHA256\"")}, wantErr: false, want: SHA256},
		{name: "Given Name - SHA256", fields: fields{}, args: args{[]byte("\"sha256\"")}, wantErr: false, want: SHA256},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := &Type{
				order: tt.fields.order,
				Name:  tt.fields.Name,
				value: tt.fields.value,
			}
			if err := typ.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Type.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if *typ != tt.want {
				t.Errorf("Type.UnmarshalJSON() = %v, want %v", typ, tt.want)
			}
		})
	}
}

func TestType_MarshalJSON(t *testing.T) {
	type fields struct {
		order int
		Name  string
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{name: "NONE", fields: fields{value: NONE.value}, wantErr: false, want: []byte("\"NONE\"")},
		{name: "MD5", fields: fields{value: MD5.value}, wantErr: false, want: []byte("\"MD5\"")},
		{name: "SHA1", fields: fields{value: SHA1.value}, wantErr: false, want: []byte("\"SHA1\"")},
		{name: "SHA256", fields: fields{value: SHA256.value}, wantErr: false, want: []byte("\"SHA256\"")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := Type{
				order: tt.fields.order,
				Name:  tt.fields.Name,
				value: tt.fields.value,
			}
			got, err := typ.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Type.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Type.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetService(t *testing.T) {
	type args struct {
		cType Type
	}
	tests := []struct {
		name    string
		args    args
		want    Service
		wantErr bool
	}{
		{name: "Default", args: args{}, wantErr: true, want: nil},
		{name: "NONE", args: args{cType: NONE}, wantErr: false, want: none{}},
		{name: "MD5", args: args{cType: MD5}, wantErr: false, want: md5Impl{}},
		{name: "SHA1", args: args{cType: SHA1}, wantErr: false, want: sha1Impl{}},
		{name: "SHA256", args: args{cType: SHA256}, wantErr: false, want: sha256Impl{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetService(tt.args.cType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetService() = %v, want %v", got, tt.want)
			}
		})
	}
}
