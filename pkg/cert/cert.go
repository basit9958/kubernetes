package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	duration365d = time.Hour * 24 * 365
)

func Generatecert() {
	now := time.Now()
	// Generate the CA configuration file, certificate, and private key
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:       []string{"Kubernetes"},
			OrganizationalUnit: []string{"CA"},
			Country:            []string{"US"},
			Locality:           []string{"San Francisco"},
			StreetAddress:      []string{"Golden Gate Bridge"},
			PostalCode:         []string{"94016"},
			CommonName:         "Kubernetes",
		},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(duration365d * 10).UTC(),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("Failed to create private and public key")
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		log.Fatalf("Failed to create ca")
	}
	certOut, err := os.Create("ca.pem")
	if err != nil {
		log.Fatalf("Failed to open ca.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		log.Fatalf("Failed to write data to ca.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing ca.pem: %v", err)
	}
	log.Print("wrote ca.pem\n")

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	keyOut, err := os.OpenFile("ca-key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open ca-key.pem for writing: %v", err)
		return
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(caPrivKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing key.pem: %v", err)
	}
	log.Print("wrote ca-key.pem\n")

	admincertsetup(caBytes, caPrivKey)
	KubeControllerManagercertsetup(caBytes, caPrivKey)
	Kubeapiservercertsetup(caBytes, caPrivKey)
	ServiceAccountcertsetup(caBytes, caPrivKey)

	//TODO: Move a directory(/var/lib/kubernetes/) and Move all the certificates to that directory
}

func admincertsetup(ca []byte, caPrivKey *rsa.PrivateKey) {
	now := time.Now()
	caCertificate, err := x509.ParseCertificate(ca)
	if err != nil {
		return
	}
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization:       []string{"system:masters"},
			OrganizationalUnit: []string{"Chaos Mesh"},
			Country:            []string{"US"},
			Locality:           []string{"San Francisco"},
			StreetAddress:      []string{"Golden Gate Bridge"},
			PostalCode:         []string{"94016"},
			CommonName:         "admin",
		},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(duration365d * 10).UTC(),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCertificate, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return
	}

	certOut, err := os.Create("admin.pem")
	if err != nil {
		log.Fatalf("Failed to open admin.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		log.Fatalf("Failed to write data to admin.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing admin.pem: %v", err)
	}
	log.Print("wrote admin.pem\n")

	keyOut, err := os.OpenFile("admin-key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open admin-key.pem for writing: %v", err)
		return
	}
	certprivBytes, err := x509.MarshalPKCS8PrivateKey(certPrivKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: certprivBytes}); err != nil {
		log.Fatalf("Failed to write data to admin-key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing admin-key.pem: %v", err)
	}
	log.Print("wrote admin-key.pem\n")
}
func KubeControllerManagercertsetup(ca []byte, caPrivKey *rsa.PrivateKey) {
	now := time.Now()
	caCertificate, err := x509.ParseCertificate(ca)
	if err != nil {
		return
	}
	KubeControllerManagercert := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization:       []string{"system:kube-controller-manage"},
			OrganizationalUnit: []string{"Chaos Mesh"},
			Country:            []string{"US"},
			Locality:           []string{"San Francisco"},
			StreetAddress:      []string{"Golden Gate Bridge"},
			PostalCode:         []string{"94016"},
			CommonName:         "system:kube-controller-manager",
		},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(duration365d * 10).UTC(),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	KubeControllerManagercertPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	KubeControllerManagercertBytes, err := x509.CreateCertificate(rand.Reader, KubeControllerManagercert, caCertificate, &KubeControllerManagercertPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return
	}

	certOut, err := os.Create("kube-controller-manager.pem")
	if err != nil {
		log.Fatalf("Failed to open kube-controller-manager.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: KubeControllerManagercertBytes}); err != nil {
		log.Fatalf("Failed to write data to kube-controller-manager.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing kube-controller-manager.pem: %v", err)
	}
	log.Print("wrote kube-controller-manager.pem\n")

	keyOut, err := os.OpenFile("kube-controller-manager-key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open kube-controller-manager-key.pem for writing: %v", err)
		return
	}
	KubeControllerManagercertprivBytes, err := x509.MarshalPKCS8PrivateKey(KubeControllerManagercertPrivKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: KubeControllerManagercertprivBytes}); err != nil {
		log.Fatalf("Failed to write data to kube-controller-manager-key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing kube-controller-manager-key.pem: %v", err)
	}
	log.Print("wrote kube-controller-manager-key.pem\n")
}

func Kubeapiservercertsetup(ca []byte, caPrivKey *rsa.PrivateKey) {
	now := time.Now()
	caCertificate, err := x509.ParseCertificate(ca)
	if err != nil {
		return
	}
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		DNSNames:     []string{"kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster", "kubernetes.svc.cluster.local"},
		IPAddresses:  []net.IP{net.ParseIP("10.32.0.1"), net.ParseIP("10.240.0.10"), net.ParseIP("10.240.0.11"), net.ParseIP("10.240.0.12"), net.ParseIP("127.0.0.1")},
		Subject: pkix.Name{
			Organization:       []string{"Kubernetes"},
			OrganizationalUnit: []string{"Chaos Mesh"},
			Country:            []string{"US"},
			Locality:           []string{"San Francisco"},
			StreetAddress:      []string{"Golden Gate Bridge"},
			PostalCode:         []string{"94016"},
			CommonName:         "kubernetes",
		},
		NotBefore:   now.UTC(),
		NotAfter:    now.Add(duration365d * 10).UTC(),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
	}
	KubeapiservercertPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	KubeapiservercertBytes, err := x509.CreateCertificate(rand.Reader, cert, caCertificate, &KubeapiservercertPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return
	}

	certOut, err := os.Create("kubernetes.pem")
	if err != nil {
		log.Fatalf("Failed to open kubernetes.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: KubeapiservercertBytes}); err != nil {
		log.Fatalf("Failed to write data to kubernetes.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing kubernetes.pem: %v", err)
	}
	log.Print("wrote kubernetes.pem\n")

	keyOut, err := os.OpenFile("kubernetes-key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open kubernetes-key.pem for writing: %v", err)
		return
	}
	KubeapiservercertprivBytes, err := x509.MarshalPKCS8PrivateKey(KubeapiservercertPrivKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: KubeapiservercertprivBytes}); err != nil {
		log.Fatalf("Failed to write data to kubernetes-key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing kubernetes-key.pem: %v", err)
	}
	log.Print("wrote kubernetes-key.pem\n")
}
func ServiceAccountcertsetup(ca []byte, caPrivKey *rsa.PrivateKey) {
	now := time.Now()
	caCertificate, err := x509.ParseCertificate(ca)
	if err != nil {
		return
	}
	ServiceAccountcert := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization:       []string{"Kubernetes"},
			OrganizationalUnit: []string{"Chaos Mesh"},
			Country:            []string{"US"},
			Locality:           []string{"San Francisco"},
			StreetAddress:      []string{"Golden Gate Bridge"},
			PostalCode:         []string{"94016"},
			CommonName:         "service-accounts",
		},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(duration365d * 10).UTC(),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	ServiceAccountcertPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	ServiceAccountcertBytes, err := x509.CreateCertificate(rand.Reader, ServiceAccountcert, caCertificate, &ServiceAccountcertPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return
	}

	certOut, err := os.Create("service-account.pem")
	if err != nil {
		log.Fatalf("Failed to open service-account.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: ServiceAccountcertBytes}); err != nil {
		log.Fatalf("Failed to write data to service-account.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing service-account.pem: %v", err)
	}
	log.Print("wrote service-account.pem\n")

	keyOut, err := os.OpenFile("service-account-key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open service-account-key-key.pem for writing: %v", err)
		return
	}
	ServiceAccountcertprivBytes, err := x509.MarshalPKCS8PrivateKey(ServiceAccountcertPrivKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: ServiceAccountcertprivBytes}); err != nil {
		log.Fatalf("Failed to write data toservice-account-key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing service-account-key.pem: %v", err)
	}
	log.Print("wrote service-account-key.pem\n")
}
