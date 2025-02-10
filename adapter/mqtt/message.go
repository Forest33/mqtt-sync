package mqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/forest33/mqtt-sync/pkg/codec"
)

type message struct {
	topic      string
	payload    []byte
	codec      codec.Codec
	data       map[string]interface{}
	payloadKey bool
}

func (m *message) Topic() string {
	return m.topic
}

func (m *message) Payload() []byte {
	return m.payload
}

func (m *message) IsPayloadKey() bool {
	return m.payloadKey
}

func (c *Client) newMessage(msg mqtt.Message) (*message, error) {
	var (
		data map[string]interface{}
		err  error
	)

	payload := msg.Payload()

	if err := c.codec.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	_, payloadKey := data[c.cfg.PayloadKey]
	if !payloadKey {
		data[c.cfg.PayloadKey] = 1
		payload, err = c.codec.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	return &message{
		topic:      msg.Topic(),
		payload:    payload,
		codec:      c.codec,
		data:       data,
		payloadKey: payloadKey,
	}, nil
}
