package json

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestPatchFromFiles(t *testing.T) {
	type args struct {
		scenario      string
		patchPath     string
		originalPath  string
		errorExpected bool
	}
	tests := []args{
		{
			scenario:     "success",
			patchPath:    "./patch.json",
			originalPath: "./original.json",
		},
		{
			scenario:      "wrong-source",
			patchPath:     "./wrong-patch.json",
			originalPath:  "./original.json",
			errorExpected: true,
		},
		{
			scenario:      "wrong-dest",
			patchPath:     "./patch.json",
			originalPath:  "./wrong-original.json",
			errorExpected: true,
		},
		{
			scenario:      "invalid-source",
			patchPath:     "./invalid-patch.json",
			originalPath:  "./original.json",
			errorExpected: true,
		},
	}

	randSrc := rand.NewSource(time.Now().UnixNano())
	random := rand.New(randSrc)
	for _, test := range tests {
		test := test
		t.Run(test.scenario, func(t *testing.T) {
			t.Parallel()
			patchPath := test.patchPath
			// The file at that path will be overwritten if this function succeeds. That would make future runs of the tests fail.
			// To prevent that, for tests that won't fail, the destination file is copied to another destination file, which is used in the test
			if !test.errorExpected {
				patchPath = fmt.Sprintf("%v-test%v", patchPath, random.Intn(10000))
				data, err := ioutil.ReadFile(test.patchPath)
				if err != nil {
					t.Fatalf("Cannot read destination file in this test, despite no error being expected")
				}
				err = ioutil.WriteFile(patchPath, data, 0644)
				if err != nil {
					t.Fatalf("Cannot copy the destination file in this test")
				}
				ioutil.ReadFile(test.patchPath)
				defer os.Remove(patchPath)
			}
			err := PatchFromFiles(test.originalPath, patchPath)
			if err == nil && test.errorExpected {
				t.Error("expected error, got none")
			}
			if err != nil && !test.errorExpected {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
