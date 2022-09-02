package collector

import (
	"github.com/chenpt0809/pt-exporter/client"
	"github.com/prometheus/client_golang/prometheus"
	"net/url"
	"sync"
)

type QbittorrentCollector struct {
	clientName                       string
	qbittorrentClient                *client.QbittorrentClient
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

func NewQbittorrentCollector(name string, c *client.QbittorrentClient, o Options) *QbittorrentCollector {
	// 标签设置
	ConstLabels := map[string]string{
		"name":   name,
		"host":   c.Address,
		"client": "qbittorrent",
	}
	namespace := "pt"
	if o.DownloaderExporter {
		namespace = "downloader"
		// 添加版本标签 兼容Downloader_exporter
		ConstLabels["version"] = "v0.0.0"
	}
	// 创建Collector
	qbColl := QbittorrentCollector{
		clientName:        name,
		qbittorrentClient: c,
		Options:           o,
	}
	// 是否在线
	qbColl.up = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "up",
		Help:        "是否启动",
		ConstLabels: ConstLabels,
	})
	// 总下载量
	qbColl.downloadBytesTotal = prometheus.NewDesc(
		namespace+"_download_bytes_total",
		"总下载 单位字节",
		nil,
		ConstLabels,
	)
	qbColl.uploadBytesTotal = prometheus.NewDesc(
		namespace+"_upload_bytes_total",
		"总上传 单位字节",
		nil,
		ConstLabels,
	)
	// 剩余磁盘空间
	qbColl.freeSpaceOnDisk = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "free_space_on_disk_bytes",
		Help:        "默认磁盘剩余空间 单位字节",
		ConstLabels: ConstLabels,
	})
	// 当前下载速度
	qbColl.downloadSpeedBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "download_speed_bytes",
		Help:        "当前下载速度 单位字节",
		ConstLabels: ConstLabels,
	})
	// 当前上传速度
	qbColl.uploadSpeedBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "upload_speed_bytes",
		Help:        "当前上传速度 单位字节",
		ConstLabels: ConstLabels,
	})
	// 种子
	qbColl.trackerTorrent = prometheus.NewDesc(
		namespace+"_tracker_torrent",
		"种子",
		[]string{"torrent_hash", "torrent_name", "tracker"},
		ConstLabels,
	)
	// 种子状态
	qbColl.trackerTorrentStatus = prometheus.NewDesc(
		namespace+"_tracker_torrent_status",
		"种子状态",
		[]string{"torrent_hash", "torrent_name", "tracker"},
		ConstLabels,
	)
	// 种子选中大小
	qbColl.trackerTorrentSizeBytes = prometheus.NewDesc(
		namespace+"_tracker_torrent_size_bytes",
		"种子大小 单位字节",
		[]string{"torrent_hash", "torrent_name", "tracker"},
		ConstLabels,
	)
	// 种子已下载
	qbColl.trackerTorrentDownloadBytesTotal = prometheus.NewDesc(
		namespace+"_tracker_torrent_download_bytes_total",
		"种子已下载 单位字节",
		[]string{"torrent_hash", "torrent_name", "tracker"},
		ConstLabels,
	)
	// 种子已上传
	qbColl.trackerTorrentUploadBytesTotal = prometheus.NewDesc(
		namespace+"_tracker_torrent_upload_bytes_total",
		"种子已上传 单位字节",
		[]string{"torrent_hash", "torrent_name", "tracker"},
		ConstLabels,
	)
	// 状态计数
	qbColl.torrentsCount = prometheus.NewDesc(
		namespace+"_torrents_count",
		"种子计数",
		[]string{"status", "tracker"},
		ConstLabels,
	)
	//
	if o.MaxDownSpeed != 0 {
		qbColl.maxDownloadSpeedBytes = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "max_download_speed_bytes",
			Help:        "Client ",
			ConstLabels: ConstLabels,
		})
		qbColl.maxDownloadSpeedBytes.Set(float64(o.MaxDownSpeed))
	}
	if o.MaxUpSpeed != 0 {
		qbColl.maxUploadSpeedBytes = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "max_upload_speed_bytes",
			Help:        "Whether if server is alive or not",
			ConstLabels: ConstLabels,
		})
		qbColl.maxUploadSpeedBytes.Set(float64(o.MaxDownSpeed))
	}
	return &qbColl
}

func (q *QbittorrentCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- q.up.Desc()
	descs <- q.uploadBytesTotal
	descs <- q.downloadBytesTotal
	descs <- q.trackerTorrentDownloadBytesTotal
	descs <- q.trackerTorrentUploadBytesTotal
	descs <- q.trackerTorrent
	descs <- q.trackerTorrentSizeBytes
	descs <- q.trackerTorrentStatus
	descs <- q.torrentsCount

	if q.Options.MaxDownSpeed != 0 {
		descs <- q.maxDownloadSpeedBytes.Desc()
	}
	if q.Options.MaxUpSpeed != 0 {
		descs <- q.maxDownloadSpeedBytes.Desc()
	}
}

func (q *QbittorrentCollector) Collect(metrics chan<- prometheus.Metric) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if !q.Options.DownloaderExporter {
		if q.Options.MaxDownSpeed != 0 {
			metrics <- q.maxDownloadSpeedBytes
		}
		if q.Options.MaxUpSpeed != 0 {
			metrics <- q.maxUploadSpeedBytes
		}
	}

	// 判断是否登录 未登录进行登录
	if !q.qbittorrentClient.IsLogin {
		if err := q.qbittorrentClient.Login(); err != nil {
			q.up.Set(0)
			metrics <- q.up
			return
		}
	}
	mainData, err := q.qbittorrentClient.GetMainData()
	if err != nil {
		q.up.Set(0)
		metrics <- q.up
		return
	} else {
		q.up.Set(1)
		metrics <- q.up
	}

	metrics <- prometheus.MustNewConstMetric(
		q.downloadBytesTotal,
		prometheus.CounterValue,
		float64(mainData.ServerState.AlltimeDl),
	)
	metrics <- prometheus.MustNewConstMetric(
		q.uploadBytesTotal,
		prometheus.CounterValue,
		float64(mainData.ServerState.AlltimeUl),
	)
	q.downloadSpeedBytes.Set(float64(mainData.ServerState.DlInfoSpeed))
	metrics <- q.downloadSpeedBytes
	q.uploadSpeedBytes.Set(float64(mainData.ServerState.UpInfoSpeed))
	metrics <- q.uploadSpeedBytes
	q.freeSpaceOnDisk.Set(float64(mainData.ServerState.FreeSpaceOnDisk))
	metrics <- q.freeSpaceOnDisk
	// torrent 相关
	state := make(map[string]map[string]int)
	for _, torrent := range mainData.Torrents {
		trackerUrl, _ := url.Parse(torrent.Tracker)
		trackerAddress := trackerUrl.Hostname()
		trackerName, isok := q.Options.RewriteTracker[trackerAddress]
		if !isok {
			trackerName = trackerAddress
		}
		// torrent
		if !q.Options.DownloaderExporter {
			metrics <- prometheus.MustNewConstMetric(
				q.trackerTorrent,
				prometheus.CounterValue,
				float64(1),
				torrent.Hash,
				torrent.Name,
				trackerName,
			)
		}
		// torrent 大小
		if !q.Options.DownloaderExporter {
			metrics <- prometheus.MustNewConstMetric(
				q.trackerTorrentSizeBytes,
				prometheus.GaugeValue,
				float64(torrent.Size),
				torrent.Hash,
				torrent.Name,
				trackerName,
			)
		}
		// 种子下载字节数
		metrics <- prometheus.MustNewConstMetric(
			q.trackerTorrentDownloadBytesTotal,
			prometheus.CounterValue,
			float64(torrent.Downloaded),
			torrent.Hash,
			torrent.Name,
			trackerName,
		)
		// 种子上传字节数
		metrics <- prometheus.MustNewConstMetric(
			q.trackerTorrentUploadBytesTotal,
			prometheus.CounterValue,
			float64(torrent.Uploaded),
			torrent.Hash,
			torrent.Name,
			trackerName,
		)

		// 种子转态重写
		if q.Options.DownloaderExporter {
			stateName := q.RewriteStatusStr(torrent.State)
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
		} else {
			metrics <- prometheus.MustNewConstMetric(
				q.trackerTorrentStatus,
				prometheus.CounterValue,
				q.RewriteStatusInt(torrent.State),
				torrent.Hash,
				torrent.Name,
				trackerName,
			)
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

func (q *QbittorrentCollector) RewriteStatusInt(status string) float64 {
	switch status {
	case "unknown":
		return 0
	case "allocating":
		return 1
	case "downloading", "metaDL", "forcedDL":
		return 2
	case "uploading", "forcedUP":
		return 3
	case "checkingUP", "checkingDL", "checkingResumeData":
		return 4
	case "missingFiles", "error":
		return 5
	case "stalledUP", "stalledDL":
		return 6
	case "queuedUP", "queuedDL":
		return 7
	case "pausedUP", "pausedDL":
		return 8
	case "moving":
		return 9
	default:
		return 10
	}
}

func (q *QbittorrentCollector) RewriteStatusStr(status string) string {
	if q.Options.Lang == "zh" {
		switch status {
		case "unknown":
			return "未知"
		case "allocating":
			return "分配"
		case "downloading", "metaDL", "forcedDL":
			return "下载中"
		case "uploading", "forcedUP":
			return "上传中"
		case "checkingUP", "checkingDL", "checkingResumeData":
			return "校验"
		case "missingFiles", "error":
			return "错误"
		case "stalledUP", "stalledDL":
			return "等待"
		case "queuedUP", "queuedDL":
			return "排队"
		case "pausedUP", "pausedDL":
			return "暂停"
		case "moving":
			return "移动中"
		default:
			return status
		}
	} else {
		switch status {
		case "unknown":
			return "Unknown"
		case "allocating":
			return "Allocating"
		case "downloading", "metaDL", "forcedDL":
			return "Downloading"
		case "uploading", "forcedUP":
			return "Uploading"
		case "checkingUP", "checkingDL", "checkingResumeData":
			return "Checking"
		case "missingFiles", "error":
			return "Errored"
		case "stalledUP", "stalledDL":
			return "Stalled"
		case "queuedUP", "queuedDL":
			return "Queued"
		case "pausedUP", "pausedDL":
			return "Paused"
		case "moving":
			return "Moving"
		default:
			return status
		}
	}
}
