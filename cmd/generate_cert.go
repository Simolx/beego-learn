package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func generatePrivateKey() (*rsa.PrivateKey, []byte, error) {
	priKey4096 := 4096
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, priKey4096)
	if err != nil {
		log.Fatalf("generate key failed, %v", err)
	}
	pkcsPrivateKey, err := x509.MarshalPKCS8PrivateKey(rsaPrivateKey)
	if err != nil {
		log.Fatalf("create private key failed, error %v", err)
	}
	privateBlock := pem.Block{Type: "RSA PRIVATE KEY", Bytes: pkcsPrivateKey}
	return rsaPrivateKey, pem.EncodeToMemory(&privateBlock), nil
}

func SignCACert(privateKey *rsa.PrivateKey, serviceName string) []byte {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}
	before := time.Now()
	after := before.AddDate(1, 0, 0)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Province:           []string{"BeiJing"},
			Organization:       []string{"devCompany"},
			OrganizationalUnit: []string{"devTeam"},
			CommonName:         serviceName,
		},
		NotBefore:             before,
		NotAfter:              after,
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageCodeSigning, x509.ExtKeyUsageEmailProtection, x509.ExtKeyUsageIPSECTunnel, x509.ExtKeyUsageIPSECUser},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{serviceName},
	}
	caCertByte, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatalf("create ca certificate failed, error %v", err)
	}
	caCertBlock := pem.Block{Type: "CERTIFICATE", Bytes: caCertByte}
	caCertBlockBytes := pem.EncodeToMemory(&caCertBlock)
	return caCertBlockBytes
}

func SignCrossCert(privateKey *rsa.PrivateKey, caCertBlockBytes, certBlockBytes []byte) []byte {
	certBlock, _ := pem.Decode(certBlockBytes)
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		log.Fatalf("parse certificate failed, error %v", err)
	}
	caCertBlock, _ := pem.Decode(caCertBlockBytes)
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		log.Fatalf("parse ca certificate failed, error %v", err)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}
	before := time.Now()
	after := before.AddDate(1, 0, 0)

	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               cert.Subject,
		NotBefore:             before,
		NotAfter:              after,
		BasicConstraintsValid: true,
		IsCA:                  false,
		KeyUsage:              cert.KeyUsage,
		ExtKeyUsage:           cert.ExtKeyUsage,
		SignatureAlgorithm:    cert.SignatureAlgorithm,
		IPAddresses:           cert.IPAddresses,
		DNSNames:              cert.DNSNames,
	}
	crossCertByte, err := x509.CreateCertificate(rand.Reader, template, caCert, cert.PublicKey, privateKey)
	if err != nil {
		log.Fatalf("create ca certificate failed, error %v", err)
	}
	crossCertBlock := pem.Block{Type: "CERTIFICATE", Bytes: crossCertByte}
	crossCertBlockBytes := pem.EncodeToMemory(&crossCertBlock)
	return crossCertBlockBytes
}

func SignServerCert(csrBlockBytes, caCertBlockBytes []byte, caPrivateKey *rsa.PrivateKey) []byte {
	csrBlock, _ := pem.Decode(csrBlockBytes)
	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		log.Fatalf("parser csr failed, error %v", err)
	}
	caCertBlock, _ := pem.Decode(caCertBlockBytes)
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		log.Fatalf("parser caCert failed, error %v", err)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}
	before := time.Now()
	after := before.AddDate(1, 0, 0)
	serverCert := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               csr.Subject,
		NotBefore:             before,
		NotAfter:              after,
		BasicConstraintsValid: true,
		IsCA:                  false,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageContentCommitment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		SignatureAlgorithm:    csr.SignatureAlgorithm,
		IPAddresses:           csr.IPAddresses,
		DNSNames:              csr.DNSNames,
	}
	serverCertByte, err := x509.CreateCertificate(rand.Reader, &serverCert, caCert, csr.PublicKey, caPrivateKey)
	if err != nil {
		log.Fatalf("create server certificate failed, error %v", err)
	}
	serverCertBlock := pem.Block{Type: "CERTIFICATE", Bytes: serverCertByte}
	serverCertBlockBytes := pem.EncodeToMemory(&serverCertBlock)
	return serverCertBlockBytes
}

func GenerateCert(ipString, serviceName string) {
	privateKey, privateKeyBlockByte, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("get privateKey Block failed, error %v", err)
	}
	log.Println(string(privateKeyBlockByte))
	csrCertificate, err := generateCsr(privateKey, ipString, serviceName)
	if err != nil {
		log.Fatalf("generate csr failed, error %v", err)
	}
	log.Println(string(csrCertificate))
}

func generateCsr(rsaPriKey *rsa.PrivateKey, ipv4string, commonName string) ([]byte, error) {
	certRequest := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         commonName,
			Country:            nil,
			Province:           nil,
			Locality:           nil,
			Organization:       nil,
			OrganizationalUnit: nil,
		},
		IPAddresses:        []net.IP{net.ParseIP(ipv4string)},
		DNSNames:           []string{commonName},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	csrByte, err := x509.CreateCertificateRequest(rand.Reader, certRequest, rsaPriKey)
	if err != nil {
		log.Printf("create certificate request failed, error %v", err)
		return nil, err
	}
	csrBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrByte})
	return csrBytes, nil
}

func writeFile(path string, content []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("create or open file %q failed, error %v", path, err)
		return err
	}
	defer f.Close()
	if _, err := f.Write(content); err != nil {
		log.Printf("write file %q failed, error %v", path, err)
		return err
	}
	return nil
}

func generateCrossCert() {
	counts := 3

	caPrivateKeys := make([]*rsa.PrivateKey, counts, counts)
	caPrivateKeyBlockBytes := make([][]byte, counts, counts)
	caCertBlockBytes := make([][]byte, counts, counts)
	var err error

	for i := 0; i < counts; i++ {
		log.Printf("generate caPrivateKey %d", i)
		caPrivateKeys[i], caPrivateKeyBlockBytes[i], err = generatePrivateKey()
		if err != nil {
			log.Fatalf("generate privatekey %d failed, err %v", i, err)
		}
		writeFile(fmt.Sprintf("conf/certs/ca%d.key", i), caPrivateKeyBlockBytes[i])
		log.Printf("generate caCertBlockBytes %d", i)
		caCertBlockBytes[i] = SignCACert(caPrivateKeys[i], fmt.Sprintf("DevCAService%d", i))
		writeFile(fmt.Sprintf("conf/certs/ca%d.crt", i), caCertBlockBytes[i])
	}
	cert01 := SignCrossCert(caPrivateKeys[0], caCertBlockBytes[0], caCertBlockBytes[1])
	writeFile(fmt.Sprintf("conf/certs/ca%s.crt", "01"), cert01)
	cert10 := SignCrossCert(caPrivateKeys[1], caCertBlockBytes[1], caCertBlockBytes[0])
	writeFile(fmt.Sprintf("conf/certs/ca%s.crt", "10"), cert10)
	cert12 := SignCrossCert(caPrivateKeys[1], caCertBlockBytes[1], caCertBlockBytes[2])
	writeFile(fmt.Sprintf("conf/certs/ca%s.crt", "12"), cert12)
	cert21 := SignCrossCert(caPrivateKeys[2], caCertBlockBytes[2], caCertBlockBytes[1])
	writeFile(fmt.Sprintf("conf/certs/ca%s.crt", "21"), cert21)

	n := "1"
	serverPrivateKey1, serverPrivateKeyBlockByte1, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("get serverPrivateKey%s Block failed, error %v", n, err)
	}
	writeFile(fmt.Sprintf("conf/certs/server%s.key", n), serverPrivateKeyBlockByte1)

	log.Printf("generate serverCsr%s ", n)
	serverCsrBlockBytes1, err := generateCsr(serverPrivateKey1, "127.0.0.1", "DevelopService")
	if err != nil {
		log.Fatalf("get serverCsr%s failed, error %v", n, err)
	}
	log.Printf("generate serverCert%s ", n)
	//    serverCertBytes1 := SignServerCert(serverCsrBlockBytes1, cert01, caPrivateKeys[1])
	serverCertBytes1 := SignServerCert(serverCsrBlockBytes1, caCertBlockBytes[1], caPrivateKeys[1])
	writeFile(fmt.Sprintf("conf/certs/server%s.crt", n), serverCertBytes1)

	log.Printf("generate clientPrivateKey%s ", n)
	clientPrivateKey1, clientPrivateKeyBlockByte1, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("get clientPrivateKey%s Block failed, error %v", n, err)
	}
	writeFile(fmt.Sprintf("conf/certs/client%s.key", n), clientPrivateKeyBlockByte1)

	log.Printf("generate clientCsr%s ", n)
	clientCsrBlockBytes1, err := generateCsr(clientPrivateKey1, "127.0.0.1", "DevelopService")
	if err != nil {
		log.Fatalf("get serverCsr%s failed, error %v", n, err)
	}
	log.Printf("generate clientCert%s ", n)
	//    clientCertBytes1 := SignServerCert(clientCsrBlockBytes1, cert01, caPrivateKeys[1])
	clientCertBytes1 := SignServerCert(clientCsrBlockBytes1, caCertBlockBytes[1], caPrivateKeys[1])
	writeFile(fmt.Sprintf("conf/certs/client%s.crt", n), clientCertBytes1)
	n = "2"
	serverPrivateKey2, serverPrivateKeyBlockByte2, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("get serverPrivateKey%s Block failed, error %v", n, err)
	}
	writeFile(fmt.Sprintf("conf/certs/server%s.key", n), serverPrivateKeyBlockByte2)

	log.Printf("generate serverCsr%s ", n)
	serverCsrBlockBytes2, err := generateCsr(serverPrivateKey2, "127.0.0.1", "DevelopService")
	if err != nil {
		log.Fatalf("get serverCsr%s failed, error %v", n, err)
	}
	log.Printf("generate serverCert%s ", n)
	//    serverCertBytes2 := SignServerCert(serverCsrBlockBytes2, cert12, caPrivateKeys[2])
	serverCertBytes2 := SignServerCert(serverCsrBlockBytes2, caCertBlockBytes[2], caPrivateKeys[2])
	writeFile(fmt.Sprintf("conf/certs/server%s.crt", n), serverCertBytes2)

	log.Printf("generate clientPrivateKey%s ", n)
	clientPrivateKey2, clientPrivateKeyBlockByte2, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("get clientPrivateKey%s Block failed, error %v", n, err)
	}
	writeFile(fmt.Sprintf("conf/certs/client%s.key", n), clientPrivateKeyBlockByte2)

	log.Printf("generate clientCsr%s ", n)
	clientCsrBlockBytes2, err := generateCsr(clientPrivateKey2, "127.0.0.1", "DevelopService")
	if err != nil {
		log.Fatalf("get serverCsr%s failed, error %v", n, err)
	}
	log.Printf("generate clientCert%s ", n)
	//    clientCertBytes2 := SignServerCert(clientCsrBlockBytes2, cert12, caPrivateKeys[2])
	clientCertBytes2 := SignServerCert(clientCsrBlockBytes2, caCertBlockBytes[2], caPrivateKeys[2])
	writeFile(fmt.Sprintf("conf/certs/client%s.crt", n), clientCertBytes2)
}

func generateCerts(n string) {
	log.Printf("generate caPrivateKey%s ", n)
	caPrivateKey1, caPrivateKeyBlockByte1, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("get caPrivateKey%s Block failed, error %v", n, err)
	}
	writeFile(fmt.Sprintf("conf/certs/ca%s.key", n), caPrivateKeyBlockByte1)
	log.Printf("generate caCert%s ", n)
	caCertBytes1 := SignCACert(caPrivateKey1, "DevService")
	writeFile(fmt.Sprintf("conf/certs/ca%s.crt", n), caCertBytes1)

	log.Printf("generate serverPrivateKey%s ", n)
	serverPrivateKey1, serverPrivateKeyBlockByte1, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("get serverPrivateKey%s Block failed, error %v", n, err)
	}
	writeFile(fmt.Sprintf("conf/certs/server%s.key", n), serverPrivateKeyBlockByte1)

	log.Printf("generate serverCsr%s ", n)
	serverCsrBlockBytes1, err := generateCsr(serverPrivateKey1, "127.0.0.1", "DevelopService")
	if err != nil {
		log.Fatalf("get serverCsr%s failed, error %v", n, err)
	}
	log.Printf("generate serverCert%s ", n)
	serverCertBytes1 := SignServerCert(serverCsrBlockBytes1, caCertBytes1, caPrivateKey1)
	writeFile(fmt.Sprintf("conf/certs/server%s.crt", n), serverCertBytes1)

	log.Printf("generate clientPrivateKey%s ", n)
	clientPrivateKey1, clientPrivateKeyBlockByte1, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("get clientPrivateKey%s Block failed, error %v", n, err)
	}
	writeFile(fmt.Sprintf("conf/certs/client%s.key", n), clientPrivateKeyBlockByte1)

	log.Printf("generate clientCsr%s ", n)
	clientCsrBlockBytes1, err := generateCsr(clientPrivateKey1, "127.0.0.1", "DevelopService")
	if err != nil {
		log.Fatalf("get serverCsr%s failed, error %v", n, err)
	}
	log.Printf("generate clientCert%s ", n)
	clientCertBytes1 := SignServerCert(clientCsrBlockBytes1, caCertBytes1, caPrivateKey1)
	writeFile(fmt.Sprintf("conf/certs/client%s.crt", n), clientCertBytes1)
}

func main() {
	//    number := os.Args[1]
	//    generateCerts(number)
	/**
	  n := "3"
	  log.Printf("generate caPrivateKey%s ", n)
	  caPrivateKey1, caPrivateKeyBlockByte1, err := generatePrivateKey()
	  if err != nil {
	      log.Fatalf("get caPrivateKey%s Block failed, error %v", n, err)
	  }
	  writeFile(fmt.Sprintf("conf/certs/ca%s.key", n), caPrivateKeyBlockByte1)
	  n = "4"
	  log.Printf("generate caPrivateKey%s ", n)
	  caPrivateKey2, caPrivateKeyBlockByte2, err := generatePrivateKey()
	  if err != nil {
	      log.Fatalf("get caPrivateKey%s Block failed, error %v", n, err)
	  }
	  writeFile(fmt.Sprintf("conf/certs/ca%s.key", n), caPrivateKeyBlockByte2)
	  log.Printf("generate crossCaCert%s ", "34")
	  crossCertBytes := SignCrossCert(caPrivateKey1, caPrivateKey2, "DevelopService")
	  writeFile(fmt.Sprintf("conf/certs/ca%s.crt", "34"), crossCertBytes)

	  log.Printf("generate serverPrivateKey%s ", n)
	  serverPrivateKey1, serverPrivateKeyBlockByte1, err := generatePrivateKey()
	  if err != nil {
	      log.Fatalf("get serverPrivateKey%s Block failed, error %v", n, err)
	  }
	  writeFile(fmt.Sprintf("conf/certs/server%s.key", n), serverPrivateKeyBlockByte1)

	  log.Printf("generate serverCsr%s ", n)
	  serverCsrBlockBytes1, err := generateCsr(serverPrivateKey1, "127.0.0.1", "DevelopService")
	  if err != nil {
	      log.Fatalf("get serverCsr%s failed, error %v",n, err)
	  }
	  log.Printf("generate serverCert%s ", n)
	  serverCertBytes1 := SignServerCert(serverCsrBlockBytes1, crossCertBytes, caPrivateKey2)
	  writeFile(fmt.Sprintf("conf/certs/server%s.crt", n), serverCertBytes1)

	  log.Printf("generate clientPrivateKey%s ", n)
	  clientPrivateKey1, clientPrivateKeyBlockByte1, err := generatePrivateKey()
	  if err != nil {
	      log.Fatalf("get clientPrivateKey%s Block failed, error %v", n, err)
	  }
	  writeFile(fmt.Sprintf("conf/certs/client%s.key", n), clientPrivateKeyBlockByte1)

	  log.Printf("generate clientCsr%s ", n)
	  clientCsrBlockBytes1, err := generateCsr(clientPrivateKey1, "127.0.0.1", "DevelopService")
	  if err != nil {
	      log.Fatalf("get serverCsr%s failed, error %v", n, err)
	  }
	  log.Printf("generate clientCert%s ", n)
	  clientCertBytes1 := SignServerCert(clientCsrBlockBytes1, crossCertBytes, caPrivateKey2)
	  writeFile(fmt.Sprintf("conf/certs/client%s.crt", n), clientCertBytes1)
	*/
	generateCrossCert()
}
