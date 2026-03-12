// Package batch TOML 配置解析器
package batch

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// TOMLConfig TOML 配置文件结构
type TOMLConfig struct {
	General  GeneralConfig  `toml:"general"`
	Themes   ThemesConfig   `toml:"themes"`
	API      APIConfig      `toml:"api"`
	Images   ImagesConfig   `toml:"images"`
	Advanced AdvancedConfig `toml:"advanced"`
	Account  AccountSection  `toml:"account"`
}

// AccountSection 账号部分
type AccountSection struct {
	Zongbaozhisheng AccountConfig `toml:"zongbaozhisheng"`
	Gongchengbao   AccountConfig `toml:"gongchengbao"`
	Zongbaoshuo    AccountConfig `toml:"zongbaoshuo"`
}

// GeneralConfig 通用配置
type GeneralConfig struct {
	MarkdownFolder  string `toml:"markdown_folder"`
	CoverFolder     string `toml:"cover_folder"`
	DefaultCover    string `toml:"default_cover"`
	ConvertMode     string `toml:"convert_mode"`
	SendInterval    int    `toml:"send_interval"`
	MaxRetries       int    `toml:"max_retries"`
	LogLevel        string `toml:"log_level"`
	Verbose         bool   `toml:"verbose"`
	ContinueOnError bool   `toml:"continue_on_error"`
}

// AccountConfig 账号配置
type AccountConfig struct {
	Name            string   `toml:"name"`
	AppID           string   `toml:"appid"`
	Secret          string   `toml:"secret"`
	Enabled         bool     `toml:"enabled"`
	PreferredThemes []string `toml:"preferred_themes"`
}

// APIConfig API 配置
type APIConfig struct {
	APIKey     string `toml:"api_key"`
	APIBaseURL string `toml:"api_base_url"`
	Timeout    int    `toml:"timeout"`
}

// ImagesConfig 图片配置
type ImagesConfig struct {
	Provider  string `toml:"provider"`
	APIKey    string `toml:"api_key"`
	APIBase   string `toml:"api_base"`
	Model     string `toml:"model"`
	Size      string `toml:"size"`
	Compress  bool   `toml:"compress"`
	MaxWidth  int    `toml:"max_width"`
	MaxSizeMB int    `toml:"max_size_mb"`
}

// AdvancedConfig 高级配置
type AdvancedConfig struct {
	Concurrency     int    `toml:"concurrency"`
	ContinueOnError bool   `toml:"continue_on_error"`
	SaveState       bool   `toml:"save_state"`
	StateFile       string `toml:"state_file"`
}

// ThemesConfig 主题配置
type ThemesConfig struct {
	RotationList []string `toml:"rotation_list"`
}

// LoadTOMLConfig 加载 TOML 配置文件
func LoadTOMLConfig(path string) (*TOMLConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg TOMLConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	setDefaults(&cfg)

	return &cfg, nil
}

// setDefaults 设置默认值
func setDefaults(cfg *TOMLConfig) {
	if cfg.General.ConvertMode == "" {
		cfg.General.ConvertMode = "api"
	}
	if cfg.General.SendInterval == 0 {
		cfg.General.SendInterval = 3
	}
	if cfg.General.MaxRetries == 0 {
		cfg.General.MaxRetries = 3
	}
	if cfg.General.LogLevel == "" {
		cfg.General.LogLevel = "info"
	}
	if cfg.API.APIBaseURL == "" {
		cfg.API.APIBaseURL = "https://www.md2wechat.cn"
	}
	if cfg.API.Timeout == 0 {
		cfg.API.Timeout = 30
	}
	if cfg.Images.Provider == "" {
		cfg.Images.Provider = "modelscope"
	}
	if cfg.Images.APIBase == "" {
		cfg.Images.APIBase = "https://api-inference.modelscope.cn"
	}
	if cfg.Images.Model == "" {
		cfg.Images.Model = "Tongyi-MAI/Z-Image-Turbo"
	}
	if cfg.Images.Size == "" {
		cfg.Images.Size = "1024x1024"
	}
	if cfg.Images.MaxWidth == 0 {
		cfg.Images.MaxWidth = 1920
	}
	if cfg.Images.MaxSizeMB == 0 {
		cfg.Images.MaxSizeMB = 5
	}
	if cfg.Advanced.Concurrency == 0 {
		cfg.Advanced.Concurrency = 1
	}
	if cfg.Advanced.StateFile == "" {
		cfg.Advanced.StateFile = "./md2nwechat_state.json"
	}
	if len(cfg.Themes.RotationList) == 0 {
		cfg.Themes.RotationList = []string{"autumn-warm", "spring-fresh", "ocean-calm"}
	}
}

// GetEnabledAccounts 获取启用的账号列表
func (c *TOMLConfig) GetEnabledAccounts() []AccountConfig {
	var accounts []AccountConfig

	accounts = append(accounts, c.Account.Zongbaozhisheng)
	accounts = append(accounts, c.Account.Gongchengbao)
	accounts = append(accounts, c.Account.Zongbaoshuo)

	var enabled []AccountConfig
	for _, acc := range accounts {
		if acc.Enabled {
			enabled = append(enabled, acc)
		}
	}
	return enabled
}

// GetMarkdownFiles 获取所有 Markdown 文件
func (c *TOMLConfig) GetMarkdownFiles() ([]string, error) {
	folder := c.General.MarkdownFolder
	if folder == "" {
		return nil, nil
	}

	if !filepath.IsAbs(folder) {
		absPath, err := filepath.Abs(folder)
		if err != nil {
			return nil, err
		}
		folder = absPath
	}

	var files []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".md") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// GetCoverForArticle 获取文章对应的封面图
func (c *TOMLConfig) GetCoverForArticle(mdFile string) string {
	baseName := strings.TrimSuffix(filepath.Base(mdFile), filepath.Ext(mdFile))

	// 尝试在封面文件夹中查找
	if c.General.CoverFolder != "" {
		coverFolder := c.General.CoverFolder
		if !filepath.IsAbs(coverFolder) {
			absPath, _ := filepath.Abs(coverFolder)
			coverFolder = absPath
		}

		extensions := []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}
		for _, ext := range extensions {
			coverPath := filepath.Join(coverFolder, baseName+ext)
			if _, err := os.Stat(coverPath); err == nil {
				return coverPath
			}
		}
	}

	// 使用默认封面
	if c.General.DefaultCover != "" {
		defaultCover := c.General.DefaultCover
		if !filepath.IsAbs(defaultCover) {
			absPath, _ := filepath.Abs(defaultCover)
			defaultCover = absPath
		}
		if _, err := os.Stat(defaultCover); err == nil {
			return defaultCover
		}
	}

	return ""
}

