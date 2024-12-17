// Package entity provides entities for business logic.
package entity

import (
	"github.com/forest33/mqtt-sync/pkg/config"
)

const (
	DefaultConfigFileName = "mqtt-sync.yaml"
)

type Config struct {
	Server  *Server  `yaml:"Server"`
	Client  *Client  `yaml:"Client"`
	MQTT    *MQTT    `yaml:"MQTT"`
	Sync    *Sync    `yaml:"Sync"`
	Logger  *Logger  `yaml:"Logger"`
	Runtime *Runtime `yaml:"Runtime"`
}

type Server struct {
	Enabled   bool       `yaml:"Enabled" default:"false"`
	Host      string     `yaml:"Host" default:""`
	Port      int        `yaml:"Port" default:"31883"`
	UseTLS    bool       `yaml:"UseTLS"  default:"false"`
	CACert    string     `yaml:"CACert"  default:""`
	Cert      string     `yaml:"Cert"  default:""`
	Key       string     `yaml:"Key" default:""`
	Keepalive *Keepalive `yaml:"Keepalive"`
}

type Client struct {
	Enabled              bool       `yaml:"Enabled" default:"false"`
	Host                 string     `yaml:"Host" default:"127.0.0.1"`
	Port                 int        `yaml:"Port" default:"31883"`
	UseTLS               bool       `yaml:"UseTLS"  default:"false"`
	CACert               string     `yaml:"CACert"  default:""`
	Cert                 string     `yaml:"Cert"  default:""`
	Key                  string     `yaml:"Key" default:""`
	InsecureSkipVerify   bool       `yaml:"InsecureSkipVerify"  default:"true"`
	ConnectRetryInterval int        `yaml:"ConnectRetryInterval" default:"3"`
	Keepalive            *Keepalive `yaml:"Keepalive"`
}

type Keepalive struct {
	PingMinTime         int  `yaml:"KeepalivePingMinTime" default:"30"`
	Time                int  `yaml:"KeepaliveTime" default:"30"`
	Timeout             int  `yaml:"KeepaliveTimeout" default:"10"`
	PermitWithoutStream bool `yaml:"KeepalivePermitWithoutStream" default:"true"`
}

type MQTT struct {
	Host                 string `yaml:"Host" default:"127.0.0.1"`
	Port                 int    `yaml:"Port" default:"1883"`
	ClientID             string `yaml:"ClientID" default:"mqtt-sync"`
	User                 string `yaml:"User" default:""`
	Password             string `yaml:"Password" default:""`
	UseTLS               bool   `yaml:"UseTLS"  default:"false"`
	ServerTLS            bool   `yaml:"ServerTLS"  default:"false"`
	CACert               string `yaml:"CACert"  default:""`
	Cert                 string `yaml:"Cert"  default:""`
	Key                  string `yaml:"Key" default:""`
	ConnectRetryInterval int    `yaml:"ConnectRetryInterval" default:"3"`
	Timeout              int    `yaml:"Timeout" default:"10"`
}

type Sync struct {
	Topics     []string `yaml:"Topics"`
	PayloadKey string   `yaml:"PayloadKey" default:"___mqtt_sync___"`
}

type Logger struct {
	Level             string `yaml:"Level" default:"debug"`
	TimeFormat        string `yaml:"TimeFormat" default:"2006-01-02T15:04:05.000000"`
	PrettyPrint       bool   `yaml:"PrettyPrint" default:"false"`
	DisableSampling   bool   `yaml:"DisableSampling" default:"true"`
	RedirectStdLogger bool   `yaml:"RedirectStdLogger" default:"true"`
	ErrorStack        bool   `yaml:"ErrorStack" default:"true"`
}

type Runtime struct {
	GoMaxProcs int `yaml:"GoMaxProcs" default:"0"`
}

type ConfigHandler interface {
	Update(data interface{})
	Save()
	GetPath() string
	AddObserver(f func(interface{})) error
}

func GetConfig(configDir string) (ConfigHandler, *Config, error) {
	cfg := &Config{}
	h, err := config.New(DefaultConfigFileName, configDir, cfg)
	if err != nil {
		return nil, nil, err
	}
	return h, cfg, nil
}
