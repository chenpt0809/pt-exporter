package client

import (
	"github.com/chenpt0809/pt-exporter/global"
	"github.com/hekmon/transmissionrpc/v2"
)

type TransmissionClient struct {
	client   *transmissionrpc.Client
	Address  string
	Username string
	Password string
	baseURL  string
	sid      string
	IsLogin  bool
}

type TransmissionOptions struct {
	Host           string
	Port           int
	UserName       string
	Password       string
	RequestTimeOut int
}

func NewTransmissionClient(Options TransmissionOptions) *TransmissionClient {
	c := &TransmissionClient{
		Address:  Options.Host,
		Username: Options.UserName,
		Password: Options.Password,
	}
	global.Logger.Debug("创建：TransmissionClient")
	tc, err := transmissionrpc.New(Options.Host, Options.UserName, Options.Password, &transmissionrpc.AdvancedConfig{Port: uint16(Options.Port)})
	if err != nil {
		global.Logger.Error("创建：TransmissionClient失败")
	}
	c.client = tc
	return c
}
