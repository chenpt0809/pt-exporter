package utils

import (
	"fmt"
	"net/url"
	"strconv"
)

func GetHostAndPort(rawURL string) (string, int, error) {
	// 解析 URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", 0, err
	}

	// 获取主机名
	host := parsedURL.Hostname()

	// 获取并转换端口号
	portStr := parsedURL.Port()
	var port int
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return "", 0, fmt.Errorf("invalid port: %s", err)
		}
	}

	return host, port, nil
}
