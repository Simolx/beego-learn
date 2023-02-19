package main

import (
	"crypto/x509/pkix"
	"net"

	"example.com/lx/beego/dev/utils"
)

func main() {
	kubernetesCaSubject := pkix.Name{
		CommonName: "kubernetes",
	}
	kubernetesCaCertBytes, kubernetesCaPrivateKeyBytes := utils.SignCACert(kubernetesCaSubject, nil, []string{"kubernetes"})
	utils.WriteFile("conf/kubernetes/ca.crt", kubernetesCaCertBytes)
	utils.WriteFile("conf/kubernetes/ca.key", kubernetesCaPrivateKeyBytes)
	/**
	frontProxyCaSubject := pkix.Name{
		CommonName: "front-proxy-ca",
	}
	frontProxyCaCertBytes, frontProxyCaPrivateKeyBytes := utils.SignCACert(frontProxyCaSubject, nil, []string{"front-proxy-ca"})
	utils.WriteFile("conf/kubernetes/front-proxy-ca.crt", frontProxyCaCertBytes)
	utils.WriteFile("conf/kubernetes/front-proxy-ca.key", frontProxyCaPrivateKeyBytes)
	*/
	apiServerCertSubject := pkix.Name{
		CommonName: "kube-apiserver",
	}
	ipAddresses := []net.IP{
		net.ParseIP("10.97.0.1"),
		net.ParseIP("10.98.66.31"),
	}
	dnsNames := []string{
		"controlplane",
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc.cluster.local",
		"master2",
	}
	apiServerCertBlockBytes, apiServerPrivateKeyBytes := utils.SignServerCert(apiServerCertSubject, ipAddresses, dnsNames, kubernetesCaCertBytes, kubernetesCaPrivateKeyBytes)
	utils.WriteFile("conf/kubernetes/apiserver.crt", apiServerCertBlockBytes)
	utils.WriteFile("conf/kubernetes/apiserver.key", apiServerPrivateKeyBytes)
}
