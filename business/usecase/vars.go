package usecase

import (
	"github.com/forest33/mqtt-sync/adapter/mqtt"
	"github.com/forest33/mqtt-sync/business/entity"
)

type MqttClient interface {
	Connect() error
	Publish(topic string, payload []byte) error
	Subscribe(topic string, handler mqtt.MessageHandler) error
	SetConnectHandler(h mqtt.ConnectHandler)
	SetDisconnectHandler(h mqtt.DisconnectHandler)
	Close()
}

type GrpcServer interface {
	Start()
	Send(m *entity.SyncMessage) error
	SetSyncUseCase(uc entity.SyncUseCase)
}

type GrpcClient interface {
	Start() error
	Send(m *entity.SyncMessage) error
	SetSyncUseCase(uc entity.SyncUseCase)
}
