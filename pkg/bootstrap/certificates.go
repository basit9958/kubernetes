package bootstrap

import (
	"context"
	"encoding/base64"
	"fmt"
	"k8s.io/kubernetes/pkg/bootstrap/certificate"
	"k8s.io/kubernetes/pkg/constants"
	"net"
	"os"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const kubeConfigAPIUrl = "https://127.0.0.1:6443"

var kubeconfigTemplate = template.Must(template.New("kubeconfig").Parse(`
apiVersion: v1
clusters:
- cluster:
    server: {{.URL}}
    certificate-authority-data: {{.CACert}}
  name: local
contexts:
- context:
    cluster: local
    namespace: default
    user: {{.User}}
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: {{.User}}
  user:
    client-certificate-data: {{.ClientCert}}
    client-key-data: {{.ClientKey}}
`))

// Certificates is a struct to manage all  certs
type Certificates struct {
	CACert      string
	CertManager certificate.Manager
	CfgVars     constants.CfgVars
}

// Init initializes all the required certificate
func (c *Certificates) Init(ctx context.Context) error {
	eg, _ := errgroup.WithContext(ctx)
	// Common CA
	caCertPath := filepath.Join(c.CfgVars.CertRootDir, "ca.crt")
	caCertKey := filepath.Join(c.CfgVars.CertRootDir, "ca.key")

	if err := c.CertManager.CreateCACertAndKeyFiles("ca", "Kubernetes"); err != nil {
		return err
	}

	// We need CA cert loaded to generate client configs
	logrus.Debugf("CA key and cert exists, loading")
	cert, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read ca cert: %w", err)
	}
	c.CACert = string(cert)

	eg.Go(func() error {
		// admin cert & kubeconfig
		adminReq := certificate.Request{
			Name:   "admin",
			CN:     "admin",
			O:      "system:masters",
			CACert: caCertPath,
			CAKey:  caCertKey,
		}
		adminCert, err := c.CertManager.CreateCertAndKeyFilesWithCA(adminReq)
		if err != nil {
			fmt.Printf("", err)
			return err
		}

		if err := kubeConfig(c.CfgVars.AdminKubeConfigPath, kubeConfigAPIUrl, c.CACert, adminCert.Cert, adminCert.Key, adminReq.CN); err != nil {
			return err
		}

		err = c.CertManager.CreateKeyPair("sa", c.CfgVars)
		if err != nil {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		ccmReq := certificate.Request{
			Name:   "ccm",
			CN:     "system:kube-controller-manager",
			O:      "system:kube-controller-manager",
			CACert: caCertPath,
			CAKey:  caCertKey,
		}
		ccmCert, err := c.CertManager.CreateCertAndKeyFilesWithCA(ccmReq)
		if err != nil {
			return err
		}

		return kubeConfig(filepath.Join(c.CfgVars.CertRootDir, "ccm.conf"), kubeConfigAPIUrl, c.CACert, ccmCert.Cert, ccmCert.Key, ccmReq.CN)
	})

	hostnames := []string{
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc.cluster",
		"kubernetes.default.svc.cluster.local",
		"localhost",
		"127.0.0.1",
	}

	localIPs, err := detectLocalIPs()
	if err != nil {
		return fmt.Errorf("error detecting local IP: %w", err)
	}
	hostnames = append(hostnames, localIPs...)

	eg.Go(func() error {
		serverReq := certificate.Request{
			Name:      "kubernetes",
			CN:        "kubernetes",
			O:         "kubernetes",
			CACert:    caCertPath,
			CAKey:     caCertKey,
			Hostnames: hostnames,
		}
		_, err = c.CertManager.CreateCertAndKeyFilesWithCA(serverReq)
		return err
	})

	eg.Go(func() error {
		saReq := certificate.Request{
			Name:   "sa",
			CN:     "service-accounts",
			O:      "kubernetes",
			CACert: caCertPath,
			CAKey:  caCertKey,
		}
		_, err := c.CertManager.CreateCertAndKeyFilesWithCA(saReq)
		return err
	})

	eg.Go(func() error {
		// Front proxy CA
		if err := c.CertManager.CreateCACertAndKeyFiles("front-proxy-ca", "kubernetes-front-proxy-ca"); err != nil {
			return err
		}

		proxyCertPath, proxyCertKey := filepath.Join(c.CfgVars.CertRootDir, "front-proxy-ca.crt"), filepath.Join(c.CfgVars.CertRootDir, "front-proxy-ca.key")

		proxyClientReq := certificate.Request{
			Name:   "front-proxy-client",
			CN:     "front-proxy-client",
			O:      "front-proxy-client",
			CACert: proxyCertPath,
			CAKey:  proxyCertKey,
		}
		_, err := c.CertManager.CreateCertAndKeyFilesWithCA(proxyClientReq)

		return err
	})

	return eg.Wait()
}

func detectLocalIPs() ([]string, error) {
	var localIPs []string
	addrs, err := net.LookupIP("localhost")
	if err != nil {
		return nil, err
	}

	if hostname, err := os.Hostname(); err == nil {
		hostnameAddrs, err := net.LookupIP(hostname)
		if err == nil {
			addrs = append(addrs, hostnameAddrs...)
		}
	}

	for _, addr := range addrs {
		if addr.To4() != nil {
			localIPs = append(localIPs, addr.String())
		}
	}
	return localIPs, nil
}

func kubeConfig(dest, url, caCert, clientCert, clientKey, user string) error {
	data := struct {
		URL        string
		CACert     string
		ClientCert string
		ClientKey  string
		User       string
	}{
		URL:        url,
		CACert:     base64.StdEncoding.EncodeToString([]byte(caCert)),
		ClientCert: base64.StdEncoding.EncodeToString([]byte(clientCert)),
		ClientKey:  base64.StdEncoding.EncodeToString([]byte(clientKey)),
		User:       user,
	}

	output, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, constants.CertSecureMode)
	if err != nil {
		return err
	}
	defer output.Close()

	if err = kubeconfigTemplate.Execute(output, &data); err != nil {
		return err
	}

	return nil
}
