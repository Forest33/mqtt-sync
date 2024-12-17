package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/forest33/mqtt-sync/adapter/grpc"
	"github.com/forest33/mqtt-sync/adapter/mqtt"
	"github.com/forest33/mqtt-sync/business/entity"
	"github.com/forest33/mqtt-sync/business/usecase"
	"github.com/forest33/mqtt-sync/pkg/automaxprocs"
	"github.com/forest33/mqtt-sync/pkg/build"
	"github.com/forest33/mqtt-sync/pkg/codec"
	"github.com/forest33/mqtt-sync/pkg/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-ctx.Done()
		cancel()
	}()

	ctx = entity.CreateWg(ctx)

	_, cfg, err := entity.GetConfig(build.EnvPrefix)
	if err != nil {
		log.Fatal(err)
	}

	l := logger.New(logger.Config{
		Level:             cfg.Logger.Level,
		TimeFormat:        cfg.Logger.TimeFormat,
		PrettyPrint:       cfg.Logger.PrettyPrint,
		RedirectStdLogger: cfg.Logger.RedirectStdLogger,
		DisableSampling:   cfg.Logger.DisableSampling,
		ErrorStack:        cfg.Logger.ErrorStack,
	})

	if err := automaxprocs.Init(cfg, l); err != nil {
		l.Fatal(err)
	}

	mqttClient, err := mqtt.New(ctx, &mqtt.Config{
		Host:                 cfg.MQTT.Host,
		Port:                 cfg.MQTT.Port,
		ClientID:             cfg.MQTT.ClientID,
		User:                 cfg.MQTT.User,
		Password:             cfg.MQTT.Password,
		UseTLS:               cfg.MQTT.UseTLS,
		ServerTLS:            cfg.MQTT.ServerTLS,
		CACert:               cfg.MQTT.CACert,
		Cert:                 cfg.MQTT.Cert,
		Key:                  cfg.MQTT.Key,
		InsecureSkipVerify:   false,
		ConnectRetryInterval: time.Duration(cfg.MQTT.ConnectRetryInterval) * time.Second,
		Timeout:              time.Duration(cfg.MQTT.Timeout) * time.Second,
		PayloadKey:           cfg.Sync.PayloadKey,
	}, l, codec.NewFastJsonCodec())
	if err != nil {
		l.Fatal(err)
	}

	var (
		srv *grpc.Server
		cli *grpc.Client
	)

	if cfg.Server.Enabled {
		srv, err = grpc.NewServer(ctx, &grpc.Config{
			Host:                         cfg.Server.Host,
			Port:                         cfg.Server.Port,
			UseTLS:                       cfg.Server.UseTLS,
			CACert:                       cfg.Server.CACert,
			Cert:                         cfg.Server.Cert,
			Key:                          cfg.Server.Key,
			KeepalivePingMinTime:         cfg.Server.Keepalive.PingMinTime,
			KeepaliveTime:                cfg.Server.Keepalive.Time,
			KeepaliveTimeout:             cfg.Server.Keepalive.Timeout,
			KeepalivePermitWithoutStream: cfg.Server.Keepalive.PermitWithoutStream,
		}, l)
		if err != nil {
			l.Fatal(err)
		}
	}

	if cfg.Client.Enabled {
		cli, err = grpc.NewClient(ctx, &grpc.Config{
			Host:                         cfg.Client.Host,
			Port:                         cfg.Client.Port,
			UseTLS:                       cfg.Client.UseTLS,
			CACert:                       cfg.Client.CACert,
			Cert:                         cfg.Client.Cert,
			Key:                          cfg.Client.Key,
			InsecureSkipVerify:           cfg.Client.InsecureSkipVerify,
			ConnectRetryInterval:         time.Duration(cfg.Client.ConnectRetryInterval) * time.Second,
			KeepaliveTime:                cfg.Client.Keepalive.Time,
			KeepaliveTimeout:             cfg.Client.Keepalive.Timeout,
			KeepalivePermitWithoutStream: cfg.Client.Keepalive.PermitWithoutStream,
		}, l)
		if err != nil {
			l.Fatal(err)
		}
	}

	_, err = usecase.NewSyncUseCase(ctx, cfg, l, mqttClient, srv, cli)
	if err != nil {
		l.Fatal(err)
	}

	entity.GetWg(ctx).Wait()
}
