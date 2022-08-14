package Config

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	. "k8s.io/client-go/tools/clientcmd/api/v1"
	"log"
	"os"
	"path/filepath"
)

func Getconfigfiles(cacert []byte, KubeCtrlManagercert []byte, KubeCtrlManagerkey []byte) {
	adminconfig := getadminconfig()
	output, err := json.Marshal(adminconfig)
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}
	err = os.WriteFile("/var/lib/kubernetes/admin.kubeconfig", output, 0644)
	if err != nil {
		log.Fatal(err)
	}
	kubecontrollermanagerconfig := getcontrollermanagerconfig(cacert, KubeCtrlManagercert, KubeCtrlManagerkey)
	kcmoutput, err := json.Marshal(kubecontrollermanagerconfig)
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}
	err = os.WriteFile("/var/lib/kubernetes/kube-controller-manager.kubeconfig", kcmoutput, 0644)
	if err != nil {
		log.Fatal(err)
	}
	getencryptionconfig()
}

func getadminconfig() *Config {
	clusterconfig := []NamedCluster{
		{Cluster: Cluster{Server: "https://127.0.0.1:6443", CertificateAuthority: "ca.pem"}, Name: "chaos-mesh"},
	}
	authinfo := []NamedAuthInfo{
		{Name: "admin", AuthInfo: AuthInfo{ClientCertificate: "admin.pem", ClientKey: "admin-key.pem"}},
	}
	contextinfo := []NamedContext{
		{Context: Context{Cluster: "chaos-mesh", AuthInfo: "admin"}, Name: "default"},
	}
	return &Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusterconfig,
		Contexts:       contextinfo,
		AuthInfos:      authinfo,
		CurrentContext: "default",
		Extensions:     nil,
	}
}
func getcontrollermanagerconfig(cacert []byte, KubeCtrlManagercert []byte, KubeCtrlManagerkey []byte) *Config {
	clusterconfig := []NamedCluster{
		{Cluster: Cluster{Server: "https://127.0.0.1:6443", CertificateAuthorityData: cacert}, Name: "chaos-mesh"},
	}
	authinfo := []NamedAuthInfo{
		{Name: "system:kube-controller-manager", AuthInfo: AuthInfo{ClientCertificateData: KubeCtrlManagercert, ClientKeyData: KubeCtrlManagerkey}},
	}
	contextinfo := []NamedContext{
		{Context: Context{Cluster: "chaos-mesh", AuthInfo: "system:kube-controller-manager"}, Name: "default"},
	}
	return &Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusterconfig,
		Contexts:       contextinfo,
		AuthInfos:      authinfo,
		CurrentContext: "default",
		Extensions:     nil,
	}
}

func getencryptionconfig() {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}
	os.Unsetenv("ENCRYTION_KEY")
	err = os.Setenv("ENCRYPTION_KEY", string(key))
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}
	dir := pkgpath() + "/pkg/Bootstrap/Config/"
	input, err := ioutil.ReadFile(dir + "encryption-config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("/var/lib/kubernetes/encryption-config.yaml", input, 0644)
	if err != nil {
		fmt.Println("Error creating", "/var/lib/kubernetes/encryption-config.yaml")
		fmt.Println(err)
		return
	}

	input, err = ioutil.ReadFile(dir + "auth.kubeconfig")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile("/var/lib/kubernetes/auth.kubeconfig", input, 0644)
	if err != nil {
		fmt.Println("Error creating", "/var/lib/kubernetes/auth.kubeconfig")
		fmt.Println(err)
		return
	}

	input, err = ioutil.ReadFile(dir + "token.kubeconfig")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile("/var/lib/kubernetes/token.kubeconfig", input, 0644)
	if err != nil {
		fmt.Println("Error creating", "/var/lib/kubernetes/token.kubeconfig.yaml")
		fmt.Println(err)
		return
	}

}
func pkgpath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}
