package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	apiV1 "github.com/forest33/mqtt-sync/api/v1"
	"github.com/forest33/mqtt-sync/business/entity"
	"github.com/forest33/mqtt-sync/pkg/logger"
)

type Client struct {
	ctx    context.Context
	cfg    *Config
	log    *logger.Logger
	cli    apiV1.MqttSyncClient
	stream apiV1.MqttSync_SyncClient
	uc     entity.SyncUseCase
}

func NewClient(ctx context.Context, cfg *Config, log *logger.Logger) (*Client, error) {
	c := &Client{
		ctx: ctx,
		cfg: cfg,
		log: log,
	}

	serverAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Duration(cfg.KeepaliveTime) * time.Second,
			Timeout:             time.Duration(cfg.KeepaliveTimeout) * time.Second,
			PermitWithoutStream: cfg.KeepalivePermitWithoutStream,
		}),
	}

	opts = append(opts)

	if cfg.UseTLS {
		tlsCredentials, err := loadTLSCredentials(cfg)
		if err != nil {
			return nil, errors.New("failed to load TLS credentials")
		}
		opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))
		log.Info().Msg("client TLS enabled")
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return nil, err
	}

	log.Info().Str("address", serverAddr).Msg("gRPC client connected")

	c.cli = apiV1.NewMqttSyncClient(conn)

	entity.GetWg(ctx).Add(1)
	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			c.log.Error().Err(err).Msg("failed to close gRPC client connection")
		}
		log.Info().Msg("gRPC client disconnected")
		entity.GetWg(ctx).Done()
	}()

	return c, nil
}

func (c *Client) SetSyncUseCase(uc entity.SyncUseCase) {
	c.uc = uc
}

func (c *Client) Start() error {
	var err error
	defer func() {
		if err != nil {
			c.reconnect()
		}
	}()

	c.stream, err = c.cli.Sync(c.ctx)
	if err != nil {
		return err
	}

	if err := c.stream.Send(&apiV1.Message{}); err != nil {
		c.log.Error().Err(err).Msg("failed to send init message")
		return err
	}

	c.log.Info().
		Bool("tls", c.cfg.UseTLS).
		Str("host", c.cfg.Host).
		Int("port", c.cfg.Port).
		Msg("successfully connected to gRPC server")

	go func() {
		var (
			err error
			req *apiV1.Message
		)

		defer func() {
			if err != nil {
				c.reconnect()
			}
		}()

		for {
			req, err = c.stream.Recv()
			if err != nil {
				c.log.Info().Str("reason", err.Error()).Msg("client stream broken")
				return
			}

			if c.uc == nil {
				continue
			}

			c.uc.OnMessage(req.Topic, req.Payload)
		}
	}()

	return nil
}

func (c *Client) Send(m entity.SyncMessage) error {
	if c.stream == nil {
		return nil
	}

	return c.stream.Send(&apiV1.Message{
		Topic:   m.Topic(),
		Payload: m.Payload(),
	})
}

func (c *Client) reconnect() {
	go func() {
		c.log.Info().
			Bool("tls", c.cfg.UseTLS).
			Str("host", c.cfg.Host).
			Int("port", c.cfg.Port).
			Msgf("gRPC client disconnected, retrying in %d seconds...", int(c.cfg.ConnectRetryInterval.Seconds()))
		time.Sleep(c.cfg.ConnectRetryInterval)
		_ = c.Start()
	}()
}
