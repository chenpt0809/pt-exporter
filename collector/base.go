package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Options 可选项
type Options struct {
	Lang                 string            // 状态语言 可以选择 zh en 其他报错
	MaxUpSpeed           int               // 最大上传带宽
	MaxDownSpeed         int               // 最大下载带宽
	DownloaderExporter   bool              // 是否使用Downloader_exporter兼容模式
	RewriteTracker       map[string]string // tracker重写列表
	UseCategoryAsTracker bool              // 使用分类名称作为tracker
}

type Collector struct {
	up                 prometheus.Gauge
	downloadBytesTotal *prometheus.Desc
	uploadBytesTotal   *prometheus.Desc
	downloadSpeedBytes prometheus.Gauge
	uploadSpeedBytes   prometheus.Gauge
	freeSpaceOnDisk    prometheus.Gauge
	//
	torrent        *prometheus.Desc
	trackerTorrent *prometheus.Desc
	//
	torrentSizeBytes        *prometheus.Desc
	trackerTorrentSizeBytes *prometheus.Desc
	//
	torrentStatus        *prometheus.Desc
	trackerTorrentStatus *prometheus.Desc
	//
	torrentDownloadBytesTotal        *prometheus.Desc
	trackerTorrentDownloadBytesTotal *prometheus.Desc
	//
	torrentUploadBytesTotal        *prometheus.Desc
	trackerTorrentUploadBytesTotal *prometheus.Desc
	//
	torrentsCount *prometheus.Desc
	//
	maxDownloadSpeedBytes prometheus.Gauge
	maxUploadSpeedBytes   prometheus.Gauge
}

func NewCollector(name string, host string, clientType string, o Options) *Collector {
	ConstLabels := map[string]string{
		"name":   name,
		"host":   host,
		"client": clientType,
	}
	namespace := "pt"
	if o.DownloaderExporter {
		namespace = "downloader"
		// 添加版本标签 兼容Downloader_exporter
		ConstLabels["version"] = "v0.0.0"
	}
	// 创建Collector
	Coll := Collector{}
	// 是否可用
	Coll.up = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "up",
		Help:        "客户端是否可用",
		ConstLabels: ConstLabels,
	})
	// 总下载量
	Coll.downloadBytesTotal = prometheus.NewDesc(
		namespace+"_download_bytes_total",
		"总下载 单位字节",
		nil,
		ConstLabels,
	)
	// 总上传量
	Coll.uploadBytesTotal = prometheus.NewDesc(
		namespace+"_upload_bytes_total",
		"总上传 单位字节",
		nil,
		ConstLabels,
	)
	// 默认下载地址剩余空间
	Coll.freeSpaceOnDisk = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "free_space_on_disk_bytes",
		Help:        "默认磁盘剩余空间 单位字节",
		ConstLabels: ConstLabels,
	})
	// 当前全局下载速度
	Coll.downloadSpeedBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "download_speed_bytes",
		Help:        "当前下载速度 单位字节",
		ConstLabels: ConstLabels,
	})
	// 当前全局上传速度
	Coll.uploadSpeedBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "upload_speed_bytes",
		Help:        "当前上传速度 单位字节",
		ConstLabels: ConstLabels,
	})
	//
	Coll.trackerTorrent = prometheus.NewDesc(
		fqNameRewrite(namespace+"_torrent", namespace+"_tracker_torrent", o.DownloaderExporter),
		"种子",
		[]string{"torrent_hash", "torrent_name", "tracker"},
		ConstLabels,
	)

	return &Coll
}

func fqNameRewrite(s1 string, s2 string, b bool) string {
	if !b {
		return s1
	} else {
		return s2
	}
}
