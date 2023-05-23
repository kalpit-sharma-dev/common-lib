package main

import (
	"encoding/base64"
	"fmt"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/encoding/aes"
)

func main() {
	inputStr := "foo"
	key, err := base64.StdEncoding.DecodeString("GN+jn59704By10RN1lsYCA==")
	if err != nil {
		fmt.Println("Invalid key!", err)
		return
	}

	fmt.Println("Encrypting", inputStr)

	encrypted, err := aes.Encrypt([]byte(inputStr), key)
	if err != nil {
		fmt.Println("Error encrypting!", err)
		return
	}

	fmt.Println("Decrypting", encrypted)

	decrypted, err := aes.Decrypt(encrypted, key)
	if err != nil {
		fmt.Println("Error decrypting!", err)
		return
	}

	fmt.Println("Successfully decrypted to", decrypted)
}
