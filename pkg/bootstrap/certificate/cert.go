package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"k8s.io/kubernetes/pkg/bootstrap/certificate/utils"
	"k8s.io/kubernetes/pkg/constants"
	"os"
	"path/filepath"

	"github.com/cloudflare/cfssl/certinfo"
	"github.com/cloudflare/cfssl/cli"
	"github.com/cloudflare/cfssl/cli/genkey"
	"github.com/cloudflare/cfssl/cli/sign"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/signer"
	"github.com/sirupsen/logrus"
)

// Request defines the certificate request fields
type Request struct {
	Name      string
	CN        string
	O         string
	CAKey     string
	CACert    string
	Hostnames []string
}

// Certificate is a helper struct to be able to return the created key and cert data
type Certificate struct {
	Key  string
	Cert string
}

// Manager is the certificate manager
type Manager struct {
	Cfg constants.CfgVars
}

// CreateCACertAndKeyFiles makes sure the given CA certs and key is created.
func (m *Manager) CreateCACertAndKeyFiles(name, cn string) error {
	keyFile := filepath.Join(m.Cfg.CertRootDir, fmt.Sprintf("%s.key", name))
	certFile := filepath.Join(m.Cfg.CertRootDir, fmt.Sprintf("%s.crt", name))

	if utils.Exists(keyFile) && utils.Exists(certFile) {
		return nil
	}

	req := new(csr.CertificateRequest)
	req.KeyRequest = csr.NewKeyRequest()
	req.KeyRequest.A = "rsa"
	req.KeyRequest.S = 2048
	req.CN = cn
	req.CA = &csr.CAConfig{
		Expiry: "8760h",
	}
	cert, _, key, err := initca.New(req)
	if err != nil {
		return err
	}

	err = os.WriteFile(keyFile, key, constants.CertSecureMode)
	if err != nil {
		return err
	}

	err = os.WriteFile(certFile, cert, constants.CertMode)
	if err != nil {
		return err
	}

	return nil
}

// CreateCertAndKeyFilesWithCA creates the specified certificate if it does not already exist
func (m *Manager) CreateCertAndKeyFilesWithCA(certReq Request) (Certificate, error) {
	keyFile := filepath.Join(m.Cfg.CertRootDir, fmt.Sprintf("%s.key", certReq.Name))
	certFile := filepath.Join(m.Cfg.CertRootDir, fmt.Sprintf("%s.crt", certReq.Name))

	// if regenerateCert returns true, it means we need to create the certs
	if m.regenerateCert(keyFile, certFile) {
		logrus.Debugf("creating certificate %s", certFile)
		req := csr.CertificateRequest{
			KeyRequest: csr.NewKeyRequest(),
			CN:         certReq.CN,
			Names: []csr.Name{
				{O: certReq.O},
			},
		}

		req.KeyRequest.A = "rsa"
		req.KeyRequest.S = 2048
		req.Hosts = utils.Unique(certReq.Hostnames)

		var key, csrBytes []byte
		g := &csr.Generator{Validator: genkey.Validator}
		csrBytes, key, err := g.ProcessRequest(&req)
		if err != nil {
			return Certificate{}, err
		}

		config := cli.Config{
			CAFile:    certReq.CACert,
			CAKeyFile: certReq.CAKey,
		}
		fmt.Println(certReq.CACert)
		s, err := sign.SignerFromConfig(config)
		if err != nil {
			return Certificate{}, err
		}
		var cert []byte
		signReq := signer.SignRequest{
			Request: string(csrBytes),
			Profile: "kubernetes",
		}

		cert, err = s.Sign(signReq)
		if err != nil {
			return Certificate{}, err
		}
		c := Certificate{
			Key:  string(key),
			Cert: string(cert),
		}
		err = os.WriteFile(keyFile, key, constants.CertSecureMode)
		if err != nil {
			return Certificate{}, err
		}
		err = os.WriteFile(certFile, cert, constants.CertMode)
		if err != nil {
			return Certificate{}, err
		}

		return c, nil
	}

	cert, err := os.ReadFile(certFile)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to read ca cert %s for %s: %w", certFile, certReq.Name, err)
	}
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to read ca key %s for %s: %w", keyFile, certReq.Name, err)
	}

	return Certificate{
		Key:  string(key),
		Cert: string(cert),
	}, nil

}

// regenerateCert returns false when certificates do not exist
func (m *Manager) regenerateCert(keyFile string, certFile string) bool {
	var err error
	// if certificate & key don't exist, return true, in order to generate certificates
	if !utils.Exists(keyFile) && !utils.Exists(certFile) {
		fmt.Printf("does not exist")
		return true
	}
	if _, err = certinfo.ParseCertificateFile(certFile); err != nil {
		logrus.Warnf("unable to parse certificate file at %s: %v", certFile, err)
		return true
	}
	return false
}

func (m *Manager) CreateKeyPair(name string, CfgVars constants.CfgVars) error {
	keyFile := filepath.Join(CfgVars.CertRootDir, fmt.Sprintf("%s.key", name))
	pubFile := filepath.Join(CfgVars.CertRootDir, fmt.Sprintf("%s.pub", name))

	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return err
	}

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	outFile, err := os.OpenFile(keyFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, constants.CertSecureMode)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = pem.Encode(outFile, privateKey)
	if err != nil {
		return err
	}

	// note to the next reader: key.Public() != key.PublicKey
	pubBytes, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		return err
	}

	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}

	pemfile, err := os.Create(pubFile)
	if err != nil {
		return err
	}
	defer pemfile.Close()

	err = pem.Encode(pemfile, pemkey)
	if err != nil {
		return err
	}

	return nil
}
