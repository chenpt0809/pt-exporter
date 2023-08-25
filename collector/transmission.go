package collector

import (
	"github.com/chenpt0809/pt-exporter/client"
	"sync"
)

type TransmissionCollector struct {
	clientName         string
	transmissionClient *client.TransmissionClient
	Options            Options

	mutex sync.Mutex
}

func NewTransmissionCollector(name string, c *client.TransmissionClient, o Options) *TransmissionCollector {
	return &TransmissionCollector{}
}
