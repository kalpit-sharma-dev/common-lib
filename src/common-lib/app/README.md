<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# app

App is a package to provide platform agnostic way to get version of binaries by embedding a version information (Version-Info file) along with binary details.

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/app"
```

**Creation of versioninfo.json**
OriginalFilename is output binary name:"OriginalFilename": "platform-asset-plugin.exe"
\# in the ProductVersion will be replaced by Jenkin Job: "ProductVersion": "1.0.#"

```json
{
  "FixedFileInfo": {
    "FileVersion": {
      "Major": 1,
      "Minor": 0,
      "Patch": 0,
      "Build": 0
    },
    "ProductVersion": {
      "Major": 1,
      "Minor": 0,
      "Patch": 0,
      "Build": 0
    },
    "FileType": "01"
  },
  "StringFileInfo": {
    "FileDescription": "platform-asset-plugin",
    "LegalCopyright": "@IT Management Platform",
    "OriginalFilename": "platform-asset-plugin.exe",
    "ProductVersion": "1.0.#",
    "ProductName": "ITSPlatform"
  },
  "VarFileInfo": {
    "Translation": {
      "LangID": "0409"
    }
  }
}
```

**Changes in existing code-base**
Update the main() function implementation to utilize common code implementation to embedded version information
app is a package having common implementation to include version information in the binary

```go
func main() {
    err := app.Create(execute)
    if err != nil {
        c := logger.Config{
          MaxSize:     100,
          MaxBackups:  5,
          FileName:    `log.log`,
          LogLevel:    logger.TRACE,
          ServiceName: "Plugin1",
        }

        log, _ := logger.Create(c)
        log.Error("T1", "app.create.failed", "Unable to process Request %+v", err)
    }
}

var execute = func(args []string) error {
    //Your code goes here
}
```

**Changes in make-file**

Add below mentioned variable and recipe in make-file. Change the variable/recipe according to correct path. Also call the respective recipe before the build recipe.

```
BUILDCOMMITSHA=`git rev-parse HEAD`
FLAG_PATH=github.com/ContinuumLLC/platform-plugin-asset/src/vendor/gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6
LDFLAGBUILDVERSION=-X $(FLAG_PATH)/app.BuildCommitSHA=$(BUILDCOMMITSHA)

set-version-info-windows:
    cp $(BUILDPATH)/src/main/versioninfo.json $(GOPATH)/src/$(FLAG_PATH)/app/generate/versioninfo.json
    go generate $(FLAG_PATH)/app/generate
    rm $(GOPATH)/src/$(FLAG_PATH)/app/generate/versioninfo.json

set-version-info-linux-darwin:
    cp $(BUILDPATH)/src/main/versioninfo.json $(GOPATH)/src/$(FLAG_PATH)/app/generate/versioninfo.json
    sed -i 's/platform-installation-manager.exe/platform-installation-manager/g' $(GOPATH)/src/$(FLAG_PATH)/app/generate/versioninfo.json
    go generate $(FLAG_PATH)/app/generate
    rm $(GOPATH)/src/$(FLAG_PATH)/app/generate/versioninfo.json
```

In case you're not using vendored dependencies - you should embed version data using go build like this:

```
BUILDCOMMITSHA=`git rev-parse HEAD`
FLAG_PATH=gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6

LDFLAGBUILDVERSION=-X '$(FLAG_PATH)/app.VersionInfo=$(shell tr -d "\n\r '" < versioninfo.json)'
LDFLAGBUILDCOMMIT=-X $(FLAG_PATH)/app.BuildCommitSHA=$(BUILDCOMMITSHA)
LDFLAGCOMPILEDON=-X '$(FLAG_PATH)/app.CompiledOn=$(shell date -u "+%a, %d %b %Y %H:%M:%S %z")'


ifndef LDFLAGS
LDFLAGS=${LDFLAGBUILDVERSION} ${LDFLAGBUILDCOMMIT} ${LDFLAGCOMPILEDON}
endif

go build -ldflags="$(LDFLAGS)" .
```

### Documents

[Binary Version Embedding](https://confluence.kksharmadevdev.com/display/PLATFORMTechnical/Continuum+2.0+-+Binary+Version+Embedding)

### Contribution

Any changes in this package should be communicated to Juno Team.
