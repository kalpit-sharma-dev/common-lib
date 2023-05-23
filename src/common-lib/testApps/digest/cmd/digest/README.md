## Digest Utility
Digest Utility creates a digest for the specified file using SHA256 hash function
and then verifies if the specified file corresponds to the digest.

## Build
(from *digest* folder)

$ go build .

## Main Use Cases
### Creation of a digest for the file
$ digest [-c] -f file [-d digest]
If *digest* is not set, *file***.sha256** is used as a result.
### Verification of the file against the digest
$ digest -v -f file [-d digest]
If *digest* is not set, *file***.sha256** is used.
### Getting help
$ digest -h

## Flags
**-h** - print help message (optional)
**-c** - create a digest (set by default)
**-v** - verify the file against the digest
**-f** - set input file (required)
**-d** - set digest file (if not set, *file***.sha256** is used.)
**-e** - set a new digest file extension (overrides the default value **.sha256**)
Only one flag **-c** (create) or **-v** (verify) could be set.
Only one flag **-f** (file to be processed) is required.