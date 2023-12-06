package client

import (
	"github.com/chenpt0809/pt-exporter/global"
	"github.com/chenpt0809/pt-exporter/utils"
	"github.com/hekmon/transmissionrpc/v2"
)

type TransmissionClient struct {
	Client   *transmissionrpc.Client
	Host     string
	Port     int
	Address  string
	UserName string
	Password string
	baseURL  string
	sid      string
	IsLogin  bool
}

type TransmissionOptions struct {
	Url            string
	UserName       string
	Password       string
	RequestTimeOut int
}

func NewTransmissionClient(Options TransmissionOptions) *TransmissionClient {
	host, port, err := utils.GetHostAndPort(Options.Url)
	if err != nil {
		global.Logger.Error("无法解析的URL:" + Options.Url)
		return nil
	}
	if port == 0 {
		port = 80
	}
	c := &TransmissionClient{
		Host:     host,
		Port:     port,
		Address:  Options.Url,
		UserName: Options.UserName,
		Password: Options.Password,
	}
	global.Logger.Debug("创建：TransmissionClient")
	tc, err := transmissionrpc.New(c.Host, c.UserName, c.Password, &transmissionrpc.AdvancedConfig{Port: uint16(c.Port)})
	if err != nil {
		global.Logger.Error("创建：TransmissionClient失败")
	}
	c.Client = tc
	return c
}
