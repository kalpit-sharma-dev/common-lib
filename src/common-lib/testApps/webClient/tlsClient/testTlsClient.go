package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"
)

// zeroSource is an io.Reader that returns an unlimited number of zero bytes.
type zeroSource struct{}

func (zeroSource) Read(b []byte) (n int, err error) {
	for i := range b {
		b[i] = 0
	}

	return len(b), nil
}

func main() {
	//var URL = "https://golang.org:443"
	// var URL = "https://internal-elliptic-test-2144715301.ap-south-1.elb.amazonaws.com/agent/version"
	//var URL = "internal-elliptic-test-2144715301.ap-south-1.elb.amazonaws.com:443"
	//var URL = "https://172.28.48.140:443"
	var URL = "https://127.0.0.1:443"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify:     true,
			SessionTicketsDisabled: false,
			ClientSessionCache:     tls.NewLRUClientSessionCache(10),
			KeyLogWriter:           os.Stdout,
		},
	}

	client := &http.Client{Transport: tr}
	for index := 0; index < 10000; index++ {
		//"https://internal-google-agent-service-elb-int-1915575479.us-east-1.elb.amazonaws.com/agent/version"
		URL = "https://integration.agent.exec.itsupport247.net/agent/version"
		r, err := client.Get(URL)
		if err != nil {
			fmt.Println(time.Now().UTC(), "Done", index, "Error", err)
			continue
		}
		fmt.Println(time.Now().UTC(), "Done", index, fmt.Sprintf("%+v", r.Cookies()), "  ", r.TLS.DidResume, "   ", r.TLS.HandshakeComplete, "   ", r.TLS.Version)
		time.Sleep(10 * time.Second)
	}

	//var URL = "golang.org:443"
	//var URL = "internal-elliptic-test-2144715301.ap-south-1.elb.amazonaws.com:443"
	// var URL = "172.28.48.140:443"

	// // Load server cert and create cert pool
	// // serverCert := getServerCertificate()
	// // serverCertPool := x509.NewCertPool()
	// // serverCertPool.AppendCertsFromPEM(serverCert)

	// cfg := &tls.Config{
	// 	InsecureSkipVerify: true,
	// 	//Rand:               zeroSource{},
	// 	//Certificates:           []tls.Certificate{serverCert},
	// 	//ClientCAs:              serverCertPool,
	// 	//SessionTicketsDisabled: false,
	// 	ClientSessionCache: tls.NewLRUClientSessionCache(1),
	// 	KeyLogWriter:       os.Stdout,
	// }

	// //cfg.BuildNameToCertificate()

	// for index := 0; index < 10; index++ {
	// 	conn, err := tls.Dial("tcp", URL, cfg)

	// 	if err != nil {
	// 		fmt.Println("Error", err)
	// 	}

	// 	err = conn.Handshake()
	// 	if err != nil {
	// 		fmt.Println("Error", err)
	// 	}

	// 	fmt.Println(conn.RemoteAddr(), "  ", conn.ConnectionState().DidResume, "   ", conn.ConnectionState().HandshakeComplete, "   ", conn.ConnectionState().Version)
	// }
}

// func getServerCertificate() []byte {
// 	return []byte(cCertString)
// }

// const cCertString = `-----BEGIN CERTIFICATE-----
// MIIEaDCCA1CgAwIBAgIQDxaYW57boEoT7a+v+Q5zSjANBgkqhkiG9w0BAQsFADBG
// MQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRUwEwYDVQQLEwxTZXJ2ZXIg
// Q0EgMUIxDzANBgNVBAMTBkFtYXpvbjAeFw0xNzAzMTQwMDAwMDBaFw0xODA0MTQx
// MjAwMDBaMCUxIzAhBgNVBAMMGiouc2VydmljZS5pdHN1cHBvcnQyNDcubmV0MIIB
// IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAm8NVLFyGGlcHw6AP+nK3Z6a0
// +oO04SBFsGogodg+iBNbhm4tpF6BHSxgAP+Wevtrf2TD6ZfoOz5FCuozkN14kleV
// BuwNSSORjE3+egxznnek5SdMf8UvhvatGshLfRXhG3d6KZIicIG//LVTfbVL0Ax0
// Tcvftpz2vCT/zgu1TD+uZLEM/0perAon6xpsD2bWsQnWu/E5xczl7Fel87GUtDM5
// 4Fkh9UIJRPm1ElwvtI9TH/Mp0IHaigvcbH60mCzsJIJymZQTr0dFCbcAmEa+/vz4
// oEiZPq+HmlRUjpAUJbxq98tV0jS8jAUC7CJv35Ana3+D3eNwRPEdWwDiqfSxrQID
// AQABo4IBcTCCAW0wHwYDVR0jBBgwFoAUWaRmBlKge5WSPKOUByeWdFv5PdAwHQYD
// VR0OBBYEFLwn1GO1060BtQpXNhIVT/izw4vcMCUGA1UdEQQeMByCGiouc2Vydmlj
// ZS5pdHN1cHBvcnQyNDcubmV0MA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUEFjAUBggr
// BgEFBQcDAQYIKwYBBQUHAwIwOwYDVR0fBDQwMjAwoC6gLIYqaHR0cDovL2NybC5z
// Y2ExYi5hbWF6b250cnVzdC5jb20vc2NhMWIuY3JsMBMGA1UdIAQMMAowCAYGZ4EM
// AQIBMHUGCCsGAQUFBwEBBGkwZzAtBggrBgEFBQcwAYYhaHR0cDovL29jc3Auc2Nh
// MWIuYW1hem9udHJ1c3QuY29tMDYGCCsGAQUFBzAChipodHRwOi8vY3J0LnNjYTFi
// LmFtYXpvbnRydXN0LmNvbS9zY2ExYi5jcnQwDAYDVR0TAQH/BAIwADANBgkqhkiG
// 9w0BAQsFAAOCAQEAtNvzhoSvC+u/RLWXOtA3ciPZtr3Kw3p+L4nYwY4QQJmSaUwi
// tOVdT11h0I3akoFR9i1+BVbJvYev1ji4G2S0gGn8PyCcRvqrs0TgT3B9FFQw2PLO
// 1vLnkrMCBQEx4Gi5WFHNumy+OGgG4bsv/0+cBnrpJ6BqiRNifuy4INjrpnIOpXW+
// CqqIdPmA+lVvfHxfIOfV+Sr3+OLW08kdigRs7NFVbZWIWF47vBM5opYUlB0eB3eK
// nKEA1/hSUiD/oT+aCNMWMePmB9dxX3YIVFLz9q2OZsAS92UsDlcyOUcM6U3F2qIX
// m83xv6owjxdxk1gH/la5cDNeigMYBfGnGYBl+g==
// -----END CERTIFICATE-----`
