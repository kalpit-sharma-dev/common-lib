<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Encoding

Wrapper around go's built-in AES-GCM encryption and decryption.

**Import Statement**

```go
import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/encoding/aes"
```
** Example

```go
package main

import (
	"encoding/base64"
	"fmt"
)

func main() {
	inputStr := "foo"
	key, err := base64.StdEncoding.DecodeString("GN+jn59704By10RN1lsYCA==")
	if err != nil {
		fmt.Println("Invalid key!", err)
		return
	}

	fmt.Println("Encrypting", inputStr)

	encrypted, err := Encrypt([]byte(inputStr), key)
	if err != nil {
		fmt.Println("Error encrypting!", err)
		return
	}

	fmt.Println("Decrypting", encrypted)

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		fmt.Println("Error decrypting!", err)
		return
	}

	fmt.Println("Successfully decrypted to", decrypted)
}
```