package cryptomgr

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

//rsaManager Crypto Manager
type rsaManager struct {
	privateKey *rsa.PrivateKey
	cfg        *Config
}

const (
	rsaPrivate = "RSA PRIVATE KEY"
	rsaPublic  = "RSA PUBLIC KEY"
)

//GenerateKeys generates new key pairs
func (cMgr *rsaManager) GenerateKeys() error {
	var err error
	cMgr.privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	return err
}

//EncodePrivateKey Get private key as string (return different cipher everytime for same key)
func (cMgr *rsaManager) EncodePrivateKey() (string, error) {
	privateKey := ""
	if nil == cMgr.privateKey {
		return privateKey, errors.New("Private key is nil")
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(cMgr.privateKey)
	if err != nil {
		return privateKey, fmt.Errorf("Marshaling private key to bytes, %+v", err)
	}

	block, err := x509.EncryptPEMBlock(rand.Reader, rsaPrivate, keyBytes, cMgr.cfg.Password, x509.PEMCipherAES256)
	if err != nil {
		return privateKey, fmt.Errorf("Encoding private key with AES256, %+v", err)
	}

	privateKey = string(pem.EncodeToMemory(block))

	return privateKey, nil
}

//LoadPrivateKey Read private key from string
func (cMgr *rsaManager) LoadPrivateKey(strPrivateKey string) error {
	privateKey, err := cMgr.decodePrivateKey(strPrivateKey)
	if err != nil {
		return fmt.Errorf("Failed to load private key, %v", err)
	}
	cMgr.privateKey = privateKey
	return nil
}

//DecodePrivateKey Read private key from string
func (cMgr *rsaManager) decodePrivateKey(strPrivateKey string) (*rsa.PrivateKey, error) {
	block, rest := pem.Decode([]byte(strPrivateKey))
	if block == nil || block.Type != rsaPrivate || len(rest) != 0 {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	//DecryptPEMBlock TODO: check doc, this is not reliable
	pemBytes, err := x509.DecryptPEMBlock(block, cMgr.cfg.Password)
	if err != nil {
		return nil, errors.New("failed to decrypt PEM block containing private key")
	}

	privateKeyImported, err := x509.ParsePKCS8PrivateKey(pemBytes)
	if err != nil {
		return nil, err
	}

	return privateKeyImported.(*rsa.PrivateKey), nil
}

//EncodePublicKey Get public key as string
func (cMgr *rsaManager) EncodePublicKey() (string, error) {
	publicKey := ""
	if nil == cMgr.privateKey {
		return publicKey, errors.New("Private key is nil")
	}

	block := &pem.Block{
		Type:  rsaPublic,
		Bytes: x509.MarshalPKCS1PublicKey(&cMgr.privateKey.PublicKey),
	}
	return string(pem.EncodeToMemory(block)), nil
}

//DecodePublicKey Read public key from string
func (cMgr *rsaManager) decodePublicKey(strPublicKey string) (*rsa.PublicKey, error) {
	block, rest := pem.Decode([]byte(strPublicKey))
	if block == nil || block.Type != rsaPublic && len(rest) == 0 {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	publicKeyImported, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to import public key")
	}
	return publicKeyImported, nil
}

//PrivateKeyInstance Returns instance of private key
func (cMgr *rsaManager) PrivateKeyInstance(strPrivateKey string) (interface{}, error) {
	return cMgr.decodePrivateKey(strPrivateKey)
}

//PublicKeyInstance Returns instance of public key
func (cMgr *rsaManager) PublicKeyInstance(strPublicKey string) (interface{}, error) {
	return cMgr.decodePublicKey(strPublicKey)
}

//Encrypt encrypt data
func (cMgr *rsaManager) Encrypt(data, strPublicKey string) (string, error) {
	publicKey, err := cMgr.decodePublicKey(strPublicKey)
	if err != nil {
		return "", fmt.Errorf("Invalid public key, %v", err)
	}
	msg := []byte(data)
	hash := sha256.New()
	step := publicKey.Size() - 2*hash.Size() - 2
	isLarge := len(msg) > step

	ciphertext, err := handler(msg, isLarge, func(message []byte) ([]byte, error) {
		return rsa.EncryptOAEP(hash, rand.Reader, publicKey, message, nil)
	}, step)

	return base64.StdEncoding.EncodeToString(ciphertext), err
}

//Decrypt decrypt data
func (cMgr *rsaManager) Decrypt(cipherData, strPrivateKey string) (string, error) {
	dataBytes, err := base64.StdEncoding.DecodeString(cipherData)
	if err != nil {
		return "", err
	}

	privateKey, err := cMgr.decodePrivateKey(strPrivateKey)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	k := privateKey.Size()
	step := privateKey.PublicKey.Size()
	isLarge := len(dataBytes) > k || k < hash.Size()*2+2

	decyptedData, err := handler(dataBytes, isLarge, func(message []byte) ([]byte, error) {
		return rsa.DecryptOAEP(hash, rand.Reader, privateKey, message, nil)
	}, step)

	return string(decyptedData), err
}

/*Function to handle encryption and decryption for large messages. If the message is not large it processes the entire message in one go. If the message is large then based on step it processes the message and concatenates the encoded data as result.

toExecute -> function as parameter is used to execute encrypt/decrypt functionality based on callers implementation on provided data.*/
func handler(msg []byte, isLarge bool, toExecute func(msg []byte) ([]byte, error), step int) ([]byte, error) {
	if !isLarge {
		// message is small and we do not want to partition it.
		return toExecute(msg)
	}
	// message is large, we want to partition it further and process individual partition.
	msgLen := len(msg)
	var data []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}
		dataBytes, err := toExecute(msg[start:finish])
		if err != nil {
			return nil, err
		}
		data = append(data, dataBytes...)
	}
	return data, nil
}

//EncryptWithCacheKey encrypt data
func (cMgr *rsaManager) EncryptWithCacheKey(data string) (string, error) {
	if nil == cMgr.privateKey {
		return "", errors.New("Private key is nil")
	}
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &cMgr.privateKey.PublicKey, []byte(data), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), err
}

//DecryptWithCacheKey decrypt data
func (cMgr *rsaManager) DecryptWithCacheKey(cipherData string) (string, error) {
	dataBytes, err := base64.StdEncoding.DecodeString(cipherData)
	if err != nil {
		return "", err
	}
	decyptedData, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, cMgr.privateKey, dataBytes, nil)
	return string(decyptedData), err
}

//SignData sign the data hash using private key
func (cMgr *rsaManager) SignData(data string, strPrivateKey string) (string, error) {
	privateKey, err := cMgr.decodePrivateKey(strPrivateKey)
	if err != nil {
		return "", err
	}
	dataHash := sha256.Sum256([]byte(data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, dataHash[:])
	return base64.StdEncoding.EncodeToString(signature), err
}

//VerifyDataWithPublicKeyInstance verify the data hash using public key instance
func (cMgr *rsaManager) VerifyDataWithPublicKeyInstance(signature, data string, publicKey interface{}) error {
	if publicKey == nil || publicKey.(*rsa.PublicKey) == nil {
		return fmt.Errorf("Invalid public key")
	}

	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("Failed to decode signature, %v", err)
	}
	dataHash := sha256.Sum256([]byte(data))

	return rsa.VerifyPKCS1v15(publicKey.(*rsa.PublicKey), crypto.SHA256, dataHash[:], sig)
}

//VerifyDataWithPublicKeyString verify the data hash using public key instance
func (cMgr *rsaManager) VerifyDataWithPublicKeyString(signature, data, strPublicKey string) error {
	publicKey, err := cMgr.decodePublicKey(strPublicKey)
	if err != nil {
		return fmt.Errorf("Invalid public key, %v", err)
	}
	return cMgr.VerifyDataWithPublicKeyInstance(signature, data, publicKey)
}
