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

func main()  {
	// 读取配置文件
	viper := viper2.New()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("config.logLevel", "info")
	viper.SetDefault("config.listen", ":9200")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("读取配置文件失败", zap.Error(err))
		return
	}
	var LogLevel zap.AtomicLevel
	switch viper.GetString("config.logLevel") {
	case "debug":
		LogLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	default:
		LogLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	global.Logger = initialize.Zap(LogLevel)

	for configKey, _ := range viper.AllSettings() {
		hostName := strings.ToUpper(configKey)
		if configKey == "config" {
			continue
		}
		clientType := viper.GetStringMapString(configKey)["type"]
		switch clientType {
		case "qbittorrent":
			qbc := client.NewQbittorrentClient(
				viper.GetStringMapString(configKey)["host"],
				viper.GetStringMapString(configKey)["username"],
				viper.GetStringMapString(configKey)["password"],
			)
			coll := collector.NewQbittorrentCollector(
				hostName,
				viper.GetStringMapString("config.rewrite"),
				qbc,
			)
			prometheus.MustRegister(coll)
			global.Logger.Info("添加监控完成\t" + hostName,)
		default:
			global.Logger.Error("暂时不支持下载器类型")
		}
	}
	http.Handle("/metrics", promhttp.Handler())
	listen := viper.GetString("config.listen")
	if ! strings.Contains(listen, ":") {
		listen = ":"+listen
	}
	global.Logger.Info("监听\t"+listen)
	if err := http.ListenAndServe(listen, nil); err != nil {
		global.Logger.Error("监听失败疑似端口被占用")
		return
	}
}