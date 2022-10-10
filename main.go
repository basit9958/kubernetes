package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"
	"k8s.io/kubernetes/cmd/kube-apiserver/app"
	ccm "k8s.io/kubernetes/cmd/kube-controller-manager/app"
	"k8s.io/kubernetes/pkg/bootstrap"
	"k8s.io/kubernetes/pkg/bootstrap/certificate"
	"k8s.io/kubernetes/pkg/bootstrap/dir"
	"k8s.io/kubernetes/pkg/bootstrap/etcd"
	"k8s.io/kubernetes/pkg/constants"
	"log"
	"os"
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
	RunApiserverinbackground(cmd)
	//sleep some time to let kube-apiserver run
	ccmcmd := ccm.NewControllerManagerCommand(cfg)
	code := cli.Run(ccmcmd)
	os.Exit(code)
}

//preflight make all the required directories and certificates for the binary
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

//RunApiserverinbackground runs NewAPIServerCommand as a goroutine
func RunApiserverinbackground(cmd *cobra.Command) {
	go cli.Run(cmd)
	time.Sleep(10 * time.Second)
}
