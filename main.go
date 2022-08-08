package kubernetes

import (
	"fmt"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	genericapiserver "k8s.io/apiserver/pkg/server"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	apiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
	controllermanager "k8s.io/kubernetes/cmd/kube-controller-manager/app"
	controllermanageropts "k8s.io/kubernetes/cmd/kube-controller-manager/app/options"
	"k8s.io/kubernetes/pkg/cert"
	. "k8s.io/kubernetes/pkg/k8s-flags"
	"time"
)

func main() {
	//Generates all the required certificates
	cert.Generatecert()
	s := options.NewServerRunOptions()
	fs := pflag.NewFlagSet("addflagstest", pflag.ContinueOnError)
	for _, f := range s.Flags().FlagSets {
		fs.AddFlagSet(f)
	}
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
