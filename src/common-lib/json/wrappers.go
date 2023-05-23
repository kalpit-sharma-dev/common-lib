package json

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"

	errorCodes "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/errorCodePair"
	exc "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exception"
)

//go:generate mockgen -package mock -destination=mock/mocks.go . FactoryJSON,DeserializerJSON,SerializerJSON

const (
	//ErrJSONEmptyFilePath error code for empty file path
	ErrJSONEmptyFilePath = "JSONEmptyFilePath"
	//ErrJSONInvalidFilePathOrUnableToRead error code for invalid file path or unable to read file content
	ErrJSONInvalidFilePathOrUnableToRead = "JSONInvalidFilePathOrUnableToRead"
	//ErrJSONFilePermissionDenied error code for file with access issue
	ErrJSONFilePermissionDenied = "JSONFilePermissionDenied"
	//ErrJSONFileNotFound error code for file does not exist
	ErrJSONFileNotFound = "JSONFileNotFound"
	//ErrJSONBlankString error code for blank string to deserialize
	ErrJSONBlankString = "JSONBlankString"
	//ErrJSONFileCreateError error code for an error while creating file on disk
	ErrJSONFileCreateError = "ErrJSONFileCreateError"
	//ErrFlushData error code for an error while creating file on disk
	ErrFlushData = "ErrFlushData"
)

// FactoryJSON is a top level interface which returns Serializer & Deserializer interface
type FactoryJSON interface {
	GetDeserializerJSON() DeserializerJSON
	GetSerializerJSON() SerializerJSON
}

// FactoryJSONImpl is an implementation of FactoryJSON interface
type FactoryJSONImpl struct{}

// GetDeserializerJSON returns the Deserializer interface
func (f FactoryJSONImpl) GetDeserializerJSON() DeserializerJSON {
	return deserializerJSONImpl{}
}

// GetSerializerJSON returns the Serializer interface
func (f FactoryJSONImpl) GetSerializerJSON() SerializerJSON {
	return serializerJSONImpl{}
}

// SerializerJSON interface exposes the serialization methods
type SerializerJSON interface {
	WriteFile(filePath string, tObject interface{}) error
	WriteByteStream(tObject interface{}) ([]byte, error)
	Write(w io.Writer, tObject interface{}) error
}

// DeserializerJSON interface has the deserialization methods
type DeserializerJSON interface {
	ReadFile(tObject interface{}, filePath string) error
	ReadString(tObject interface{}, data string) error
}

// deserializerJSONImpl is the implementation of Deserializer interface
type deserializerJSONImpl struct{}

// ReadFile deserializes the file content from a path specified
func (deserializerJSONImpl) ReadFile(tObject interface{}, filePath string) error {
	if filePath == "" {
		return exc.New(ErrJSONEmptyFilePath, nil)
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if isOsError, osErrorMessage := GetOsErrorMessage(err); isOsError {
			return exc.New(osErrorMessage, err)
		}
		return exc.New(ErrJSONInvalidFilePathOrUnableToRead, err)
	}
	return Deserialize(tObject, bytes.NewReader(data))
}

// ReadString deserializes the string data passed in
func (deserializerJSONImpl) ReadString(tObject interface{}, data string) error {
	if data == "" {
		return exc.New(ErrJSONBlankString, nil)
	}
	return Deserialize(tObject, strings.NewReader(data))
}

// changes for agent autoupdate error standardization START HERE. To be refactored as per common-lib standards for comming rollouts

func GetOsErrorMessage(err error) (isOsError bool, osErrorMsg string) {
	switch err.(type) {
	case *os.PathError, *os.SyscallError, *os.LinkError:
		if os.IsNotExist(err) {
			return true, ErrJSONFileNotFound
		} else if os.IsPermission(err) {
			return true, ErrJSONFilePermissionDenied
		}
	}
	return
}

// Determine error code pairs for Json Deserialize failures
func DetermineJsonDeserializeErrors(err error) (mainErrorCode, subErrorCode string) {

	if err.Error() == ErrJSONFilePermissionDenied {
		mainErrorCode, subErrorCode = errorCodes.FileSystem, errorCodes.AccessDenied
	} else if err.Error() == ErrJSONFileNotFound {
		mainErrorCode, subErrorCode = errorCodes.FileSystem, errorCodes.FileNotFound
	} else {
		mainErrorCode, subErrorCode = errorCodes.Internal, errorCodes.Operational
	}
	return
}

// changes for agent autoupdate error standardization END HERE

// serializerJSONImpl is the implementation of Serializer interface
type serializerJSONImpl struct{}

// WriteFile serializes object into the specified file
func (serializerJSONImpl) WriteFile(filePath string, tObject interface{}) error {
	if filePath == "" {
		return exc.New(ErrJSONEmptyFilePath, nil)
	}

	fp, err := os.Create(filePath)
	if err != nil {
		return exc.New(ErrJSONFileCreateError, err)
	}
	return Serialize(fp, tObject)
}

// WriteByteStream serializes object into the specified file
func (serializerJSONImpl) WriteByteStream(tObject interface{}) ([]byte, error) {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := Serialize(w, tObject)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, exc.New(ErrFlushData, err)
	}
	return b.Bytes(), nil
}

// WriteByteStream serializes object into the specified file
func (serializerJSONImpl) Write(w io.Writer, tObject interface{}) error {
	err := Serialize(w, tObject)
	if err != nil {
		return err
	}
	return nil
}
