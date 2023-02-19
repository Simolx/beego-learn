package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"time"
)

type CertConfig struct {
	ServerCert     string
	ServerKey      string
	ServerPassword string
	CaCert         string
}

const (
	certVersionV3 = 3
)

func generatePrivateKey() (*rsa.PrivateKey, []byte, error) {
	priKey4096 := 4096
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, priKey4096)
	if err != nil {
		log.Printf("generate key failed, %v", err)
		return nil, nil, err
	}
	pkcsPrivateKey, err := x509.MarshalPKCS8PrivateKey(rsaPrivateKey)
	if err != nil {
		log.Printf("create private key failed, error %v", err)
		return nil, nil, err
	}
	privateBlock := pem.Block{Type: "RSA PRIVATE KEY", Bytes: pkcsPrivateKey}
	return rsaPrivateKey, pem.EncodeToMemory(&privateBlock), nil
}

func SignCACert(pkixName pkix.Name, ipAddresses []net.IP, dnsNames []string) ([]byte, []byte) {
	privateKey, rsaPrivateKeyBytes, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("generate private key failed in sign ca cert, %v", err)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}
	before := time.Now()
	after := before.AddDate(1, 0, 0)
	template := x509.Certificate{
		Version:      certVersionV3,
		SerialNumber: serialNumber,
		/**
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Province:           []string{"BeiJing"},
			Organization:       []string{"devCompany"},
			OrganizationalUnit: []string{"devTeam"},
			CommonName:         serviceName,
		},
		*/
		Subject:               pkixName,
		NotBefore:             before,
		NotAfter:              after,
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageCodeSigning, x509.ExtKeyUsageEmailProtection, x509.ExtKeyUsageIPSECTunnel, x509.ExtKeyUsageIPSECUser},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		// IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	if ipAddresses != nil {
		template.IPAddresses = ipAddresses
	}
	if dnsNames != nil {
		template.DNSNames = dnsNames
	}
	caCertByte, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatalf("create ca certificate failed, error %v", err)
	}
	caCertBlock := pem.Block{Type: "CERTIFICATE", Bytes: caCertByte}
	caCertBlockBytes := pem.EncodeToMemory(&caCertBlock)
	return caCertBlockBytes, rsaPrivateKeyBytes
}

func generateCsr(rsaPriKey *rsa.PrivateKey, pkixName pkix.Name, ipAddresses []net.IP, dnsNames []string) ([]byte, error) {
	certRequest := &x509.CertificateRequest{
		Subject:            pkixName,
		IPAddresses:        ipAddresses,
		DNSNames:           dnsNames,
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

func SignServerCert(pkixName pkix.Name, ipAddress []net.IP, dnsNames []string, caCertBlockBytes []byte, caPrivateKeyBlockBytes []byte) ([]byte, []byte) {
	serverPrivateKey, serverRsaPrivateKeyBytes, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("generate private key failed in sign server cert, %v", err)
	}

	csrBlockBytes, err := generateCsr(serverPrivateKey, pkixName, ipAddress, dnsNames)
	if err != nil {
		log.Fatalf("generate csr failed in sign server cert, %v", err)
	}
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
	caPrivateKyeBlock, _ := pem.Decode(caPrivateKeyBlockBytes)
	caPrivateKey, err := x509.ParsePKCS8PrivateKey(caPrivateKyeBlock.Bytes)
	if err != nil {
		log.Fatalf("parser private key failed in sign server cert, %v", err)
	}
	serverCertByte, err := x509.CreateCertificate(rand.Reader, &serverCert, caCert, csr.PublicKey, caPrivateKey)
	if err != nil {
		log.Fatalf("create server certificate failed, error %v", err)
	}
	serverCertBlock := pem.Block{Type: "CERTIFICATE", Bytes: serverCertByte}
	serverCertBlockBytes := pem.EncodeToMemory(&serverCertBlock)
	return serverCertBlockBytes, serverRsaPrivateKeyBytes
}