package collector

import (
	"context"
	"github.com/chenpt0809/pt-exporter/client"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type TransmissionCollector struct {
	clientName         string
	Options            Options
	Coll               *Collector
	transmissionClient *client.TransmissionClient
	mutex              sync.Mutex
}

func NewTransmissionCollector(name string, c *client.TransmissionClient, o Options) *TransmissionCollector {
	Coll := NewCollector(name, c.Address, "Transmission", o)
	return &TransmissionCollector{
		clientName:         name,
		Options:            o,
		Coll:               Coll,
		transmissionClient: c,
	}
}

func (t *TransmissionCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- t.Coll.up.Desc()
	descs <- t.Coll.uploadBytesTotal
	descs <- t.Coll.downloadBytesTotal
	descs <- t.Coll.freeSpaceOnDisk.Desc()
	descs <- t.Coll.torrent
	descs <- t.Coll.torrentStatus
	descs <- t.Coll.torrentSizeBytes
	descs <- t.Coll.torrentDownloadBytesTotal
	descs <- t.Coll.torrentUploadBytesTotal
	if !t.Options.DownloaderExporter {
		if t.Options.MaxDownSpeed != 0 {
			descs <- t.Coll.maxDownloadSpeedBytes.Desc()
		}
		if t.Options.MaxUpSpeed != 0 {
			descs <- t.Coll.maxUploadSpeedBytes.Desc()
		}
	}

}

func (t *TransmissionCollector) Collect(metrics chan<- prometheus.Metric) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if !t.Options.DownloaderExporter {
		if t.Options.MaxDownSpeed != 0 {
			metrics <- t.Coll.maxDownloadSpeedBytes
		}
		if t.Options.MaxUpSpeed != 0 {
			metrics <- t.Coll.maxUploadSpeedBytes
		}
	}
	status, err := t.transmissionClient.Client.SessionStats(context.TODO())
	if err != nil {
		t.Coll.up.Set(0)
		metrics <- t.Coll.up
		return
	}
	torrents, err := t.transmissionClient.Client.TorrentGetAll(context.TODO())
	if err != nil {
		t.Coll.up.Set(0)
		metrics <- t.Coll.up
		return
	}
	downloadDir, err := t.transmissionClient.Client.SessionArgumentsGet(context.TODO(), []string{"download-dir"})
	if err != nil {
		t.Coll.up.Set(0)
		metrics <- t.Coll.up
		return
	}
	freeSpace, _ := t.transmissionClient.Client.FreeSpace(context.TODO(), *downloadDir.DownloadDir)
	metrics <- prometheus.MustNewConstMetric(
		t.Coll.downloadBytesTotal,
		prometheus.CounterValue,
		float64(status.CumulativeStats.DownloadedBytes),
	)
	metrics <- prometheus.MustNewConstMetric(
		t.Coll.uploadBytesTotal,
		prometheus.CounterValue,
		float64(status.CumulativeStats.UploadedBytes),
	)
	t.Coll.downloadSpeedBytes.Set(float64(status.DownloadSpeed))
	metrics <- t.Coll.downloadSpeedBytes
	t.Coll.uploadSpeedBytes.Set(float64(status.UploadSpeed))
	metrics <- t.Coll.uploadSpeedBytes
	t.Coll.freeSpaceOnDisk.Set(freeSpace.Byte())
	metrics <- t.Coll.freeSpaceOnDisk
	t.Coll.up.Set(1)
	metrics <- t.Coll.up

}
