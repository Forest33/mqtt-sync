package grpc

import (
	"maps"
	"sync"

	"github.com/forest33/mqtt-sync/business/entity"
	"github.com/forest33/mqtt-sync/pkg/logger"
)

const (
	initialQueueSize = 10
)

type queue struct {
	log      *logger.Logger
	messages map[string]entity.SyncMessage
	sync.Mutex
}

type stream interface {
	Send(m entity.SyncMessage) (err error)
}

func newQueue(log *logger.Logger) *queue {
	return &queue{
		log:      log,
		messages: make(map[string]entity.SyncMessage, initialQueueSize),
	}
}

func (q *queue) Push(message entity.SyncMessage) {
	q.Lock()
	q.messages[message.Topic()] = message
	q.Unlock()
}

func (q *queue) Pop(s stream) {
	q.Lock()
	messages := maps.Clone(q.messages)
	clear(q.messages)
	q.Unlock()

	q.log.Info().Int("size", len(messages)).Msg("sending saved messages from the queue")

	for k := range messages {
		if err := s.Send(messages[k]); err != nil {
			q.log.Error().Err(err).Msg("failed to send message")
		}
	}
}
