<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Checksum

This is a Standard checksum implementation used by all the Go projects in the google.

## Import Statement

```go
import	(
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/checksum"
)
```

## Interface

```go

//Service interface for methods for checksums
Calculate(reader io.Reader) (string, error)
Validate(reader io.Reader, checksum string) (bool, error)
```

## Types

```go
//NONE is a type used for no check sum calculation and validation
checksum.NONE

//MD5 is a type used for md5 check sum calculation and validation
checksum.MD5

//SHA1 is a type used for SHA1 check sum calculation and validation
checksum.SHA1

//SHA256 is a type used for SHA256 check sum calculation and validation
checksum.SHA256
```

## Utility Functions

```go
// GetType - helper to get type by name
checksum.GetType(data string) Type

//GetService is the method to get the checksum type sum method
checksum.GetService(cType Type) (Service, error)
```

## Errors

```go
//ErrChecksumInvalid is returned if validation failed du to invalid value
checksum.ErrChecksumInvalid

//ErrUnsupportedType is returned if invalid type is supplied
checksum.ErrUnsupportedType
```

### Contribution

Any changes in this package should be communicated to Juno Team.
