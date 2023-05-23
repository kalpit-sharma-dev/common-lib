package json

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	jsonpatch "github.com/evanphx/json-patch"
)

// patches should be in format described in RFC6902
// merging applies according RFC7396

// PatchFromFiles reads patches from file src and appying it to dest file
func PatchFromFiles(src, dest string) error {
	original, patch, err := openFiles(src, dest)
	if err != nil {
		return err
	}

	modified, err := Patch(original, patch)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, modified, 0766)
	if err != nil {
		return fmt.Errorf("error while writing modified JSON to file: %s with error: %s", dest, err)
	}

	return nil
}

// Patch reads patches from src and appying it to dest
func Patch(src, dest io.Reader) ([]byte, error) {

	srcBuf, destBuf := new(bytes.Buffer), new(bytes.Buffer)

	_, err := srcBuf.ReadFrom(src)
	if err != nil {
		return nil, err
	}

	_, err = destBuf.ReadFrom(dest)
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.DecodePatch(srcBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error while decoding patch JSON: %s with error: %s", dest, err)
	}

	modified, err := patch.ApplyIndent(destBuf.Bytes(), "    ")
	if err != nil {
		return nil, fmt.Errorf("error while applying patch: %v with error: %s", patch, err)
	}

	return modified, err
}

func openFiles(src, dest string) (original, patch io.Reader, err error) {
	original, err = os.Open(dest)
	if err != nil {
		return original, patch, fmt.Errorf("error reading file: %s with error: %s", dest, err)
	}
	patch, err = os.Open(src)
	if err != nil {
		return original, patch, fmt.Errorf("error reading file: %s with error: %s", src, err)
	}
	return
}
