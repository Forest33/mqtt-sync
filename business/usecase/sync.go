package usecase

import (
	"context"

	"github.com/forest33/mqtt-sync/adapter/grpc"
	"github.com/forest33/mqtt-sync/business/entity"
	"github.com/forest33/mqtt-sync/pkg/logger"
)

type SyncUseCase struct {
	ctx  context.Context
	cfg  *entity.Config
	log  *logger.Logger
	mqtt MqttClient
	srv  *grpc.Server
	cli  *grpc.Client
}

func NewSyncUseCase(ctx context.Context, cfg *entity.Config, log *logger.Logger, mqtt MqttClient, srv *grpc.Server, cli *grpc.Client) (*SyncUseCase, error) {
	uc := &SyncUseCase{
		ctx:  ctx,
		cfg:  cfg,
		log:  log,
		mqtt: mqtt,
		srv:  srv,
		cli:  cli,
	}

	if uc.srv != nil {
		uc.srv.SetSyncUseCase(uc)
		uc.srv.Start()
	}

	if uc.cli != nil {
		uc.cli.SetSyncUseCase(uc)
		if err := uc.cli.Start(); err != nil {
			return nil, err // TODO вернуть!
		}
	}

	uc.mqtt.SetConnectHandler(uc.OnConnect)
	if err := uc.mqtt.Connect(); err != nil {
		return nil, err
	}

	return uc, nil
}

func (uc *SyncUseCase) OnConnect() {
	for _, t := range uc.cfg.Sync.Topics {
		if err := uc.mqtt.Subscribe(t, uc.mqttMessage); err != nil {
			uc.log.Fatalf("failed to subscribe to topic %s: %v", t, err)
		}
		uc.log.Info().Str("topic", t).Msg("subscribed to topic")
	}
}

func (uc *SyncUseCase) OnMessage(topic string, payload []byte) {
	uc.log.Debug().Str("topic", topic).Str("payload", string(payload)).Msg("peer message")
	if err := uc.mqtt.Publish(topic, payload); err != nil {
		uc.log.Error().Err(err).Msg("failed to publish message")
	}
}

func (uc *SyncUseCase) mqttMessage(m entity.SyncMessage) {
	if m.IsPayloadKey() {
		return
	}

	uc.log.Debug().Str("topic", m.Topic()).Str("payload", string(m.Payload())).Msg("MQTT message")

	var err error
	switch {
	case uc.srv != nil:
		err = uc.srv.Send(m)
	case uc.cli != nil:
		err = uc.cli.Send(m)
	}
	if err != nil {
		uc.log.Error().Err(err).Msg("failed to send message")
	}
}
