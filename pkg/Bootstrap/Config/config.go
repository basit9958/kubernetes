package Config

import (
	"crypto/rand"
	"fmt"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	. "k8s.io/client-go/tools/clientcmd/api"
	"log"
	"os"
)

func DefaultConfig() *Config {
	return &Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Preferences:    *NewPreferences(),
		Clusters:       make(map[string]*Cluster),
		AuthInfos:      make(map[string]*AuthInfo),
		Contexts:       make(map[string]*Context),
		Extensions:     make(map[string]runtime.Object),
		CurrentContext: "",
	}
}

func Getconfigfiles() {
	getadminconfig()
	getcontrollermanagerconfig()
	getencryptionconfig()
}

func getadminconfig() {
	defaultConfig := DefaultConfig()
	defaultConfig.Clusters["chaos-mesh"] = &Cluster{
		Server:               "https://127.0.0.1:6443",
		CertificateAuthority: "/var/lib/kubernetes/ca.pem",
	}
	defaultConfig.AuthInfos["admin"] = &AuthInfo{
		ClientCertificate: "/var/lib/kubernetes/admin.pem",
		ClientKey:         "/var/lib/kubernetes/admin-key.pem",
	}

	defaultConfig.Contexts["default"] = &Context{
		Cluster:  "chaos-mesh",
		AuthInfo: "admin",
	}
	defaultConfig.CurrentContext = "default"

	output, err := yaml.Marshal(defaultConfig)
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}

	err = os.WriteFile("admin.kubeconfig", output, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func getcontrollermanagerconfig() {
	defaultConfig := DefaultConfig()
	defaultConfig.Clusters["chaos-mesh"] = &Cluster{
		Server:               "https://127.0.0.1:6443",
		CertificateAuthority: "/var/lib/kubernetes/ca.pem",
	}
	defaultConfig.AuthInfos["system:kube-controller-manager"] = &AuthInfo{
		ClientCertificate: "/var/lib/kubernetes/kube-controller-manager.pem",
		ClientKey:         "/var/lib/kubernetes/kube-controller-manager-key.pem",
	}

	defaultConfig.Contexts["default"] = &Context{
		Cluster:  "chaos-mesh",
		AuthInfo: "system:kube-controller-manager",
	}
	defaultConfig.CurrentContext = "default"

	output, err := yaml.Marshal(defaultConfig)
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}

	err = os.WriteFile("kube-controller-manager.kubeconfig", output, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func getencryptionconfig() {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}
	err = os.Setenv("ENCRYPTION_KEY", string(key))
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}

}
