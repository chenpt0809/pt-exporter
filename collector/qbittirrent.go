package collector

import (
	"github.com/chenpt0809/pt-exporter/client"
	"github.com/prometheus/client_golang/prometheus"
	"net/url"
	"sync"
)

type QbittorrentCollector struct {
	namespace                        string
	qbittorrentClient                *client.QbittorrentClient
	rewriteTracker                   map[string]string
	up                               prometheus.Gauge
	downloadBytesTotal               *prometheus.Desc
	uploadBytesTotal                 *prometheus.Desc
	downloadSpeedBytes               prometheus.Gauge
	uploadSpeedBytes                 prometheus.Gauge
	trackerTorrentDownloadBytesTotal *prometheus.Desc
	trackerTorrentUploadBytesTotal   *prometheus.Desc
	torrentsCount                    *prometheus.Desc
	mutex                            sync.Mutex
}

func NewQbittorrentCollector(name string, rewriteTracker map[string]string, c *client.QbittorrentClient) *QbittorrentCollector {
	ConstLabels := map[string]string{
		"name":   name,
		"host":   c.Address,
		"client": "qbittorrent",
	}
	return &QbittorrentCollector{
		namespace:         "downloader",
		qbittorrentClient: c,
		rewriteTracker:    rewriteTracker,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "downloader",
			Name:        "up",
			Help:        "Whether if server is alive or not",
			ConstLabels: ConstLabels,
		}),
		downloadBytesTotal: prometheus.NewDesc(
			"downloader_download_bytes_total",
			"downloadBytesTotal",
			nil,
			ConstLabels,
		),
		uploadBytesTotal: prometheus.NewDesc(
			"downloader_upload_bytes_total",
			"uploadBytesTotal",
			nil,
			ConstLabels,
		),
		downloadSpeedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "downloader",
			Name:        "download_speed_bytes",
			Help:        "download_speed_bytes",
			ConstLabels: ConstLabels,
		}),
		uploadSpeedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "downloader",
			Name:        "upload_speed_bytes",
			Help:        "upload_speed_bytes",
			ConstLabels: ConstLabels,
		}),
		trackerTorrentDownloadBytesTotal: prometheus.NewDesc(
			"downloader_tracker_torrent_download_bytes_total",
			"trackerTorrentDownloadBytesTotal",
			[]string{"torrent_name", "tracker"},
			ConstLabels,
		),
		trackerTorrentUploadBytesTotal: prometheus.NewDesc(
			"downloader_tracker_torrent_upload_bytes_total",
			"trackerTorrentUploadBytesTotal",
			[]string{"torrent_name", "tracker"},
			ConstLabels,
		),
		torrentsCount: prometheus.NewDesc(
			"downloader_torrents_count",
			"torrentsCount",
			[]string{"status", "tracker"},
			ConstLabels,
		),
	}
}

func (q *QbittorrentCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- q.up.Desc()
	descs <- q.uploadBytesTotal
	descs <- q.downloadBytesTotal
	descs <- q.trackerTorrentDownloadBytesTotal
	descs <- q.trackerTorrentUploadBytesTotal
	descs <- q.torrentsCount
}

func (q *QbittorrentCollector) Collect(metrics chan<- prometheus.Metric) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	status, err := q.qbittorrentClient.GetStatus()
	if err != nil {
		q.up.Set(0)
		metrics <- q.up
		return
	} else {
		q.up.Set(1)
		metrics <- q.up
	}
	torrents, err := q.qbittorrentClient.GetTorrent()
	if err != nil {
		return
	}
	metrics <- prometheus.MustNewConstMetric(
		q.downloadBytesTotal,
		prometheus.CounterValue,
		float64(status.Downloaded),
	)
	metrics <- prometheus.MustNewConstMetric(
		q.uploadBytesTotal,
		prometheus.CounterValue,
		float64(status.Uploaded),
	)
	q.downloadSpeedBytes.Set(float64(status.DownloadSpeed))
	metrics <- q.downloadSpeedBytes
	q.uploadSpeedBytes.Set(float64((status.UploadSpeed)))
	metrics <- q.uploadSpeedBytes
	state := make(map[string]map[string]int)
	for _, torrent := range torrents {
		trackerUrl, _ := url.Parse(torrent.Tracker)
		trackerAddress := trackerUrl.Hostname()
		trackerName, isok := q.rewriteTracker[trackerAddress]
		if !isok {
			trackerName = trackerAddress
		}
		metrics <- prometheus.MustNewConstMetric(
			q.trackerTorrentDownloadBytesTotal,
			prometheus.CounterValue,
			float64(torrent.Downloaded),
			torrent.Name,
			trackerName,
		)
		metrics <- prometheus.MustNewConstMetric(
			q.trackerTorrentUploadBytesTotal,
			prometheus.CounterValue,
			float64(torrent.Uploaded),
			torrent.Name,
			trackerName,
		)
		var stateName string
		switch torrent.State {
		case "unknown":
			stateName = "未知"
		case "allocating":
			stateName = "分配"
		case "downloading", "metaDL", "forcedDL":
			stateName = "下载中"
		case "uploading", "forcedUP":
			stateName = "上传中"
		case "checkingUP", "checkingDL", "checkingResumeData":
			stateName = "校验"
		case "missingFiles", "error":
			stateName = "错误"
		case "stalledUP", "stalledDL":
			stateName = "等待"
		case "queuedUP", "queuedDL":
			stateName = "排队"
		case "pausedUP", "pausedDL":
			stateName = "暂停"
		case "moving":
			stateName = "移动中"
		default:
			stateName = torrent.State

		}

		_, stateOk := state[stateName]
		if stateOk {
			_, trackerOk := state[stateName][trackerName]
			if trackerOk {
				state[stateName][trackerName]++
			} else {
				t := make(map[string]int)
				t[trackerName] = 1
				state[stateName][trackerName] = 1
			}
		} else {
			t := make(map[string]int)
			t[trackerName] = 1
			state[stateName] = t
		}
	}

	for status, v := range state {
		for tracker, vv := range v {
			metrics <- prometheus.MustNewConstMetric(
				q.torrentsCount,
				prometheus.GaugeValue,
				float64(vv),
				status,
				tracker,
			)
		}
	}

}
