package aes

import (
	"encoding/base64"
	"testing"
)

func TestEncrypt(t *testing.T) {
	inputStr := "asd"
	key, err := base64.StdEncoding.DecodeString("GN+jn59704By10RN1lsYCA==")
	if err != nil {
		t.Fatalf("You used in invalid key")
	}
	encrypted, err := Encrypt([]byte(inputStr), key)
	if err != nil {
		t.Fatalf("Error encrypting! %v", err)
	}
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Error decrypting! %v", err)
	}
	if string(decrypted) != "asd" {
		t.Errorf("Decryption result was \"%v\" want=\"%v\"", inputStr, string(decrypted))
	}
}
