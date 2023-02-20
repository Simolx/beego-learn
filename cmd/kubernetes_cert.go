package main

import (
	"crypto/x509/pkix"
	"net"
	"path/filepath"

	"example.com/lx/beego/dev/utils"
)

const (
	certLocation = "out/kubernetes/pki"
)

func main() {
	kubernetesCaCertInfo := &utils.CertInfo{
		CertName: "ca",
		Subject: pkix.Name{
			CommonName: "kubernetes",
		},
		IpAddresses: nil,
		DnsNames:    []string{"kubernetes"},
		Location:    certLocation,
		ServerCerts: make([]*utils.CertInfo, 0, 5),
	}
	kubernetesCaCertInfo.ServerCerts = append(kubernetesCaCertInfo.ServerCerts,
		&utils.CertInfo{
			CertName: "apiserver",
			Subject: pkix.Name{
				CommonName: "kube-apiserver",
			},
			IpAddresses: []net.IP{
				net.ParseIP("10.97.0.1"),
				net.ParseIP("10.98.66.31"),
			},
			DnsNames: []string{
				"controlplane",
				"kubernetes",
				"kubernetes.default",
				"kubernetes.default.svc",
				"kubernetes.default.svc.cluster.local",
				"master1",
			},
			Location: certLocation,
		})
	kubernetesCaCertInfo.ServerCerts = append(kubernetesCaCertInfo.ServerCerts,
		&utils.CertInfo{
			CertName: "apiserver-kubelet-client",
			Subject: pkix.Name{
				Organization: []string{"system:masters"},
				CommonName:   "kube-apiserver-kubelet-client",
			},
			IpAddresses: nil,
			DnsNames:    nil,
			Location:    certLocation,
		})
	kubernetesCaCertInfo.SignCerts()

	frontProxyCaCertInfo := &utils.CertInfo{
		CertName: "front-proxy-ca",
		Subject: pkix.Name{
			CommonName: "front-proxy-ca",
		},
		IpAddresses: nil,
		DnsNames:    []string{"front-proxy-ca"},
		Location:    certLocation,
		ServerCerts: make([]*utils.CertInfo, 0, 5),
	}
	frontProxyCaCertInfo.ServerCerts = append(frontProxyCaCertInfo.ServerCerts,
		&utils.CertInfo{
			CertName: "front-proxy-client",
			Subject: pkix.Name{
				CommonName: "front-proxy-client",
			},
			IpAddresses: nil,
			DnsNames:    nil,
			Location:    certLocation,
		})
	frontProxyCaCertInfo.SignCerts()

	etcdCaCertInfo := &utils.CertInfo{
		CertName: "ca",
		Subject: pkix.Name{
			CommonName: "etcd-ca",
		},
		IpAddresses: nil,
		DnsNames:    []string{"etcd-ca"},
		Location:    filepath.Join(certLocation, "etcd"),
		ServerCerts: make([]*utils.CertInfo, 0, 5),
	}
	etcdCaCertInfo.ServerCerts = append(etcdCaCertInfo.ServerCerts,
		&utils.CertInfo{
			CertName: "apiserver-etcd-client",
			Subject: pkix.Name{
				Organization: []string{"system:masters"},
				CommonName:   "kube-apiserver-etcd-client",
			},
			IpAddresses: nil,
			DnsNames:    nil,
			Location:    certLocation,
		})
	etcdCaCertInfo.ServerCerts = append(etcdCaCertInfo.ServerCerts,
		&utils.CertInfo{
			CertName: "healthcheck-client",
			Subject: pkix.Name{
				Organization: []string{"system:masters"},
				CommonName:   "kube-etcd-healthcheck-client",
			},
			IpAddresses: nil,
			DnsNames:    nil,
			Location:    filepath.Join(certLocation, "etcd"),
		})
	etcdCaCertInfo.ServerCerts = append(etcdCaCertInfo.ServerCerts,
		&utils.CertInfo{
			CertName: "peer",
			Subject: pkix.Name{
				CommonName: "master1",
			},
			IpAddresses: []net.IP{
				net.ParseIP("10.98.66.30"),
				net.ParseIP("127.0.0.1"),
				net.ParseIP("0:0:0:0:0:0:0:1"),
			},
			DnsNames: []string{
				"localhost",
				"master1",
			},
			Location: filepath.Join(certLocation, "etcd"),
		})
	etcdCaCertInfo.ServerCerts = append(etcdCaCertInfo.ServerCerts,
		&utils.CertInfo{
			CertName: "server",
			Subject: pkix.Name{
				CommonName: "master1",
			},
			IpAddresses: []net.IP{
				net.ParseIP("10.98.66.30"),
				net.ParseIP("127.0.0.1"),
				net.ParseIP("0:0:0:0:0:0:0:1"),
			},
			DnsNames: []string{
				"localhost",
				"master1",
			},
			Location: filepath.Join(certLocation, "etcd"),
		})
	etcdCaCertInfo.SignCerts()
}
