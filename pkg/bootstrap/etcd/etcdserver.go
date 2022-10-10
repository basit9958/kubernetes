package etcd

import (
	"context"
	"k8s.io/kubernetes/pkg/constants"
	"time"

	"go.etcd.io/etcd/server/v3/embed"

	"go.uber.org/zap"
	"log"
)

type ETCD struct {
	Config *EtcdConfig
	Logger *zap.Logger
	ctx    context.Context
}

func Run(ctx context.Context, cfg constants.CfgVars) error {

	EtcdConfig := EtcdConfig{cfg: cfg}
	etcdConfig := EtcdConfig.LoadEtcdConfig(ctx)
	etcdConfig.cfg = cfg
	e, err := NewEtcd(etcdConfig)
	if err != nil {
		return err
	}
	RunEtcd(ctx, etcdConfig, e)
	return nil
}

func RunEtcd(ctx context.Context, etcdConfig *EtcdConfig, e *ETCD) {
	go e.StartEtcd(ctx, etcdConfig)
	time.Sleep(5 * time.Second)
}

func NewEtcd(config *EtcdConfig) (*ETCD, error) {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &ETCD{
		Config: config,
		Logger: lg,
	}, nil
}

func (d *ETCD) StartEtcd(ctx context.Context, config *EtcdConfig) {
	e, err := embed.StartEtcd(config.ToEmbedEtcdConfig())
	if err != nil {
		d.Logger.Warn("Unable to start etcd.", zap.Error(err))
	}
	defer e.Close()
	select {
	case <-e.Server.ReadyNotify():
		d.Logger.Info("etcd Server is ready!")
	case <-time.After(180 * time.Second):
		d.Logger.Warn("etcd didn't start on time.")
	}
	etcdErr := <-e.Err()

	d.Logger.Error("ETCD fatal mishap", zap.Error(etcdErr))
}
