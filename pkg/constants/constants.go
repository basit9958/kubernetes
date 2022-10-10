package constants

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// Data directories
const (
	DataDirDefault    = "/var/lib/k8s"
	WinDataDirDefault = "C:\\var\\lib\\k8s" // WinDataDirDefault is the data-dir for windows
)

var certDir string

const (

	// CACertName defines certificate name
	CACertName = "ca.crt"
	// CAKeyName defines certificate name
	CAKeyName = "ca.key"
	// APIServerCertName defines API's server certificate name
	APIServerCertName = "kubernetes.crt"
	// APIServerKeyName defines API's server key name
	APIServerKeyName = "kubernetes.key"
	// EtcdServerCertName defines etcd's server certificate name
	EtcdServerCertName = "kubernetes.crt"
	// EtcdServerKeyName defines etcd's server key name
	EtcdServerKeyName = "kubernetes.key"
	// EtcdPeerCertName defines etcd's peer certificate name
	EtcdPeerCertName = "kubernetes.crt"
	// ServiceAccountKeyName defines service account key name
	ServiceAccountKeyName = "sa.key"
	// ServiceAccountCertName defines service account cert name
	ServiceAccountCertName = "sa.crt"
	// DataDirMode is the expected directory permissions for DataDirDefault
	DataDirMode = 0755
	// EtcdDataDirMode is the expected directory permissions for EtcdDataDir.
	EtcdDataDirMode = 0700
	// CertRootDirMode is the expected directory permissions for CertRootDir.
	CertRootDirMode = 0751
	// CertMode is the expected permissions for certificates.
	CertMode = 0666
	// CertSecureMode is the expected file permissions for secure files, this is applicable to files like: admin.conf and certificate files
	CertSecureMode = 0666
)

// CfgVars is a struct that holds all the required config variables
type CfgVars struct {
	AdminKubeConfigPath string // The cluster admin kubeconfig location
	BinDir              string // location for all pki related binaries
	CertRootDir         string // CertRootDir defines the root location for all pki related artifacts
	DataDir             string // Data directory of our binary
	EtcdCertDir         string // EtcdCertDir contains etcd certificates
	EtcdDataDir         string // EtcdDataDir contains etcd state
	ManifestsDir        string // location for all stack manifests

}

// GetConfig returns the pointer to a Config struct
func GetConfig(dataDir string) CfgVars {

	if dataDir == "" {
		switch runtime.GOOS {
		case "windows":
			dataDir = WinDataDirDefault
		default:
			dataDir = DataDirDefault
		}
	}

	// fetch absolute path for dataDir
	dataDir, err := filepath.Abs(dataDir)
	if err != nil {
		panic(err)
	}
	switch runtime.GOOS {
	case "windows":
		dataDir = strings.Trim(dataDir, "C:")
	}

	certDir = formatPath(dataDir, "pki")

	return CfgVars{
		AdminKubeConfigPath: formatPath(certDir, "admin.conf"),
		BinDir:              formatPath(dataDir, "bin"),
		CertRootDir:         certDir,
		DataDir:             dataDir,
		EtcdCertDir:         formatPath(certDir, "etcd"),
		EtcdDataDir:         formatPath(dataDir, "etcd"),
	}
}

func formatPath(dir string, file string) string {
	return fmt.Sprintf("%s/%s", dir, file)
}
