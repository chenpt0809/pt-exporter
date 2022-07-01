package utils

import (
	"errors"
	"github.com/chenpt0809/pt-exporter/global"
	"regexp"
	"strconv"
	"strings"
)

func SpeedToInt(s string) (size float64, err error) {
	global.Logger.Debug("速度转换 " + s)
	s = strings.ToUpper(s)
	re := regexp.MustCompile(`\d+`)
	nums := re.Find([]byte(s))
	num, err := strconv.Atoi(string(nums))
	if err != nil {
		global.Logger.Error("未匹配到数字")
		return float64(0), errors.New("未匹配到数据")
	}
	if strings.HasSuffix(s, "GBPS") {
		return float64(num * 1024 * 1024 * 1024), nil
	} else if strings.HasSuffix(s, "MBPS") {
		return float64(num * 1024 * 1024), nil
	} else {
		global.Logger.Error("不支持的进制 仅支持 Gbps、Mbps")
		return float64(0), errors.New("不支持的进制 仅支持 Gbps、Mbps")
	}
}
