<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Crypto Manager

This is a common implementation for encryption and decryption management which can be used by all the Go projects in the Continuum.
It supports generation of keys and encryption and decryption of data using these keys

### Third-Party Libraries

No third party library used

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cryptomgr"
```

**Cryptomgr Instance**

```go
//GetRSA returns instance of RSA Crypto Manager
cMgr := cryptomgr.GetRSA(&cryptomgr.Config{Password: []byte("testpassword")})
```

**Encryption/Decrypting Data**

```go
//Encrypt returns encrypted data using plubic key
encryptedData, err := cMgr.Encrypt("user data", "public key")

//Decrypt returns decrypted data using private key
decryptedData, err := cMgr.Decrypt("encrypted data", "private key")

//EncryptWithCacheKey returns encrypted data using cache public key
encryptedData, err := cMgr.EncryptWithCacheKey("user data") 

//DecryptWithCacheKey returns decrypted data using cache private key
decryptedData, err := cMgr.DecryptWithCacheKey("user data")
```

## Example

```go
package main

import (
	"fmt"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cryptomgr"
)

func main() {
	cMgr := cryptomgr.GetRSA(&cryptomgr.Config{Password: []byte("testpassword")})
	if cMgr == nil {
		fmt.Println("Failed to get Crypto Manager")
		return
	}

	//Generate keys if not availble
	err := cMgr.GenerateKeys()
	if err != nil {
		fmt.Println("Failed to generate keys, Err: ", err)
		return
	}

	//Get private key as string -> save this securely
	privateKeyString, err := cMgr.EncodePrivateKey()
	if err != nil {
		fmt.Println("cMgr.EncodePrivateKey(), Failed to encode private key, Err: ", err)
		return
	}

	//Get public key as string -> distribute this to others
	publicKeyString, err := cMgr.EncodePublicKey()
	if err != nil {
		fmt.Println("cMgr.EncodePublicKey(), Failed to encode public key, Err: ", err)
		return
	}

	//Load the key if you had already generated and was saved somewhere
	err = cMgr.LoadPrivateKey(privateKeyString)
	if err != nil {
		fmt.Println("Failed to load private key, Err: ", err)
		return
	}

	//Encrypt data using public key
	testData := "This is test data to check enc and dec"
	eData, err := cMgr.Encrypt(testData, publicKeyString)
	if err != nil {
		fmt.Println("Failed to encrypt data, Err: ", err)
		return
	}
	fmt.Println("\nUser's input data: ", testData)
	fmt.Println("\nUser's encrypted data: ", eData)

	//Decrypt data using private key
	userData, err := cMgr.Decrypt(eData, privateKeyString)
	if err != nil {
		fmt.Println("Failed to decrypt data, Err: ", err)
		return
	}
	fmt.Println("\nUser decrypted data: ", userData)

	// EncryptWithCacheKey Encrypt data using cache public key
	eData, err = cMgr.EncryptWithCacheKey(testData)
	if err != nil {
		fmt.Println("Failed to encrypt data, Err: ", err)
		return
	}
	fmt.Println("\nUser's input data: ", testData)
	fmt.Println("\nUser's encrypted data: ", eData)

	//DecryptWithCacheKey Decrypt data using cache private key
	userData, err = cMgr.DecryptWithCacheKey(eData)
	if err != nil {
		fmt.Println("Failed to decrypt data, Err: ", err)
		return
	}
	fmt.Println("\nUser decrypted data: ", userData)
	//Sign the data has with private key
	signature, err := cMgr.SignData(userData, privateKeyString)
	if err != nil {
		fmt.Println("Failed to sign data, Err: ", err)
		return
	}
	fmt.Println("\nDecrypted data signed with private key, Signature: ", signature)

	//Verify the decrypted data with signature and public key
	err = cMgr.VerifyDataWithPublicKeyString(signature, userData, publicKeyString)
	if err != nil {
		fmt.Println("Integrity of decrypted data failed, Err: ", err)
		return
	}
	fmt.Println("\nSuccessfully verified decrypted data")
}

"""
Sample Output

juno@juno:~/gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6$ go run cryptomgr/example/example.go 

User's input data:  This is test data to check enc and dec

User's encrypted data:  jt+NorNnWNt6Q6SyRiIwz63kiq4ttXTOKY2gPmgq3DV0Xm8LvURzdox98LKBuZlBPBVb6qqrrZD7clZP4MrmpWwi3KKfylIyicRxvshunvgwSI6iEK2mppvY1eZRWSkBhbbkP56XVmLyBXd8rRiHXCfHi/FBS0n1yLaPzVKD2eQU+A3Dh4Fmm8fLdzVq2z6Z4JCnT1bFIebN0FX0jUzLKciQhW2LDb0WUEmV1q98/9AR1YHbAXXfKa/+/vipFUQa6Rw6lDYMZtk3LoLtIA7NIeFrBWxIpakJTJXWB3DqeEbebcldb1eVO5XyihlJZ4L1AbdDt8dZIDMkreGPAALMAw==

User decrypted data:  This is test data to check enc and dec

User's input data:  This is test data to check enc and dec

User's encrypted data:  f5njcpVCzoJEF3Y0n7e79jhK6peOd4i7BVQnGjno5YmSs4rwk9B/WJ4AwJTvLOP/hDDD7zksw8qqvj40Cl48NQ/XtR1OPClLMY+fvGfARz7Wl7qdFAjYtilDPZMP9C2TcmVQ6Gc4nlIBkJxP14lwq08ehzHAXT5SaxP6G+T9w/KPPOSpZ+5maimZc7IBdLnJygRj/s54mFl9XjoUwEFEmBC+bHFJaVnJdMGdcxLZD2kX+a9KFMomtswSZfXvSwJScP0COASA81cQGsZbzXnzVYjZO7QGKYoBsfEE9wpA/NACKJz8W6HsUY16cqokKz+TJw5w/pSEwJynwlKhtJOOEA==

User decrypted data:  This is test data to check enc and dec

Decrypted data signed with private key, Signature:  LOnvWrxZPyKZuGEo2vjW33ChL8w3rOzki/ZbbaH1gO2UMvR46R7fumhHCtgbaDBqycV1IsnuvUqH9BxQZCKskGGj0xcKr2PhaoxQ2LEbG5ruj7G/chGXKTFcDuETUY5qVo9gDNtwIg/bp+HsekQIKmiXpkybuFre0C04VeAbgbbDPWPGRs4dwwr+PoLTSvEgiw5aCrNfQs8nOnBbgAZjJfIqsn2QG0bL5PvlNWyz1fZ5AQon80bDLu681eykInY9txMQmeCpPfk4ExE5qEb4Mpt1dnWabVLF/k6KdglnXHVvaKTaVoJ11dGu9AFhd1x5WZYKJLoYxYrvVGZkVV7+sA==

Successfully verified decrypted data
"""
```

### Contribution

Any changes in this package should be communicated to Juno Team.
