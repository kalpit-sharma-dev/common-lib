package pagination

import "testing"

var (
	companyID   = "01d3b488-b123-4411-8cd4-7955d4b5a412"
	companyName = "avengers4"
	b64Cursor   = "eyJ1bmlxdWVJZCI6IjAxZDNiNDg4LWIxMjMtNDQxMS04Y2Q0LTc5NTVkNGI1YTQxMiIsIm9yZGVyaW5nS2V5IjoiYXZlbmdlcnM0In0="
)

var testCursor = Cursor{
	UniqueID:    companyID,
	OrderingKey: companyName,
}

func TestCursor_Encode(t *testing.T) {
	type fields struct {
		UniqueID    string
		OrderingKey string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				UniqueID:    companyID,
				OrderingKey: companyName,
			},
			want:    b64Cursor,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cursor{
				UniqueID:    tt.fields.UniqueID,
				OrderingKey: tt.fields.OrderingKey,
			}
			got, err := c.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Cursor.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Cursor.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCursor_Decode(t *testing.T) {
	type fields struct {
		UniqueID    string
		OrderingKey string
	}
	type args struct {
		encoded string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    Cursor
	}{
		{
			name: "success",
			fields: fields{
				UniqueID:    "",
				OrderingKey: "",
			},
			args: args{
				encoded: b64Cursor,
			},
			wantErr: false,
			want:    testCursor,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cursor{
				UniqueID:    tt.fields.UniqueID,
				OrderingKey: tt.fields.OrderingKey,
			}
			if err := c.Decode(tt.args.encoded); (err != nil) != tt.wantErr {
				t.Errorf("Cursor.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if *c != tt.want {
				t.Errorf("Cursor.Decode() got = %v, want %v", c, tt.want)
			}
		})
	}
}
