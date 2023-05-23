// Package json provides methods for serializing interface{} to JSON.
package json

import (
	"encoding/json"
	"io"

	exc "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exception"
)

// Serialize serializes the given interface
func Serialize(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "   ")
	err := enc.Encode(v)
	if err != nil {
		return exc.New(ErrJSONFailedToSerialize, err)
	}
	return nil
}
