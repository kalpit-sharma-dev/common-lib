package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		name            string
		initialContents string
		newContents     string
		partialUpdates  bool
		want            []UpdatedConfig
		wantErr         bool
		wantedContents  string // Should be written to the file
	}{
		{
			name:            "",
			initialContents: "{\"a\": {\"b\": 3}, \"x\": 3, \"arr\": [{\"foo\": \"bar\"}]}",
			newContents:     "{\"a\": {\"b\": 4, \"c\": 2}, \"d\": 1, \"arr\": [{\"baz\": \"quux\"}]}",
			partialUpdates:  true,
			want: []UpdatedConfig{
				{
					Key:      "a",
					Existing: nil, // Note that this is incorrect (the file in this example did have contents here), but it's the behavior of this module
					Updated: []UpdatedConfig{
						{
							Key:      "b",
							Existing: float64(3),
							Updated:  float64(4),
						},
					},
				},
				{
					Key: "arr",
					// Note that these objects were not merged - it's not like {"foo": "bar", "baz": "quux"}, it's just {"baz": "quux"}
					Existing: []interface{}{map[string]interface{}{"foo": "bar"}},
					Updated:  []interface{}{map[string]interface{}{"baz": "quux"}},
				},
				// Note that you don't receive UpdatedConfigs about added keys (like d), nor do you receive UpdatedConfigs about missing keys (like x)
			},
			wantErr:        false,
			wantedContents: "{\n\t\"a\": {\n\t\t\"b\": 4,\n\t\t\"c\": 2\n\t},\n\t\"arr\": [\n\t\t{\n\t\t\t\"baz\": \"quux\"\n\t\t}\n\t],\n\t\"d\": 1,\n\t\"x\": 3\n}\n",
		},
	}

	for index, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a file with the initial contents so that Update can read it.
			// This targets go 1.15 and testing/fstest is for go 1.16+.
			// Using testing/fstest and upgrading go.mod to 1.16 may cause an issue.
			// According to https://golang.org/ref/mod#go-mod-file-go:
			// "In lower versions, all also includes tests of packages imported by packages in the main module, tests of those packages, and so on."
			fileName := fmt.Sprint("config_test_%i.json", index)
			err := ioutil.WriteFile(fileName, []byte(tt.initialContents), 0644)
			if err != nil {
				t.Fatalf("Unable to create the test file! Err: %v", err)
			}
			defer os.Remove(fileName)

			updates, err := GetConfigurationService().Update(Configuration{
				FilePath:      fileName,
				Content:       tt.newContents,
				PartialUpdate: tt.partialUpdates,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Unexpected error %v", err)
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(updates, tt.want) {
					t.Errorf("Did not return the expected updates!\nWant: %#v\nGot:  %#v", tt.want, updates)
				}
			}

			bytes, err := ioutil.ReadFile(fileName)
			if err != nil {
				t.Errorf("Unable to read the config file that was written %v", err)
			} else if tt.wantedContents != string(bytes) {
				t.Errorf("Incorrect output in files\nWant: %#v\nGot:  %#v", tt.wantedContents, string(bytes))
			}
		})
	}
}
