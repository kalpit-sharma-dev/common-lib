package cryptomgr

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func setupKeyPair(t *testing.T) (cMgr CryptoMgr, privateKeyStr string, publicKeyStr string) {
	cMgr = GetRSA(&Config{Password: []byte("testpassword")})
	if cMgr == nil {
		t.Errorf("setupKeyPair(), Failed to get Crypto Manager")
	}

	err := cMgr.GenerateKeys()
	if err != nil {
		t.Errorf("setupKeyPair(), Failed to generate keys, Err: %v", err)
	}

	privateKeyStr, err = cMgr.EncodePrivateKey()
	if err != nil {
		t.Errorf("setupKeyPair(), Failed to encode private key, Err: %v", err)
	}

	publicKeyStr, err = cMgr.EncodePublicKey()
	if err != nil {
		t.Errorf("setupKeyPair(), Failed to encode public key, Err: %v", err)
	}

	return cMgr, privateKeyStr, publicKeyStr
}

func Test_key_HappyPath(t *testing.T) {
	cMgr, privateKeyStr, publicKeyStr := setupKeyPair(t)

	err := cMgr.LoadPrivateKey(privateKeyStr)
	if err != nil {
		t.Errorf("Test_key_HappyPath.LoadPrivateKey(), Failed to load private key, Err: %v", err)
	}

	testData := "This is test data to check enc and dec"
	eData, err := cMgr.Encrypt(testData, publicKeyStr)
	if err != nil {
		t.Errorf("Test_key_HappyPath.Encrypt(), error = %+v", err)
	}
	got, err := cMgr.Decrypt(eData, privateKeyStr)
	if err != nil || !reflect.DeepEqual(got, testData) {
		t.Errorf("Test_key_HappyPath.Decrypt(), error = %+v, got = %+v, want = %+v", err, got, testData)
	}
	eDataNoKey, err := cMgr.EncryptWithCacheKey(testData)
	if err != nil {
		t.Errorf("Test_key_HappyPath.Encrypt(), error = %+v", err)
	}
	gotNoKey, err := cMgr.DecryptWithCacheKey(eDataNoKey)
	if err != nil || !reflect.DeepEqual(gotNoKey, testData) {
		t.Errorf("Test_key_HappyPath.Decrypt(), error = %+v, got = %+v, want = %+v", err, gotNoKey, testData)
	}
}

func Test_key_UnhappyPath(t *testing.T) {
	cMgr := GetRSA(&Config{Password: []byte("testpassword")})
	if cMgr == nil {
		t.Errorf("cMgr.GetRSA(), Failed to get Crypto Manager")
	}

	//Failed to get private key as string
	if _, err := cMgr.EncodePrivateKey(); err == nil {
		t.Errorf("Error expected as private key is empty, got nil")
	}

	//Failed to get public key as string
	if _, err := cMgr.EncodePublicKey(); err == nil {
		t.Errorf("Error expected as public key is empty, got nil")
	}

	//Failed to load key
	if cMgr.LoadPrivateKey("") == nil {
		t.Errorf("Error expected while loading private key, got nil")
	}

	err := cMgr.GenerateKeys()
	if err != nil {
		t.Errorf("cMgr.GenerateKeys(), Failed to generate keys, Err: %v", err)
	}

	//Failed to load key
	if cMgr.LoadPrivateKey("") == nil {
		t.Errorf("Error expected while loading private key, got nil")
	}

	//Failed decrypt private key
	if cMgr.LoadPrivateKey("-----BEGIN RSA PRIVATE KEY-----\nProc-Type: 4,ENCRYPTED\nDEK-Info: AES-256-CBC,68020bedfc83a27b641d9eb706e0eeaf\n\nmNLIxCytuh2odV1bLmypVFamooOh/aLtacTpjw8W0vlnpJLGyxQ4ukSk/4UAfkHj\ntBHT6U0DXqxchx9cDNQBfCYl7Uus/9IRP+yscYGAqbH+OiFyQ1gzvCdMW3jF7qKz\n9CL6GocQgcjhaWN31NCYvVdCRvkPxKUFioLQJcvccpY9eYFauBoakXHkCUtic07C\n8UOSdGTIKk5mf60HVC7wUf5N9SFPmqMwSHjojbpIQt2q7HathWPUWYEbATYUKAcL\n4fAEtvFPaK1nfoYfxnOL9DMF4yZtr44UQBXHZdibwK4KMdn+JvbzZDzk8Bw5gKVS\nbAaebaATX2KN++xBnrmkwMrnKY1Ivam1D1YxAFaKwFZAQzLWUBABaNGuKSKCCgpF\nNwBvCzp1724U/biFkj1+qjSWanFduRdtiolXg9IoNp1MZAm4dBlsB6qWeJlZtObZ\nB/xgZbtBPYDaLFOARcqfDBFGOWPu6K4nSvJ6c7vSZGDDsz7IUObrpNVFFHdrBern\n/0kvc1pI2cWCfDhIlMI6gJoOEQWZxl+XLQVaUIYTVrJFD0yUfcVHNskg7dDmXDoY\niA5uLUboLPqyoToRm/slQ/5GooMhdUb6q6pXk9pBl0FmPFLCiSVX907ncF201VuR\n+iYe1O8XuI15bde6PZ0WBKyigCSIzEIe1aY4vHcL8tCotyQGMtAij1yv0+p9i3wQ\nt2vXD8I+ftZr4RXB6B1Fci6pcQhVH/Kk1+bo5z3BGR9hdA30TClIZMVcjsGPPcVo\nYzyQXfL5Q+f+wmo05PArVuqhxBh0obP9OgGv/Bq4DBkUYhceeyxphO4W1fhg4Yt5\noV4Qp/9SP7jiR/CRWfV7qrZ8muiXm1QKtaYv5t+zvdTBCEN4NisYyOy0uHjJw02Y\n1QXxLFXpMDfnK1IacliaLqOYvRe/iZ1omuZeYcngDb5/LN4KVNA2K68+KtOG+AwD\nHdeEYv/jdN0q4Mf612DaHKOptqvMqHXngEPFs7KWr2y4XlZFGbCrHCB/TA8bqhVn\nipAYHnnOdiRzQcpDgb19QDONaeflfR9GtWCQb2pUlG5w2JdR6E99HE5H2rHx1AMv\nSL/YVCkHXNko0+H/iRhpMLCDgNkr14kDrXxGPUQ5C6c0SwvIJd5mLWYvqq9oNHnM\n4zFWSlXloax2WZUNjZ6CRCkM/1lDCzsKPsIZD2ulQdCiX97K6lvUqcVM4fndJCxR\neWTjrSYW9WOxxFonFQrSmS2D/amCKkNBlzRrfHlsr1DpYiFFiqPNSgP5g17g4rtZ\nM7eY3ktExB7//+jukC59J9YgwFc+ixGb4IPJw7i0Uk9FbUBBtwOtEi+TAjX9r+Eu\nhKvgQEN5Buo4PWKNKsCKplHL5IGg4jAAxhrU9bv6Xxjmyk5YJ+VxYWjUpNQeiw3R\n9FIWC3c/G0UY0743wE4knMkxGND9GMPRclelAgtr4P/ddCGYgZ98JMc45SQOdOqr\n8+v6ZdfKer/NaDwaJk5MXo8qwp0vkmlvtbLoCz2n9bku4Wu3re4x7BVYhLqqUNdT\n1MYubZf526IJ/LjW5XlzG+9bddWozJMox6ojwtWaQj9nL3hgoZkemzRAs+1b31gI\nSQ2ckiJcZSb7ecgDr3jiPO6dGY0MAYkAGZWfLHvTaNo=\n-----END RSA PRIVATE KEY-----\n") == nil {
		t.Errorf("Error expected while decrypting private, got nil")
	}

	//Failed encrypt data
	if _, err := cMgr.Encrypt("testdata", "dummy public key"); err == nil {
		t.Errorf("Error expected while encrypting data, got nil")
	}

	//Failed decrypt data
	if _, err := cMgr.Decrypt("testdata", "dummy private key"); err == nil {
		t.Errorf("Error expected while encrypting data, got nil")
	}

	//Failed decrypt data - invalid cipher data
	if _, err := cMgr.Decrypt("asd", "dummy private key"); err == nil {
		t.Errorf("Error expected while encrypting data, got nil")
	}

	//Failed encrypt data
	if _, err := cMgr.EncryptWithCacheKey("asd"); err != nil {
		t.Errorf("Failed to encrypting data, got nil")
	}

	//Failed decrypt data - invalid cipher data
	if _, err := cMgr.DecryptWithCacheKey("asd"); err == nil {
		t.Errorf("Error expected while decrypting data, got nil")
	}
}

func Test_rsaManager_SignData(t *testing.T) {
	type args struct {
		data          string
		strPrivateKey string
	}
	tests := []struct {
		rsaMgr     *rsaManager
		name       string
		data       string
		privateKey string
		wantErr    bool
		err        error
	}{
		{
			name: "Sign data successfully",
			privateKey: func() string {
				_, pk, _ := setupKeyPair(t)
				return pk
			}(),
			data:    "Test data to be signed",
			wantErr: false,
			err:     nil,
		},
		{
			name:    "Sign data with invalid private key",
			data:    "Test data to be signed",
			wantErr: true,
			err:     fmt.Errorf("failed to decode PEM block containing private key"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (&rsaManager{cfg: &Config{Password: []byte("testpassword")}}).SignData(tt.data, tt.privateKey)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("rsaManager.SignData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
func Test_rsaManager_VerifyDataWithPublicKeyInstance(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (data, signature string, publicKey *rsa.PublicKey)
		err   error
	}{
		{
			name: "Successfully verify the signed hash - data intact",
			setup: func() (data, signature string, publicKey *rsa.PublicKey) {
				cMgr, privateKeyStr, publicKeyStr := setupKeyPair(t)
				_publicKey, err := cMgr.PublicKeyInstance(publicKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				data = "Test data to be signed"
				signature, err = cMgr.SignData(data, privateKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				return data, signature, _publicKey.(*rsa.PublicKey)
			},
			err: nil,
		},
		{
			name: "Successfully verify the signed hash - data tampered",
			setup: func() (data, signature string, publicKey *rsa.PublicKey) {
				cMgr, privateKeyStr, publicKeyStr := setupKeyPair(t)
				_publicKey, err := cMgr.PublicKeyInstance(publicKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				data = "Test data to be signed"
				signature, err = cMgr.SignData(data, privateKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				return data + "data tampered", signature, _publicKey.(*rsa.PublicKey)
			},
			err: fmt.Errorf("crypto/rsa: verification error"),
		},
		{
			name: "Successfully verify the signed hash - invalid signature",
			setup: func() (data, signature string, publicKey *rsa.PublicKey) {
				cMgr, privateKeyStr, publicKeyStr := setupKeyPair(t)
				_publicKey, err := cMgr.PublicKeyInstance(publicKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				data = "Test data to be signed"
				signature, err = cMgr.SignData(data, privateKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				return data, signature + "-invalid signature", _publicKey.(*rsa.PublicKey)
			},
			err: fmt.Errorf("Failed to decode signature, illegal base64 data at input byte 344"),
		},
		{
			name: "Public key nil",
			setup: func() (data, signature string, publicKey *rsa.PublicKey) {
				cMgr, privateKeyStr, _ := setupKeyPair(t)
				data = "Test data to be signed"
				signature, err := cMgr.SignData(data, privateKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				return data, signature, nil
			},
			err: fmt.Errorf("Invalid public key"),
		},
		{
			name: "Public key changed",
			setup: func() (data, signature string, publicKey *rsa.PublicKey) {
				cMgr, privateKeyStr, _ := setupKeyPair(t)
				data = "Test data to be signed"
				signature, err := cMgr.SignData(data, privateKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}

				newCMgr, _, newPublicKeyStr := setupKeyPair(t)
				_newPublicKey, err := newCMgr.PublicKeyInstance(newPublicKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				return data, signature, _newPublicKey.(*rsa.PublicKey)
			},
			err: fmt.Errorf("crypto/rsa: verification error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, signature, publicKey := tt.setup()
			if err := (&rsaManager{cfg: &Config{Password: []byte("testpassword")}}).VerifyDataWithPublicKeyInstance(signature, data, publicKey); !reflect.DeepEqual(err, tt.err) {
				t.Errorf("rsaManager.VerifyDataWithPublicKeyInstance() want = %v, got %v", err, tt.err)
			}
		})
	}
}

func Test_rsaManager_VerifyDataWithPublicKeyString(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (data, signature, publicKey string)
		err   error
	}{
		{
			name: "Successfully verify the signed hash - data intact",
			setup: func() (data, signature, publicKey string) {
				cMgr, privateKeyStr, publicKeyStr := setupKeyPair(t)
				data = "Test data to be signed"
				signature, err := cMgr.SignData(data, privateKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				return data, signature, publicKeyStr
			},
			err: nil,
		},
		{
			name: "Successfully verify the signed hash - data tampered",
			setup: func() (data, signature, publicKey string) {
				cMgr, privateKeyStr, publicKeyStr := setupKeyPair(t)
				data = "Test data to be signed"
				signature, err := cMgr.SignData(data, privateKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				return data + "data tampered", signature, publicKeyStr
			},
			err: fmt.Errorf("crypto/rsa: verification error"),
		},
		{
			name: "Successfully verify the signed hash - invalid signature",
			setup: func() (data, signature, publicKey string) {
				cMgr, privateKeyStr, publicKeyStr := setupKeyPair(t)
				data = "Test data to be signed"
				signature, err := cMgr.SignData(data, privateKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				return data, signature + "-invalid signature", publicKeyStr
			},
			err: fmt.Errorf("Failed to decode signature, illegal base64 data at input byte 344"),
		},
		{
			name: "Public key invalid",
			setup: func() (data, signature, publicKey string) {
				cMgr, privateKeyStr, _ := setupKeyPair(t)
				data = "Test data to be signed"
				signature, err := cMgr.SignData(data, privateKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}
				return data, signature, ""
			},
			err: fmt.Errorf("Invalid public key, failed to decode PEM block containing public key"),
		},
		{
			name: "Public key changed",
			setup: func() (data, signature, publicKey string) {
				cMgr, privateKeyStr, _ := setupKeyPair(t)
				data = "Test data to be signed"
				signature, err := cMgr.SignData(data, privateKeyStr)
				if err != nil {
					t.Errorf("Err: %v", err)
				}

				_, _, newPublicKeyStr := setupKeyPair(t)
				return data, signature, newPublicKeyStr
			},
			err: fmt.Errorf("crypto/rsa: verification error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, signature, publicKey := tt.setup()
			if err := (&rsaManager{cfg: &Config{Password: []byte("testpassword")}}).VerifyDataWithPublicKeyString(signature, data, publicKey); !reflect.DeepEqual(err, tt.err) {
				t.Errorf("rsaManager.VerifyDataWithPublicKeyInstance() error = %v, got %v", err, tt.err)
			}
		})
	}
}

func Test_rsaManager_PrivateKeyInstance(t *testing.T) {
	type args struct {
		strPrivateKey string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "Valid Private key",
			args: args{
				strPrivateKey: func() string {
					_, pk, _ := setupKeyPair(t)
					return pk
				}(),
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "Invalid Private key",
			args: args{
				strPrivateKey: "",
			},
			wantErr: true,
			err:     fmt.Errorf("failed to decode PEM block containing private key"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&rsaManager{cfg: &Config{Password: []byte("testpassword")}}).PrivateKeyInstance(tt.args.strPrivateKey)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("rsaManager.PrivateKeyInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == false && (got == nil || got.(*rsa.PrivateKey) == nil) {
				t.Errorf("rsaManager.PrivateKeyInstance() got nil Private key, want instance")
			}
		})
	}
}

func Test_rsaManager_PublicKeyInstance(t *testing.T) {
	type args struct {
		strPublicKey string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "Valid public key",
			args: args{
				strPublicKey: func() string {
					_, _, pk := setupKeyPair(t)
					return pk
				}(),
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "Invalid public key",
			args: args{
				strPublicKey: "",
			},
			wantErr: true,
			err:     fmt.Errorf("failed to decode PEM block containing public key"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&rsaManager{cfg: &Config{Password: []byte("testpassword")}}).PublicKeyInstance(tt.args.strPublicKey)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("rsaManager.PublicKeyInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == false && (got == nil || got.(*rsa.PublicKey) == nil) {
				t.Errorf("rsaManager.PublicKeyInstance() got nil public key, want instance")
			}
		})
	}
}

func Test_handler(t *testing.T) {
	smallMsg := "This is a sample of small message written."
	largeMsg := "[{\\\"address\\\":\\\"10.2.00.00\\\",\\\"username\\\":\\\"admin\\\",\\\"password\\\":\\\"admin@123\\\",\\\"uuid\\\":\\\"af801984-451d-11ec-81d3-0242ac130014\\\",\\\"vcenterIP\\\":\\\"10.2.40.81\\\"},{\\\"address\\\":\\\"10.2.00.01\\\",\\\"username\\\":\\\"admin\\\",\\\"password\\\":\\\"admin@123\\\",\\\"uuid\\\":\\\"af801984-451d-11ec-81d3-0242ac130015\\\",\\\"vcenterIP\\\":\\\"10.2.40.81\\\"}]"
	type arg struct {
		msg       []byte
		isLarge   bool
		toExecute func([]byte) ([]byte, error)
		step      int
	}
	tests := []struct {
		name    string
		arg     arg
		want    []byte
		wantErr bool
	}{
		{
			name: "Test1-SmallMessage",
			arg: arg{msg: []byte(smallMsg), isLarge: false, step: 0, toExecute: func(b []byte) ([]byte, error) {
				return []byte(b), nil
			}},
			want:    []byte(smallMsg),
			wantErr: false,
		},
		{
			name: "Test2-LargeMessageErr",
			arg: arg{msg: []byte(largeMsg), isLarge: true, step: 10, toExecute: func(b []byte) ([]byte, error) {
				return nil, errors.New("LargeEncoreErr")
			}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test3-LargeMessageSuccess",
			arg: arg{msg: []byte(largeMsg), isLarge: true, step: 10, toExecute: func(b []byte) ([]byte, error) {
				return []byte(b), nil
			}},
			want:    []byte(largeMsg),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := handler(tt.arg.msg, tt.arg.isLarge, tt.arg.toExecute, tt.arg.step)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handler() got = %v, want = %v", string(got), string(tt.want))
			}
			if gotErr == nil && tt.wantErr {
				t.Errorf("handler() gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_EncryptDecrypt(t *testing.T) {
	cMgr, privateKeyStr, publicKeyStr := setupKeyPair(t)

	err := cMgr.LoadPrivateKey(privateKeyStr)
	if err != nil {
		t.Errorf("Test_EncryptDecrypt.LoadPrivateKey(), Failed to load private key, Err: %v", err)
	}

	testData := "This is test data to check enc and dec small msg"
	eData, err := cMgr.Encrypt(testData, publicKeyStr)
	if err != nil {
		t.Errorf("Test_EncryptDecrypt.Encrypt(), error = %+v", err)
	}
	got, err := cMgr.Decrypt(eData, privateKeyStr)
	if err != nil || !reflect.DeepEqual(got, testData) {
		t.Errorf("Test_EncryptDecrypt.Decrypt(), error = %+v, got = %+v, want = %+v", err, got, testData)
	}
	longMsg := "[{\\\"address\\\":\\\"10.2.00.00\\\",\\\"username\\\":\\\"admin\\\",\\\"password\\\":\\\"admin@123\\\",\\\"uuid\\\":\\\"af801984-451d-11ec-81d3-0242ac130014\\\",\\\"vcenterIP\\\":\\\"10.2.40.81\\\"},{\\\"address\\\":\\\"10.2.00.01\\\",\\\"username\\\":\\\"admin\\\",\\\"password\\\":\\\"admin@123\\\",\\\"uuid\\\":\\\"af801984-451d-11ec-81d3-0242ac130015\\\",\\\"vcenterIP\\\":\\\"10.2.40.81\\\"}]"
	eData, err = cMgr.Encrypt(longMsg, publicKeyStr)
	if err != nil {
		t.Errorf("Test_EncryptDecrypt.Encrypt(), error = %+v", err)
	}
	got, err = cMgr.Decrypt(eData, privateKeyStr)
	if err != nil || !reflect.DeepEqual(got, longMsg) {
		t.Errorf("Test_EncryptDecrypt.Decrypt(), error = %+v, got = %+v, want = %+v", err, got, testData)
	}
}
