package etcd

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/server/v3/embed"
	"go.uber.org/zap"
	"k8s.io/kubernetes/pkg/constants"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	DefaultPeerPort   int    = 2380
	DefaultClientPort int    = 2379
	DefaultScheme     string = "https"
)

//EtcdConfig is a struct that holds all the config variables required for *embed.config struct
type EtcdConfig struct {
	// Human-readable name for this member.
	Name string `yaml:"name,omitempty"`

	// Path to the data directory.
	DataDir string `yaml:"data-dir,omitempty"`

	// Path to the dedicated wal directory.
	WalDir string `yaml:"wal-dir,omitempty"`

	// List of comma separated URLs to listen on for peer traffic.
	ListenPeerUrls string `yaml:"listen-peer-urls,omitempty"`

	// List of comma separated URLs to listen on for client traffic.
	ListenClientUrls string `yaml:"listen-client-urls,omitempty"`

	// List of this member's peer URLs to advertise to the rest of the cluster.
	// The URLs needed to be a comma-separated list.
	InitialAdvertisePeerUrls string `yaml:"initial-advertise-peer-urls,omitempty"`

	// List of this member's client URLs to advertise to the public.
	// The URLs needed to be a comma-separated list.
	AdvertiseClientUrls string `yaml:"advertise-client-urls,omitempty"`

	// Initial cluster configuration for bootstrapping.
	InitialCluster string `yaml:"initial-cluster,omitempty"`

	// Initial cluster token for the etcd cluster during bootstrap.
	InitialClusterToken string `yaml:"initial-cluster-token,omitempty"`

	// Initial cluster state ('new' or 'existing').
	InitialClusterState string `yaml:"initial-cluster-state,omitempty"`

	ClientTransportSecurity ClientTransportSecurityInfo `yaml:"client-transport-security,omitempty"`

	PeerTransportSecurity PeerTransportSecurityInfo `yaml:"peer-transport-security,omitempty"`

	// Enable debug-level logging for etcd.
	Debug bool `yaml:"debug,omitempty"`

	// Maximum number of snapshot files to retain (0 is unlimited).
	MaxSnapshots int `yaml:"max-snapshots,omitempty"`

	// Maximum number of wal files to retain (0 is unlimited).
	MaxWals int `yaml:"max-wals,omitempty"`

	// Contains required Path for certificates and directories
	cfg constants.CfgVars
}

//ClientTransportSecurityInfo is a struct that holds the path of required client certificates
type ClientTransportSecurityInfo struct {

	// Path to the client server TLS cert file.
	CertFile string `yaml:"cert-file,omitempty"`

	// Path to the client server TLS key file.
	KeyFile string `yaml:"key-file,omitempty"`

	// Enable client cert authentication.
	ClientCertAuth bool `yaml:"client-cert-auth,omitempty"`

	// Path to the client server TLS trusted CA cert file.
	TrustedCaFile string `yaml:"trusted-ca-file,omitempty"`

	// Client TLS using generated certificates
	AutoTls bool `yaml:"auto-tls,omitempty"`
}

//PeerTransportSecurityInfo is a struct that holds the path of required peer certificates
type PeerTransportSecurityInfo struct {
	// Path to the peer server TLS cert file.
	CertFile string `yaml:"cert-file,omitempty"`

	// Path to the peer client server TLS key file.
	KeyFile string `yaml:"key-file,omitempty"`

	// Enable peer client cert authentication.
	ClientCertAuth bool `yaml:"client-cert-auth,omitempty"`

	// Path to the peer client server TLS trusted CA cert file.
	TrustedCaFile string `yaml:"trusted-ca-file,omitempty"`

	// Peer TLS using generated certificates
	AutoTls bool `yaml:"auto-tls,omitempty"`
}

//LoadEtcdConfig loads the default configuration for etcd
//TODO:support usage of external etcd
func (e *EtcdConfig) LoadEtcdConfig(ctx context.Context) *EtcdConfig {
	e.NewConfig(ctx)
	return e
}

//NewConfig set all the default required configurations needed
func (e *EtcdConfig) NewConfig(_ context.Context) {
	e.Name, _ = os.Hostname()

	e.DataDir = e.cfg.EtcdDataDir

	defaultIp := getDefaultIP()

	e.ListenPeerUrls = DefaultScheme + "://" + defaultIp + ":" + fmt.Sprint(DefaultPeerPort)

	e.ListenClientUrls = DefaultScheme + "://" + defaultIp + ":" + fmt.Sprint(DefaultClientPort) + "," + DefaultScheme + "://" + "127.0.0.1:" + fmt.Sprint(DefaultClientPort)

	e.InitialCluster = e.Name + "=" + e.ListenPeerUrls

	e.InitialAdvertisePeerUrls = DefaultScheme + "://" + defaultIp + ":" + fmt.Sprint(DefaultPeerPort)

	e.AdvertiseClientUrls = DefaultScheme + "://" + defaultIp + ":" + fmt.Sprint(DefaultClientPort)

	e.InitialClusterState = "new"

	e.InitialClusterToken = "etcd-cluster"

	e.ClientTransportSecurity.CertFile = filepath.Join(e.cfg.CertRootDir, constants.EtcdServerCertName)
	e.ClientTransportSecurity.KeyFile = filepath.Join(e.cfg.CertRootDir, constants.EtcdServerKeyName)
	e.ClientTransportSecurity.TrustedCaFile = filepath.Join(e.cfg.CertRootDir, constants.CACertName)
	e.ClientTransportSecurity.ClientCertAuth = true

	e.PeerTransportSecurity.CertFile = filepath.Join(e.cfg.CertRootDir, constants.EtcdPeerCertName)
	e.PeerTransportSecurity.KeyFile = filepath.Join(e.cfg.CertRootDir, constants.EtcdServerKeyName)
	e.PeerTransportSecurity.TrustedCaFile = filepath.Join(e.cfg.CertRootDir, constants.CACertName)
	e.PeerTransportSecurity.ClientCertAuth = true

}

//ToEmbedEtcdConfig pipes the required data in EtcdConfig to *embed.Config
func (e *EtcdConfig) ToEmbedEtcdConfig() *embed.Config {
	var embedConfig = embed.NewConfig()

	embedConfig.Name = e.Name

	embedConfig.Dir = e.DataDir
	embedConfig.WalDir = e.WalDir

	embedConfig.LPUrls = e.toUrl(e.ListenPeerUrls)

	embedConfig.LCUrls = e.toUrl(e.ListenClientUrls)

	embedConfig.InitialCluster = e.InitialCluster
	embedConfig.InitialClusterToken = e.InitialClusterToken

	embedConfig.APUrls = e.toUrl(e.InitialAdvertisePeerUrls)

	embedConfig.ACUrls = e.toUrl(e.AdvertiseClientUrls)

	embedConfig.ClusterState = e.InitialClusterState

	embedConfig.StrictReconfigCheck = false
	embedConfig.ClientAutoTLS = true
	embedConfig.ClientTLSInfo.CertFile = e.ClientTransportSecurity.CertFile
	embedConfig.ClientTLSInfo.KeyFile = e.ClientTransportSecurity.KeyFile
	embedConfig.ClientTLSInfo.ClientCertAuth = e.ClientTransportSecurity.ClientCertAuth
	embedConfig.ClientTLSInfo.TrustedCAFile = e.ClientTransportSecurity.TrustedCaFile

	embedConfig.PeerAutoTLS = true
	embedConfig.PeerTLSInfo.CertFile = e.PeerTransportSecurity.CertFile
	embedConfig.PeerTLSInfo.KeyFile = e.PeerTransportSecurity.KeyFile
	embedConfig.PeerTLSInfo.ClientCertAuth = e.PeerTransportSecurity.ClientCertAuth
	embedConfig.PeerTLSInfo.TrustedCAFile = e.PeerTransportSecurity.TrustedCaFile
	embedConfig.MaxSnapFiles = 0
	embedConfig.MaxWalFiles = 0
	embedConfig.Logger = "zap"
	if e.Debug {
		embedConfig.LogLevel = "debug"
	} else {
		embedConfig.LogLevel = "info"
	}

	embedConfig.LogOutputs = []string{"stderr"}
	return embedConfig
}

//toUrl is a utility to convert string to url
func (e *EtcdConfig) toUrl(commaSeparatedUrl string) []url.URL {
	lg, _ := zap.NewProduction()

	list := strings.Split(commaSeparatedUrl, ",")
	var urls []url.URL
	for _, item := range list {
		url, err := url.Parse(item)
		if err != nil {
			lg.Fatal("Unable to convert to URL", zap.String("urlInString", item))
		}
		urls = append(urls, *url)
	}
	return urls
}

//getDefaultIP is the utility which returns ip as a string
func getDefaultIP() string {
	ip, _ := GetDefaultIPV4()
	return ip
}
