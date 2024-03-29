package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chenpt0809/pt-exporter/global"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type QbittorrentClient struct {
	client      *http.Client
	Address     string
	Username    string
	Password    string
	baseURL     string
	sid         string
	IsLogin     bool
	statusReq   *http.Request
	torrentReq  *http.Request
	mainDataReq *http.Request
}

type QbittorrentStatus struct {
	Connection        string `json:"connection_status"`
	DHTNodes          int64  `json:"dht_nodes"`
	Downloaded        int64  `json:"dl_info_data"`
	DownloadSpeed     int64  `json:"dl_info_speed"`
	DownloadRateLimit int64  `json:"dl_rate_limit"`
	Uploaded          int64  `json:"up_info_data"`
	UploadSpeed       int64  `json:"up_info_speed"`
	UploadRateLimit   int64  `json:"up_rate_limit"`
}

type QbittorrentTorrent struct {
	AddedOn                int64   `json:"added_on"`
	AmountLeft             int64   `json:"amount_left"`
	AutoTMM                bool    `json:"auto_tmm"`
	Category               string  `json:"category"`
	Completed              int64   `json:"completed"`
	CompletionOn           int64   `json:"completion_on"`
	DownloadLimit          int64   `json:"dl_limit"`
	DownloadSpeed          int64   `json:"dlspeed"`
	Downloaded             int64   `json:"downloaded"`
	DownloadedSession      int64   `json:"downloaded_session"`
	ETA                    int64   `json:"eta"`
	FirstLastPiecePriority bool    `json:"f_l_piece_prio"`
	ForceStart             bool    `json:"force_start"`
	Hash                   string  `json:"hash"`
	LastActivity           int64   `json:"last_activity"`
	MagnetURI              string  `json:"magnet_uri"`
	MaxRatio               float64 `json:"max_ratio"`
	MaxSeedingTime         int64   `json:"max_seeding_time"`
	Name                   string  `json:"name"`
	NumComplete            int64   `json:"num_complete"`
	NumIncomplete          int64   `json:"num_incomplete"`
	NumLeechs              int64   `json:"num_leechs"`
	NumSeeds               int64   `json:"num_seeds"`
	Priority               int64   `json:"priority"`
	Progress               float64 `json:"progress"`
	Ratio                  float64 `json:"ratio"`
	RatioLimit             int64   `json:"ratio_limit"`
	SavePath               string  `json:"save_path"`
	SeedingTimeLimit       int64   `json:"seeding_time_limit"`
	SeenComplete           int64   `json:"seen_complete"`
	SeqDownload            bool    `json:"seq_dl"`
	Size                   int64   `json:"size"`
	State                  string  `json:"state"`
	SuperSeeding           bool    `json:"super_seeding"`
	Tags                   string  `json:"tags"`
	TimeActive             int64   `json:"time_active"`
	TotalSize              int64   `json:"total_size"`
	Tracker                string  `json:"tracker"`
	UploadLimit            int64   `json:"up_limit"`
	Uploaded               int64   `json:"uploaded"`
	UploadedSession        int64   `json:"uploaded_session"`
	UploadSpeed            int64   `json:"upspeed"`
}

type QbittirrentMainData struct {
	Rid             int                           `json:"rid"`
	FullUpdate      bool                          `json:"full_update"`
	Torrents        map[string]QbittorrentTorrent `json:"torrents"`
	TorrentsRemoved []string                      `json:"torrents_removed"`
	Categories      map[string]struct {
		Name     string `json:"name"`
		SavePath string `json:"savePath"`
	} `json:"categories"`
	CategoriesRemoved []string `json:"categories_removed"`
	Tags              []string `json:"tags"`
	TagsRemoved       []string `json:"tags_removed"`
	ServerState       struct {
		AlltimeDl            int64  `json:"alltime_dl"`
		AlltimeUl            int64  `json:"alltime_ul"`
		AverageTimeQueue     int    `json:"average_time_queue"`
		ConnectionStatus     string `json:"connection_status"`
		DhtNodes             int    `json:"dht_nodes"`
		DlInfoData           int64  `json:"dl_info_data"`
		DlInfoSpeed          int    `json:"dl_info_speed"`
		DlRateLimit          int    `json:"dl_rate_limit"`
		FreeSpaceOnDisk      int64  `json:"free_space_on_disk"`
		GlobalRatio          string `json:"global_ratio"`
		QueuedIoJobs         int    `json:"queued_io_jobs"`
		Queueing             bool   `json:"queueing"`
		ReadCacheHits        string `json:"read_cache_hits"`
		ReadCacheOverload    string `json:"read_cache_overload"`
		RefreshInterval      int    `json:"refresh_interval"`
		TotalBuffersSize     int64  `json:"total_buffers_size"`
		TotalPeerConnections int    `json:"total_peer_connections"`
		TotalQueuedSize      int    `json:"total_queued_size"`
		TotalWastedSession   int64  `json:"total_wasted_session"`
		UpInfoData           int64  `json:"up_info_data"`
		UpInfoSpeed          int    `json:"up_info_speed"`
		UpRateLimit          int    `json:"up_rate_limit"`
		UseAltSpeedLimits    bool   `json:"use_alt_speed_limits"`
		WriteCacheOverload   string `json:"write_cache_overload"`
	} `json:"server_state"`
}

type QbittorrentOptions struct {
	Url            string
	UserName       string
	Password       string
	RequestTimeOut int
}

func NewQbittorrentClient(Options QbittorrentOptions) *QbittorrentClient {
	global.Logger.Debug("创建：QbittorrentClient")
	c := &QbittorrentClient{
		client:   http.DefaultClient,
		Address:  Options.Url,
		Username: Options.UserName,
		Password: Options.Password,
		baseURL:  fmt.Sprintf("%s/api/v2", Options.Url),
	}
	// 设置请求超时时长
	c.client.Timeout = time.Second * time.Duration(Options.RequestTimeOut)
	// 尝试登录
	global.Logger.Debug(fmt.Sprintf("初次登录： %s", Options.Url))
	if err := c.Login(); err != nil {
		global.Logger.Error("初次登录失败", zap.Error(err))
	}
	return c
}

// Login 登录
func (c *QbittorrentClient) Login() error {
	global.Logger.Debug("开始登录")
	loginInfo := url.Values{}
	loginInfo.Set("username", c.Username)
	loginInfo.Set("password", c.Password)
	resp, err := http.PostForm(fmt.Sprintf("%s/auth/login", c.baseURL), loginInfo)
	if err != nil {
		global.Logger.Error("登录失败：", zap.Error(err))
		c.IsLogin = false
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		global.Logger.Error("登录失败-解析错误", zap.Error(err))
		c.IsLogin = false
		return err
	}
	bodyStr := string(body)
	global.Logger.Debug("登录信息：" + c.Address + " " + bodyStr)
	if err != nil {
		global.Logger.Error("登录失败", zap.Error(err))
		c.IsLogin = false
		return err
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SID" {
			c.sid = cookie.Value
			c.IsLogin = true
		}
	}
	mainDataReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/sync/maindata", c.baseURL), nil)
	mainDataReq.AddCookie(&http.Cookie{Name: "SID", Value: c.sid})
	c.mainDataReq = mainDataReq
	statusReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/transfer/info", c.baseURL), nil)
	statusReq.AddCookie(&http.Cookie{Name: "SID", Value: c.sid})
	c.statusReq = statusReq
	torrentReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/torrents/info", c.baseURL), nil)
	torrentReq.AddCookie(&http.Cookie{Name: "SID", Value: c.sid})
	c.torrentReq = torrentReq
	c.IsLogin = true
	return nil
}

// GetStatus 获取下载器状态
func (c *QbittorrentClient) GetStatus() (QbittorrentStatus, error) {
	global.Logger.Debug("获取下载器状态" + c.Address)
	var status QbittorrentStatus
	resp, err := c.client.Do(c.statusReq)
	if err != nil {
		global.Logger.Error("请求发送失败" + c.Address)
		return status, err
	}
	if resp.StatusCode != 200 {
		global.Logger.Error("获取下载器信息失败" + c.Address + "状态码非200 状态码为:" + strconv.Itoa(resp.StatusCode))
		_ = c.Login()
		return status, errors.New("获取下载器信息失败" + c.Address + "状态码非200 状态码为:" + strconv.Itoa(resp.StatusCode))
	}
	defer resp.Body.Close()
	global.Logger.Debug("解析返回状态" + c.Address)
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		global.Logger.Error("解析客户端状态返回错误"+c.Address, zap.Error(err))
		return status, err
	}
	global.Logger.Debug("获取状态成功" + c.Address)
	return status, nil
}

// GetTorrent 获取种子状态
func (c *QbittorrentClient) GetTorrent() ([]QbittorrentTorrent, error) {
	global.Logger.Debug("获取种子信息" + c.Address)
	var torrents []QbittorrentTorrent
	resp, err := c.client.Do(c.torrentReq)
	if err != nil {
		global.Logger.Error("获取种子信息失败"+c.Address, zap.Error(err))
		return torrents, err
	}
	if resp.StatusCode != 200 {
		global.Logger.Error("获取种子信息失败" + c.Address + "状态码非200 状态码为:" + strconv.Itoa(resp.StatusCode))
		_ = c.Login()
		return torrents, errors.New("获取种子信息失败" + c.Address + "状态码非200 状态码为:" + strconv.Itoa(resp.StatusCode))
	}
	defer resp.Body.Close()
	global.Logger.Debug("解析种子信息" + c.Address)
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		global.Logger.Error("解析种子信息失败"+c.Address, zap.Error(err))
		return torrents, err
	}
	global.Logger.Debug("获取种子信息完成" + c.Address)
	return torrents, err
}

// GetMainData 获取主要数据
func (c *QbittorrentClient) GetMainData() (QbittirrentMainData, error) {
	global.Logger.Debug("获取主要数据" + c.Address)
	var mainData QbittirrentMainData
	resp, err := c.client.Do(c.mainDataReq)
	if err != nil {
		global.Logger.Error("获取主要数据失败"+c.Address, zap.Error(err))
		return mainData, err
	}
	if resp.StatusCode == 403 {
		global.Logger.Error("获取主要数据失败" + c.Address + "状态码非200 状态码为:" + strconv.Itoa(resp.StatusCode))
		_ = c.Login()
		return mainData, err
	} else if resp.StatusCode != 200 {
		global.Logger.Error("获取主要数据失败" + c.Address + "状态码非200 状态码为:" + strconv.Itoa(resp.StatusCode))
		return mainData, err
	}
	defer resp.Body.Close()
	global.Logger.Debug("解析主要数据" + c.Address)
	if err := json.NewDecoder(resp.Body).Decode(&mainData); err != nil {
		global.Logger.Error("解析种子主要数据"+c.Address, zap.Error(err))
		return mainData, err
	}
	global.Logger.Debug("获取主要信息完成" + c.Address)
	return mainData, err
}
