package collector

import (
	"github.com/chenpt0809/pt-exporter/client"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type TransmissionCollector struct {
	clientName                       string
	transmissionClient               *client.TransmissionClient
	Options                          Options
	up                               prometheus.Gauge
	downloadBytesTotal               *prometheus.Desc
	uploadBytesTotal                 *prometheus.Desc
	downloadSpeedBytes               prometheus.Gauge
	uploadSpeedBytes                 prometheus.Gauge
	freeSpaceOnDisk                  prometheus.Gauge
	trackerTorrent                   *prometheus.Desc
	trackerTorrentSizeBytes          *prometheus.Desc
	trackerTorrentStatus             *prometheus.Desc
	trackerTorrentDownloadBytesTotal *prometheus.Desc
	trackerTorrentUploadBytesTotal   *prometheus.Desc
	torrentsCount                    *prometheus.Desc
	maxDownloadSpeedBytes            prometheus.Gauge
	maxUploadSpeedBytes              prometheus.Gauge

	mutex sync.Mutex
}
