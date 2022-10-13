package collector

// Options 可选项
type Options struct {
	Lang                 string            // 状态语言 可以选择 zh en 其他报错
	MaxUpSpeed           int               // 最大上传带宽
	MaxDownSpeed         int               // 最大下载带宽
	DownloaderExporter   bool              // 是否使用Downloader_exporter兼容模式
	RewriteTracker       map[string]string // tracker重写列表
	UseCategoryAsTracker bool              // 使用分类名称作为tracker
}
