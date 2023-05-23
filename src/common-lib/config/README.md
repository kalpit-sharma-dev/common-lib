<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# config

JSON configuration reading, writing and diffing

**Import Statement**

```go
import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/config"
```

## Example

```go
updates, err := config.GetConfigurationService().Update(config.Configuration{
  // This file should have the old config, and will be written to with the new config
  FilePath: "/tmp/config.json", // Example contents: {"a": {"b": 3}, "x": 3, "arr": [{"foo": "bar"}]}
  // After calling this Update, the file should be updated to have these contents
	Content: "{\"a\": {\"b\": 4, \"c\": 2}, \"d\": 1, \"arr\": [{\"baz\": \"quux\"}]}",
  // Setting this to false will prevent you from getting any errors or any info about the changes (the updates variable will be empty), and won't do any merging.
  // Setting this to false is like asking for a straight "write these contents to this file path"
  PartialUpdate: true,
})

if err != nil {
  fmt.Println("Could throw due to a file reading/writing error, or because a key which was a JSON object in the contents of the original file is now not a JSON object in contents")
}

// Now the file has these contents (but prettier, with newlines and tabs): {"a": {"b": 4, "c": 2}, "arr": [{"baz": "quux"}], "d": 1, "x": 3}

reflect.DeepEqual(updates, []UpdatedConfig{
  {
    Key: "a",
    Existing: nil, // Note that this is incorrect (the file in this example did have contents here), but it's the behavior of this module
    Updated: []UpdatedConfig{
      {
        Key: "b",
        Existing: float64(3),
        Updated: float64(4),
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
})
```

If you would like a more consistent behavior, you may want something like [github.com/wI2L/jsondiff](https://pkg.go.dev/github.com/wI2L/jsondiff) which will produce proper [JSON Patch](http://jsonpatch.com/) output.
