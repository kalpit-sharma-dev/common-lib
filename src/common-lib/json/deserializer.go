// Package json provides methods for deserializing JSON string to interface{}. This also does error handling at different levels.
package json

import (
	"encoding/json"

	"io"
	"reflect"

	exc "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exception"
)

const (
	//ErrJSONInvalidStream handles error code for invalid stream
	ErrJSONInvalidStream = "JSONInvalidStream"
	//ErrJSONNotAPointerOrNil handles error code if an interface{} is nil or not a pointer
	ErrJSONNotAPointerOrNil = "JSONNotAPointerOrNil"
	//ErrJSONFailedToDeserialize handles error code for deserialization failure
	ErrJSONFailedToDeserialize = "JSONFailedToDeserialize"
	//ErrJSONFailedToSerialize handles error code for marshaling failed
	ErrJSONFailedToSerialize = "ErrJSONFailedToSerialize"
)

// Deserialize reads the JSON from the stream and deserialize into tObject
func Deserialize(tObject interface{}, stream io.Reader) error {
	err := isValidObject(tObject)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(stream)
	if !decoder.More() {
		return exc.New(ErrJSONInvalidStream, nil)
	}

	err = decoder.Decode(&tObject)
	if err != nil {
		return exc.New(ErrJSONFailedToDeserialize, err)
	}

	return nil
}

func isValidObject(tObject interface{}) error {
	val := reflect.ValueOf(tObject)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return exc.New(ErrJSONNotAPointerOrNil, nil)
	}
	return nil
}

// DeserializeBytes function expects input json as transformable json data of map
// return error : incase input is not a json transformable
func DeserializeBytes(input []byte) (interface{}, error) {
	var output interface{}
	err := DeserializeBytesToStruct(input, &output)
	return output, err
}

// DeserializeBytesToStruct function expects input json as transformable json data of map
// return error : incase input is not a json transformable
func DeserializeBytesToStruct(input []byte, output interface{}) error {
	err := isValidObject(output)
	if err != nil {
		return err
	}
	err = json.Unmarshal(input, output)
	if err != nil {
		return err
	}
	return nil
}
