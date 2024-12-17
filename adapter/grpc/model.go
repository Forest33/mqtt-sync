package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc/credentials"
)

type Config struct {
	Host                         string
	Port                         int
	UseTLS                       bool
	CACert                       string
	Cert                         string
	Key                          string
	InsecureSkipVerify           bool
	ConnectRetryInterval         time.Duration
	KeepalivePingMinTime         int
	KeepaliveTime                int
	KeepaliveTimeout             int
	KeepalivePermitWithoutStream bool
}

func loadTLSCredentials(cfg *Config) (credentials.TransportCredentials, error) {
	ca, err := os.ReadFile(cfg.CACert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	cert, err := os.ReadFile(cfg.Cert)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	key, err := os.ReadFile(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate key: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	serverCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{serverCert},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		ClientCAs:          certPool,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	return credentials.NewTLS(tlsConfig), nil
}
