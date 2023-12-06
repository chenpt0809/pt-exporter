package collector

import (
	"context"
	"fmt"
	"github.com/chenpt0809/pt-exporter/client"
	"github.com/chenpt0809/pt-exporter/global"
	"github.com/prometheus/client_golang/prometheus"
	"net/url"
	"sync"
	"time"
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
	var stime int64
	stime = time.Now().Unix()
	status, err := t.transmissionClient.Client.SessionStats(context.TODO())
	if err != nil {
		global.Logger.Debug(fmt.Sprintf("%s 获取状态信息失败 %e", t.clientName, err))
		t.Coll.up.Set(0)
		metrics <- t.Coll.up
		return
	} else {
		global.Logger.Debug(fmt.Sprintf("%s 获取状态信息成功 时间:%d秒", t.clientName, time.Now().Unix()-stime))
	}
	stime = time.Now().Unix()
	torrents, err := t.transmissionClient.Client.TorrentGetAll(context.TODO())
	if err != nil {
		global.Logger.Debug(fmt.Sprintf("%s 获取种子信息失败 %e", t.clientName, err))
		t.Coll.up.Set(0)
		metrics <- t.Coll.up
		return
	} else {
		global.Logger.Debug(fmt.Sprintf("%s 获取种子信息成功 时间:%d秒", t.clientName, time.Now().Unix()-stime))
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
	for _, torrent := range torrents {
		trackerUrl, _ := url.Parse(torrent.Trackers[0].Announce)
		trackerAddress := trackerUrl.Hostname()
		trackerName, isok := t.Options.RewriteTracker[trackerAddress]
		if !isok {
			trackerName = trackerAddress
		}
		if !t.Options.DownloaderExporter {
			metrics <- prometheus.MustNewConstMetric(
				t.Coll.torrent,
				prometheus.CounterValue,
				float64(1),
				*torrent.HashString,
				*torrent.Name,
				trackerName,
			)
		}
		if !t.Options.DownloaderExporter {
			metrics <- prometheus.MustNewConstMetric(
				t.Coll.torrentSizeBytes,
				prometheus.GaugeValue,
				float64(*torrent.TotalSize),
				*torrent.HashString,
				*torrent.Name,
				trackerName,
			)
		}
		metrics <- prometheus.MustNewConstMetric(
			t.Coll.torrentDownloadBytesTotal,
			prometheus.CounterValue,
			float64(*torrent.DownloadedEver),
			*torrent.HashString,
			*torrent.Name,
			trackerName,
		)
		// 种子上传字节数
		metrics <- prometheus.MustNewConstMetric(
			t.Coll.torrentUploadBytesTotal,
			prometheus.CounterValue,
			float64(*torrent.UploadedEver),
			*torrent.HashString,
			*torrent.Name,
			trackerName,
		)
	}

	t.Coll.up.Set(1)
	metrics <- t.Coll.up
}

func (t *TransmissionCollector) RewriteStatusStr(status string) string {
	if t.Options.Lang == "zh" {
		switch status {
		case "downloading":
			return "下载中"
		case "seeding":
			return "上传中"
		case "checking":
			return "校验"
		case "check pending", "download pending", "seed pending":
			return "排队"
		case "stopped":
			return "暂停"
		default:
			return status
		}
	} else {
		switch status {
		case "downloading":
			return "Downloading"
		case "seeding":
			return "Uploading"
		case "checking":
			return "Checking"
		case "check pending", "download pending", "seed pending":
			return "Queued"
		case "stopped":
			return "Paused"
		default:
			return status
		}
	}
}
