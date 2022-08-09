package kubernetes

import (
	"fmt"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	apiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
	controllermanager "k8s.io/kubernetes/cmd/kube-controller-manager/app"
	controllermanageropts "k8s.io/kubernetes/cmd/kube-controller-manager/app/options"
	. "k8s.io/kubernetes/pkg/Bootstrap/Args-kubernetes"
	"k8s.io/kubernetes/pkg/Bootstrap/Config"
	"k8s.io/kubernetes/pkg/Bootstrap/cert"
	"time"
)

func main() {

	cert.Generatecert()
	Config.Getconfigfiles()
	s := options.NewServerRunOptions()
	fs := pflag.NewFlagSet("addflagstest", pflag.ContinueOnError)
	for _, f := range s.Flags().FlagSets {
		fs.AddFlagSet(f)
	}
	// silence client-go warnings.
	// kube-apiserver loopback clients should not log self-issued warnings.
	rest.SetDefaultWarningHandler(rest.NoWarnings{})
	apiserverargs := Apiserverflags()
	fs.Parse(apiserverargs)
	completedOptions, err := apiserver.Complete(s)
	if err != nil {
		fmt.Println(err)
	}
	if errs := completedOptions.Validate(); len(errs) != 0 {
		utilerrors.NewAggregate(errs)
	}
	apiserver.Run(completedOptions, genericapiserver.SetupSignalHandler())
	// silence client-go warnings.
	//sleep some time to let kube-apiserver run
	time.Sleep(10 * time.Second)

	restclient.SetDefaultWarningHandler(restclient.NoWarnings{})
	c, err := controllermanageropts.NewKubeControllerManagerOptions()
	if err != nil {
		klog.Fatalf("unable to initialize command options: %v", err)
	}
	fc := pflag.NewFlagSet("addflagstest", pflag.ContinueOnError)
	for _, f := range c.Flags([]string{""}, []string{""}).FlagSets {
		fc.AddFlagSet(f)
	}
	controllermanagerargs := Kubecontrollermanagerflags()
	fs.Parse(controllermanagerargs)
	cm, err := c.Config(controllermanager.KnownControllers(), controllermanager.ControllersDisabledByDefault.List())
	if err != nil {
		fmt.Println(err)
	}

	go controllermanager.Run(cm.Complete(), wait.NeverStop)
}
