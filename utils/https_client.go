package utils

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net/http"
	"os"
)

func GetRequest(url string, certConfig *CertConfig) {
	pool := x509.NewCertPool()
	if caCert, err := os.ReadFile(certConfig.CaCert); err != nil {
		log.Fatalf("read ca cert file %q failed, error %v", certConfig.CaCert, err)
	} else {
		pool.AppendCertsFromPEM(caCert)
	}
	certificates := make([]tls.Certificate, 1, 1)
	if serverCrt, err := tls.LoadX509KeyPair(certConfig.ServerCert, certConfig.ServerKey); err != nil {
		log.Fatalf("load server cert %q, key %q failed, error %v", certConfig.ServerCert, certConfig.ServerKey, err)
	} else {
		certificates[0] = serverCrt
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			Certificates: certificates,
			//            InsecureSkipVerify: false,
		},
	}
	client := &http.Client{Transport: transport}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("GET url %q failed, error %v", url, err)
	}
	defer resp.Body.Close()
	if body, err := io.ReadAll(resp.Body); err != nil {
		log.Fatalf("read request failed error %v", err)
	} else {
		log.Printf("response is: %q", string(body))
	}
}
