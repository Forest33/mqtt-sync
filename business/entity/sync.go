package entity

type SyncMessage interface {
	Topic() string
	Payload() []byte
	IsPayloadKey() bool
}

type SyncUseCase interface {
	OnMessage(topic string, payload []byte)
}
