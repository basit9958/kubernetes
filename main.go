package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"
	chaoscm "k8s.io/kubernetes/cmd/chaos-controller-manager"
	"k8s.io/kubernetes/cmd/kube-apiserver/app"
	ccm "k8s.io/kubernetes/cmd/kube-controller-manager/app"
	"k8s.io/kubernetes/pkg/bootstrap"
	"k8s.io/kubernetes/pkg/bootstrap/certificate"
	"k8s.io/kubernetes/pkg/bootstrap/dir"
	"k8s.io/kubernetes/pkg/bootstrap/etcd"
	"k8s.io/kubernetes/pkg/constants"
	"log"
	"os"
	"os/exec"
	"time"
)

var (
	Datadir string
)

func main() {
	ctx := context.Background()
	cfg := constants.GetConfig(Datadir)
	if err := preflight(cfg, ctx); err != nil {
		os.Exit(1)
	}

	if err := etcd.Run(ctx, cfg); err != nil {
		log.Fatal("server run failed")
		os.Exit(1)
	}
	cmd := app.NewAPIServerCommand(cfg)
	RunApiserver(cmd)
	//sleep some time to let kube-apiserver run
	ccmcmd := ccm.NewControllerManagerCommand(cfg)
	Runccm(ccmcmd)

	execmd := exec.Command("make", "install")
	execmd.Stdout = os.Stdout
	execmd.Stderr = os.Stderr
	err := execmd.Run()
	if err != nil {
		log.Fatalf("installation of CRD failed with %s\n", err)
	}

	chaoscm.Start()
}

//preflight generate all the required directories and certificates for the binary
func preflight(cfg constants.CfgVars, ctx context.Context) error {
	if err := dir.Init(cfg.DataDir, constants.DataDirMode); err != nil {
		return err
	}
	if err := dir.Init(cfg.CertRootDir, constants.CertRootDirMode); err != nil {
		return err
	}
	if err := dir.Init(cfg.EtcdDataDir, constants.EtcdDataDirMode); err != nil {
		return err
	}
	certificateManager := certificate.Manager{Cfg: cfg}
	certs := &bootstrap.Certificates{
		CertManager: certificateManager,
		CfgVars:     cfg,
	}
	if err := certs.Init(ctx); err != nil {
		fmt.Printf("", err)
		return err
	}
	return nil
}

//RunApiserver starts kube-apiserver
func RunApiserver(cmd *cobra.Command) {
	go func() {
		code := cli.Run(cmd)
		os.Exit(code)
	}()
	time.Sleep(10 * time.Second)
}

//Runccm starts kube-controller-manager
func Runccm(cmd *cobra.Command) {
	go func() {
		code := cli.Run(cmd)
		os.Exit(code)
	}()
	time.Sleep(15 * time.Second)
}
