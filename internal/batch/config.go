// Package batch 批量推送功能 - 完整配置解析器
package batch

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// INIConfig INI 配置文件结构
type INIConfig struct {
	General       GeneralConfig
	Accounts      AccountsConfig
	API           APIConfig
	Images        ImagesConfig
	Advanced      AdvancedConfig
	Themes        ThemesConfig
	AccountConfigs map[string]AccountConfig
}

// GeneralConfig 通用配置
type GeneralConfig struct {
	MarkdownFolder string
	CoverFolder    string
	DefaultCover   string
	ConvertMode    string
	SendInterval   int
	MaxRetries     int
	LogLevel       string
	Verbose        bool
}

// AccountsConfig 账号配置
type AccountsConfig struct {
	AccountList []string
}

// AccountConfig 单个账号配置
type AccountConfig struct {
	Name            string
	AppID           string
	Secret          string
	Enabled         bool
	PreferredThemes []string
}

// APIConfig API 配置
type APIConfig struct {
	APIKey     string
	APIBaseURL string
	Timeout    int
}

// ImagesConfig 图片配置
type ImagesConfig struct {
	Provider  string
	APIKey    string
	APIBase   string
	Model     string
	Size      string
	Compress  bool
	MaxWidth  int
	MaxSizeMB int
}

// AdvancedConfig 高级配置
type AdvancedConfig struct {
	Concurrency     int
	ContinueOnError bool
	SaveState       bool
	StateFile       string
}

// ThemesConfig 主题配置
type ThemesConfig struct {
	RotationList []string
}

// LoadINIConfig 加载 INI 配置文件
func LoadINIConfig(path string) (*INIConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开配置文件失败: %w", err)
	}
	defer file.Close()

	cfg := &INIConfig{
		General: GeneralConfig{
			ConvertMode:  "api",
			SendInterval: 3,
			MaxRetries:   3,
			LogLevel:     "info",
			Verbose:      true,
		},
		Accounts: AccountsConfig{
			AccountList: []string{},
		},
		AccountConfigs: make(map[string]AccountConfig),
		API: APIConfig{
			APIBaseURL: "https://www.md2wechat.cn",
			Timeout:    30,
		},
		Images: ImagesConfig{
			Provider:  "modelscope",
			APIBase:   "https://api-inference.modelscope.cn",
			Model:     "Tongyi-MAI/Z-Image-Turbo",
			Size:      "1024x1024",
			Compress:  true,
			MaxWidth:  1920,
			MaxSizeMB: 5,
		},
		Advanced: AdvancedConfig{
			Concurrency:     1,
			ContinueOnError: true,
			SaveState:       true,
			StateFile:       "./md2nwechat_state.json",
		},
		Themes: ThemesConfig{
			RotationList: []string{"autumn-warm", "spring-fresh", "ocean-calm"},
		},
	}

	var currentSection string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			if strings.HasPrefix(currentSection, "account.") {
				accKey := strings.TrimPrefix(currentSection, "account.")
				cfg.AccountConfigs[accKey] = AccountConfig{
					Enabled:         true,
					PreferredThemes: []string{},
				}
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "general":
			parseGeneralConfig(&cfg.General, key, value)
		case "accounts":
			parseAccountsConfig(&cfg.Accounts, key, value)
		case "api":
			parseAPIConfig(&cfg.API, key, value)
		case "images":
			parseImagesConfig(&cfg.Images, key, value)
		case "advanced":
			parseAdvancedConfig(&cfg.Advanced, key, value)
		case "themes":
			parseThemesConfig(&cfg.Themes, key, value)
		default:
			if strings.HasPrefix(currentSection, "account.") {
				accKey := strings.TrimPrefix(currentSection, "account.")
				if acc, ok := cfg.AccountConfigs[accKey]; ok {
					parseAccountConfig(&acc, key, value)
					cfg.AccountConfigs[accKey] = acc
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseGeneralConfig(cfg *GeneralConfig, key, value string) {
	switch key {
	case "markdown_folder":
		cfg.MarkdownFolder = value
	case "cover_folder":
		cfg.CoverFolder = value
	case "default_cover":
		cfg.DefaultCover = value
	case "convert_mode":
		cfg.ConvertMode = value
	case "send_interval":
		cfg.SendInterval = parseInt(value, 3)
	case "max_retries":
		cfg.MaxRetries = parseInt(value, 3)
	case "log_level":
		cfg.LogLevel = value
	case "verbose":
		cfg.Verbose = parseBool(value, true)
	}
}

func parseAccountsConfig(cfg *AccountsConfig, key, value string) {
	if key == "account_list" {
		accounts := strings.Split(value, ",")
		for i, acc := range accounts {
			accounts[i] = strings.TrimSpace(acc)
		}
		cfg.AccountList = accounts
	}
}

func parseAccountConfig(cfg *AccountConfig, key, value string) {
	switch key {
	case "name":
		cfg.Name = value
	case "appid":
		cfg.AppID = value
	case "secret":
		cfg.Secret = value
	case "enabled":
		cfg.Enabled = parseBool(value, true)
	case "preferred_themes":
		themes := strings.Split(value, ",")
		for i, t := range themes {
			themes[i] = strings.TrimSpace(t)
		}
		cfg.PreferredThemes = themes
	}
}

func parseAPIConfig(cfg *APIConfig, key, value string) {
	switch key {
	case "api_key":
		cfg.APIKey = value
	case "api_base_url":
		cfg.APIBaseURL = value
	case "timeout":
		cfg.Timeout = parseInt(value, 30)
	}
}

func parseImagesConfig(cfg *ImagesConfig, key, value string) {
	switch key {
	case "provider":
		cfg.Provider = value
	case "api_key":
		cfg.APIKey = value
	case "api_base":
		cfg.APIBase = value
	case "model":
		cfg.Model = value
	case "size":
		cfg.Size = value
	case "compress":
		cfg.Compress = parseBool(value, true)
	case "max_width":
		cfg.MaxWidth = parseInt(value, 1920)
	case "max_size_mb":
		cfg.MaxSizeMB = parseInt(value, 5)
	}
}

func parseAdvancedConfig(cfg *AdvancedConfig, key, value string) {
	switch key {
	case "concurrency":
		cfg.Concurrency = parseInt(value, 1)
	case "continue_on_error":
		cfg.ContinueOnError = parseBool(value, true)
	case "save_state":
		cfg.SaveState = parseBool(value, true)
	case "state_file":
		cfg.StateFile = value
	}
}

func parseThemesConfig(cfg *ThemesConfig, key, value string) {
	if key == "rotation_list" {
		themes := strings.Split(value, ",")
		for i, t := range themes {
			themes[i] = strings.TrimSpace(t)
		}
		cfg.RotationList = themes
	}
}

// Validate 验证配置
func (c *INIConfig) Validate() error {
	if c.General.MarkdownFolder == "" {
		return fmt.Errorf("markdown_folder 配置项不能为空")
	}

	if len(c.Accounts.AccountList) == 0 {
		return fmt.Errorf("account_list 配置项不能为空")
	}

	for _, accName := range c.Accounts.AccountList {
		acc, ok := c.AccountConfigs[accName]
		if !ok {
			return fmt.Errorf("账号 %s 的配置未找到", accName)
		}
		if acc.AppID == "" {
			return fmt.Errorf("账号 %s 的 appid 不能为空", accName)
		}
		if acc.Secret == "" {
			return fmt.Errorf("账号 %s 的 secret 不能为空", accName)
		}
	}

	return nil
}

// GetEnabledAccounts 获取启用的账号列表
func (c *INIConfig) GetEnabledAccounts() []string {
	var enabled []string
	for _, accName := range c.Accounts.AccountList {
		if acc, ok := c.AccountConfigs[accName]; ok && acc.Enabled {
			enabled = append(enabled, accName)
		}
	}
	return enabled
}

// GetAccountByKey 通过 key 获取账号配置
func (c *INIConfig) GetAccountByKey(key string) (*AccountConfig, error) {
	acc, ok := c.AccountConfigs[key]
	if !ok {
		return nil, fmt.Errorf("账号 %s 未找到", key)
	}
	return &acc, nil
}

// GetMarkdownFiles 获取所有 Markdown 文件
func (c *INIConfig) GetMarkdownFiles() ([]string, error) {
	var files []string

	folder := c.General.MarkdownFolder
	if !filepath.IsAbs(folder) {
		absPath, err := filepath.Abs(folder)
		if err != nil {
			return nil, fmt.Errorf("获取绝对路径失败: %w", err)
		}
		folder = absPath
	}

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return nil, fmt.Errorf("Markdown 文件夹不存在: %s", folder)
	}

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

	if err != nil {
		return nil, fmt.Errorf("遍历文件夹失败: %w", err)
	}

	return files, nil
}

// GetCoverForArticle 获取文章对应的封面图
func (c *INIConfig) GetCoverForArticle(mdFile string) string {
	baseName := strings.TrimSuffix(filepath.Base(mdFile), filepath.Ext(mdFile))

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

func parseInt(value string, defaultVal int) int {
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultVal
	}
	return result
}

func parseBool(value string, defaultVal bool) bool {
	value = strings.ToLower(value)
	switch value {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultVal
	}
}
