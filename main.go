package main

import (
	"fmt"
	"github.com/chenpt0809/pt-exporter/client"
	"github.com/chenpt0809/pt-exporter/collector"
	"github.com/chenpt0809/pt-exporter/global"
	"github.com/chenpt0809/pt-exporter/initialize"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	viper2 "github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func main() {
	// 读取配置文件
	viper := viper2.New()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	// 配置默认日志等级
	viper.SetDefault("config.logLevel", "info")
	// 配置默认监听端口
	viper.SetDefault("config.listen", ":9200")
	// 配置默认 Downloader-exporter 兼容模式
	viper.SetDefault("config.downloader-exporter", false)
	// 配置默认语言
	viper.SetDefault("config.lang", "zh")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("读取配置文件失败", zap.Error(err))
		return
	}
	// 配置默认请求超时时间
	viper.SetDefault("config.timeout", 10)
	// 配置日志相关
	var LogLevel zap.AtomicLevel
	switch viper.GetString("config.logLevel") {
	case "debug":
		LogLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	default:
		LogLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	global.Logger = initialize.Zap(LogLevel)
	// 配置下载器
	for configKey, _ := range viper.AllSettings() {
		hostName := strings.ToUpper(configKey)
		if configKey == "config" {
			continue
		}
		clientType := viper.GetStringMapString(configKey)["type"]
		// 根据下载器配置
		switch clientType {
		case "qbittorrent":
			qbc := client.NewQbittorrentClient(
				client.QbittorrentOptions{
					Url:            viper.GetStringMapString(configKey)["host"],
					UserName:       viper.GetStringMapString(configKey)["username"],
					Password:       viper.GetStringMapString(configKey)["password"],
					RequestTimeOut: viper.GetInt("config.timeout"),
				},
			)
			collOpt := collector.Options{
				Lang:               viper.GetString("config.lang"),
				MaxUpSpeed:         viper.GetInt("config.maxupspeed"),
				MaxDownSpeed:       viper.GetInt("config.maxdownspeed"),
				DownloaderExporter: viper.GetBool("config.downloader-exporter"),
				RewriteTracker:     viper.GetStringMapString("config.rewrite"),
			}
			coll := collector.NewQbittorrentCollector(
				hostName,
				qbc,
				collOpt,
			)
			prometheus.MustRegister(coll)
			global.Logger.Info("添加监控完成\t" + hostName)
		default:
			global.Logger.Error("暂时不支持下载器类型")
		}
	}

	// 配置路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/metrics", 302) })
	http.Handle("/metrics", promhttp.Handler())
	// 配置监听
	listen := viper.GetString("config.listen")
	if !strings.Contains(listen, ":") {
		listen = ":" + listen
	}
	global.Logger.Info("监听\t" + listen)
	if err := http.ListenAndServe(listen, nil); err != nil {
		global.Logger.Error("监听失败疑似端口被占用")
		return
	}
}
