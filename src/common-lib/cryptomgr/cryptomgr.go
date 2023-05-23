package cryptomgr

//go:generate mockgen -package mock -destination=mock/mocks.go . CryptoMgr

//CryptoMgr Encryption Crypto Manager
type CryptoMgr interface {
	//GenerateKeys Generate new key pair
	GenerateKeys() error
	//LoadPrivateKey Decodes private key from given string
	LoadPrivateKey(strPrivateKey string) error

	//PublicKeyInstance Returns instance of private key
	PrivateKeyInstance(strPrivateKey string) (interface{}, error)
	//PublicKeyInstance Returns instance of public key
	PublicKeyInstance(strPrivateKey string) (interface{}, error)

	//EncodePrivateKey Returns string format of the private
	EncodePrivateKey() (string, error)
	//EncodePublicKey Returns string format of the public
	EncodePublicKey() (string, error)

	//Encrypt Encrypts the given data with provided public key and returns the encrypted data as string
	Encrypt(data, strPublicKey string) (string, error)
	//Decrypt Decrypts the given data with provided private key and returns the decrypted data as string
	Decrypt(cipherData, strPrivateKey string) (string, error)

	//EncryptWithCacheKey Encrypts the given data with cache public key and returns the encrypted data as string
	EncryptWithCacheKey(data string) (string, error)
	//DecryptWithCacheKey Decrypts the given data with cache private key and returns the decrypted data as string
	DecryptWithCacheKey(cipherData string) (string, error)

	//SignData sign the data hash using private key
	SignData(data, strPrivateKey string) (string, error)
	//VerifyDataWithPublicKeyInstance verify the data hash using public key instance
	VerifyDataWithPublicKeyInstance(signature, data string, publicKey interface{}) error
	//VerifyDataWithPublicKeyString verify the data hash using public key string
	VerifyDataWithPublicKeyString(signature, data, strPublicKey string) error
}

//Config config used by CryptoMgr
type Config struct {
	//Password used to encrypt private key, used as label, etc
	Password []byte `json:"Password"`
}

//GetRSA Returns instance of RSA Crypto Manager
func GetRSA(cfg *Config) CryptoMgr {
	return &rsaManager{cfg: cfg}
}
