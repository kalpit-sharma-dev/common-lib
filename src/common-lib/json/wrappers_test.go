package json

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	errorCodes "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/errorCodePair"
	exc "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exception"
)

func TestReadFileSuccess(t *testing.T) {
	var conf Kafkaconfig
	err := FactoryJSONImpl{}.GetDeserializerJSON().ReadFile(&conf, "test.txt")
	if err != nil {
		t.Errorf("expecting no error, got %v", err)
	}
}

func TestReadFileBlankFilePath(t *testing.T) {
	var conf Kafkaconfig
	err := FactoryJSONImpl{}.GetDeserializerJSON().ReadFile(&conf, "")
	exce, ok := err.(exc.Exception)
	if ok {
		if exce.GetErrorCode() != ErrJSONEmptyFilePath {
			t.Errorf("Expected JSONEmptyFilePath but got %v", err)
		}
	} else {
		t.Error("Expecting Exception type")
	}
}

func TestReadFileWrongPath(t *testing.T) {
	var conf Kafkaconfig
	err := FactoryJSONImpl{}.GetDeserializerJSON().ReadFile(&conf, "wrapper_test.json")
	exce, ok := err.(exc.Exception)
	if ok {
		if exce.GetErrorCode() != ErrJSONFileNotFound {
			t.Errorf("Expected JSONInvalidFilePathOrUnableToRead but got %v", exce)
		}
	} else {
		t.Error("Expecting Exception type")
	}
}

func TestReadStringSuccess(t *testing.T) {
	var conf Kafkaconfig
	err := FactoryJSONImpl{}.GetDeserializerJSON().ReadString(&conf, "{}")
	if err != nil {
		t.Errorf("expecting no error, got %v", err)
	}
}

func TestReadStringBlank(t *testing.T) {
	var conf Kafkaconfig
	err := FactoryJSONImpl{}.GetDeserializerJSON().ReadString(&conf, "")
	exce, ok := err.(exc.Exception)
	if ok {
		if exce.GetErrorCode() != ErrJSONBlankString {
			t.Errorf("Expected ErrJSONBlankString but got %v", exce)
		}
	} else {
		t.Error("Expecting Exception type")
	}
}

func TestWriteFile(t *testing.T) {
	type test struct {
	}
	err := FactoryJSONImpl{}.GetSerializerJSON().WriteFile("test.txt", &test{})
	if err != nil {
		t.Error("Unexpected error")
	}
}

func TestWriteFileEmptyFile(t *testing.T) {
	type test struct {
	}
	err := serializerJSONImpl{}.WriteFile("", &test{})
	if err == nil || !strings.HasPrefix(err.Error(), "JSONEmptyFilePath") {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestWriteByteStream(t *testing.T) {
	type test struct {
		s string
	}
	var a = "initial"
	var abyte = []byte(`initial`)

	bStream, err := serializerJSONImpl{}.WriteByteStream(&a)
	if err != nil {
		t.Error("Unexpected error")
	}
	fmt.Println(bStream)
	fmt.Println(a == string(bStream))
	//TODO comparison should be not equal to
	if bytes.Compare(bStream, abyte) == 0 {
		// a less b
		t.Errorf("Unexpected error ... %v", string(bStream))
	}
	if string(bStream) == "initial" {
		// a less b
		t.Errorf("Unexpected string returned, Expected %s, Returned %s", a, string(bStream))
	}
}

func TestWrite(t *testing.T) {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := serializerJSONImpl{}.Write(w, Kafkaconfig{KafkaAddress: []string{"abc", "pqr"}})
	if err != nil {
		t.Errorf("error not expected, got %v", err)
	}
}

func TestGetOsErrorMessage(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name           string
		args           args
		wantIsOsError  bool
		wantOsErrorMsg string
	}{
		{
			name:           "underlying error is an OS error",
			args:           args{err: &os.PathError{Op: "open", Path: "path", Err: os.ErrPermission}},
			wantIsOsError:  true,
			wantOsErrorMsg: ErrJSONFilePermissionDenied,
		},
		{
			name:           "underlying error is NOT an OS error",
			args:           args{err: errors.New("Not OS error")},
			wantIsOsError:  false,
			wantOsErrorMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsOsError, gotOsErrorMsg := GetOsErrorMessage(tt.args.err)
			if gotIsOsError != tt.wantIsOsError {
				t.Errorf("GetOsErrorMessage() gotIsOsError = %v, want %v", gotIsOsError, tt.wantIsOsError)
			}
			if gotOsErrorMsg != tt.wantOsErrorMsg {
				t.Errorf("GetOsErrorMessage() gotOsErrorMsg = %v, want %v", gotOsErrorMsg, tt.wantOsErrorMsg)
			}
		})
	}
}

func TestDetermineJsonDeserializeErrors(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name              string
		args              args
		wantMainErrorCode string
		wantSubErrorCode  string
	}{
		{
			name:              "when error is error is an Access issue",
			args:              args{err: errors.New(ErrJSONFilePermissionDenied)},
			wantMainErrorCode: errorCodes.FileSystem,
			wantSubErrorCode:  errorCodes.AccessDenied,
		},
		{
			name:              "when error is error is an JSON File Not Found",
			args:              args{err: errors.New(ErrJSONFileNotFound)},
			wantMainErrorCode: errorCodes.FileSystem,
			wantSubErrorCode:  errorCodes.FileNotFound,
		},
		{
			name:              "when error is error is an unidentified issue",
			args:              args{err: errors.New("unidentified")},
			wantMainErrorCode: errorCodes.Internal,
			wantSubErrorCode:  errorCodes.Operational,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMainErrorCode, gotSubErrorCode := DetermineJsonDeserializeErrors(tt.args.err)
			if gotMainErrorCode != tt.wantMainErrorCode {
				t.Errorf("DetermineJsonDeserializeErrors() gotMainErrorCode = %v, want %v", gotMainErrorCode, tt.wantMainErrorCode)
			}
			if gotSubErrorCode != tt.wantSubErrorCode {
				t.Errorf("DetermineJsonDeserializeErrors() gotSubErrorCode = %v, want %v", gotSubErrorCode, tt.wantSubErrorCode)
			}
		})
	}
}
