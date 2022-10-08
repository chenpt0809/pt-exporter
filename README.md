# pt-exporter

使用 golang 语言编写采集下载器数据。

推荐多盒用户使用，保种机器不推荐使用。

> 当前提供兼容`downloader_exporter`功能，后期如有产生冲突可能会去掉该功能。

## 计划支持客户端

| 客户端        | 支持 | 说明  |
|------------| ---- |-----|
| qbittorent |✅ |     |
| de         | ❌ |     |
| tr         | ❌ |     |

> 项目正在前期开发中，核心功能 qbittorent 实现后将支持其他客户端。

## 数据说明

| 字段                                        |    类型     | 说明                      | 默认是否开启 | 完成状态 |
|-------------------------------------------|:---------:|-------------------------|:------:|:----:|
| `pt_up`                                   |  `Gauge`  | 客户端存活状态`1：存活 0：非存活`     |   ✅    |  ✅   |
| `pt_download_bytes_total`                 | `Counter` | 客户端下载字节数                |   ✅    |  ✅   |
| `pt_upload_bytes_total`                   | `Counter` | 客户端上传字节数                |   ✅    |  ✅   |
| `pt_download_speed`                       |  `Gauge`  | 客户端下载速度                 |   ✅    |  ✅   |
| `pt_upload_speed`                         |  `Gauge`  | 客户端上传速度                 |   ✅    |  ✅   |
| `pt_download_speed_bytes`                 |  `Gauge`  | 客户端下载速度字节数              |   ✅    |  ✅   |
| `pt_upload_speed_bytes`                   |  `Gauge`  | 客户端上传速度字节数              |   ✅    |  ✅   |
| `pt_tracker_torrent`                      |  `Counter`| 种子                      |    ✅    |  ✅   |
| `pt_tracker_torrent_status`               | `Gauge`   | 种子状态                    |   ✅     |  ✅   |
| `pt_tracker_torrent_size_bytes`           |  `Gauge`  | 种子大小                    |   ✅    |  ✅   |
| `pt_tracker_torrent_download_bytes_total` | `Counter` | 种子下载字节数                 |   ✅    |  ✅   |
| `pt_tracker_torrent_upload_bytes_total`   | `Counter` | 种子上传字节数                 |   ✅    |  ✅   |
| `pt_torrents_count`                       |  `Gauge`  | 站点种子转态数量总数 downloader兼容 |   ✅    |  ✅   |

### pt_tracker_status 值说明

|  值  | 中文说明 |    英文说明     |
|:---:|:---------:|:-----------:|
|  0  | 未知 |   Unknown   |
|  1  | 分配 | Allocating  |
|  2  | 下载中 | Downloading |
|  3  | 上传中 |  Uploading  |
|  4  |    校验  |  Checking   |
|  5  |    错误  |   Errored   |
|  6  |    等待  |   Stalled   |
|  7  |    排队  |   Queued    |
|  8  |    暂停  |   Paused    |
|  9  |    移动中  |   Moving    |
|  10 | 未定义 |      Undefined       |

> 状态 10 代表此状态为能正确匹配，可能是下载器新增加的状态，可以联系研发者。

### aip 功能
> 暂未开发
> 
> 计划添加AIP接口通过promenade tables link 功能实现删种功能