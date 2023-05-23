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

	eData, err = cMgr.EncryptWithCacheKey(testData)
	if err != nil {
		fmt.Println("Failed to encrypt data, Err: ", err)
		return
	}
	fmt.Println("\nUser's input data: ", testData)
	fmt.Println("\nUser's encrypted data: ", eData)

	//Decrypt data using private key
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
