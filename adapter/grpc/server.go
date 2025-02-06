package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	apiV1 "github.com/forest33/mqtt-sync/api/v1"
	"github.com/forest33/mqtt-sync/business/entity"
	"github.com/forest33/mqtt-sync/pkg/logger"
)

type Server struct {
	ctx    context.Context
	cfg    *Config
	log    *logger.Logger
	queue  *queue
	lst    net.Listener
	srv    *grpc.Server
	uc     entity.SyncUseCase
	stream apiV1.MqttSync_SyncServer
}

func NewServer(ctx context.Context, cfg *Config, log *logger.Logger) (*Server, error) {
	s := &Server{
		ctx:   ctx,
		cfg:   cfg,
		log:   log,
		queue: newQueue(log),
	}

	var err error
	s.lst, err = net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		return nil, err
	}

	opts := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             time.Duration(cfg.KeepalivePingMinTime) * time.Second,
			PermitWithoutStream: false,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    time.Duration(cfg.KeepaliveTime) * time.Second,
			Timeout: time.Duration(cfg.KeepaliveTimeout) * time.Second,
		}),
	}

	if cfg.UseTLS {
		tlsCredentials, err := loadTLSCredentials(cfg)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(tlsCredentials))
	}

	s.srv = grpc.NewServer(opts...)
	apiV1.RegisterMqttSyncServer(s.srv, s)

	entity.GetWg(ctx).Add(1)
	go func() {
		<-ctx.Done()
		s.srv.GracefulStop()
		s.log.Info().Msg("gRPC server stopped")
		entity.GetWg(ctx).Done()
	}()

	return s, nil
}

func (s *Server) SetSyncUseCase(uc entity.SyncUseCase) {
	s.uc = uc
}

func (s *Server) Start() {
	s.log.Info().
		Bool("tls", s.cfg.UseTLS).
		Str("host", s.cfg.Host).
		Int("port", s.cfg.Port).
		Msg("gRPC server started")
	go func() {
		if err := s.srv.Serve(s.lst); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

func (s *Server) Sync(stream apiV1.MqttSync_SyncServer) error {
	var (
		ctx = stream.Context()
	)

	s.stream = stream

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-ctx.Done():
			s.log.Debug().Str("reason", ctx.Err().Error()).Msg("stream closed")
			return ctx.Err()
		default:
			req, err := stream.Recv()
			if err != nil {
				if status.Code(err) != codes.Canceled && err != io.EOF {
					s.log.Error().Err(err).Msg("stream broken")
				}
				return err
			}

			if s.uc == nil || len(req.Topic) == 0 {
				s.queue.Pop(s)
				md, _ := metadata.FromIncomingContext(ctx)
				s.log.Debug().Interface("peer", md).Msg("peer connected")
				continue
			}

			s.uc.OnMessage(req.Topic, req.Payload)
		}
	}
}

func (s *Server) Send(m entity.SyncMessage) (err error) {
	err = s.send(m)
	if err != nil {
		s.queue.Push(m)
	}
	return
}

func (s *Server) send(m entity.SyncMessage) error {
	if s.stream == nil {
		return entity.ErrStreamDisabled
	}

	return s.stream.SendMsg(&apiV1.Message{
		Topic:   m.Topic(),
		Payload: m.Payload(),
	})
}
