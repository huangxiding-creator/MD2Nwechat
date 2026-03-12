// Package batch 批量推送功能 - Publisher 实现
package batch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/geekjourneyx/md2wechat-skill/internal/config"
	"github.com/geekjourneyx/md2wechat-skill/internal/converter"
	"github.com/geekjourneyx/md2wechat-skill/internal/draft"
	"github.com/geekjourneyx/md2wechat-skill/internal/wechat"
	"go.uber.org/zap"
)

// PublishState 发布状态（用于断点续传）
type PublishState struct {
	sync.RWMutex
	LastFile      string    `json:"last_file"`
	LastAccount   string    `json:"last_account"`
	LastTheme     string    `json:"last_theme"`
	FileIndex     int       `json:"file_index"`
	AccountIndex  int       `json:"account_index"`
	ThemeIndex    int       `json:"theme_index"`
	Completed     []string  `json:"completed"`
	Failed        []string  `json:"failed"`
	LastUpdate    time.Time `json:"last_update"`
}

// PublishResult 发布结果
type PublishResult struct {
	TotalFiles    int             `json:"total_files"`
	SuccessCount  int             `json:"success_count"`
	FailCount     int             `json:"fail_count"`
	SkippedCount  int             `json:"skipped_count"`
	Details       []PublishDetail `json:"details"`
	StartTime     time.Time       `json:"start_time"`
	EndTime       time.Time       `json:"end_time"`
	Duration      time.Duration   `json:"duration"`
}

// PublishDetail 发布详情
type PublishDetail struct {
	File        string        `json:"file"`
	Title       string        `json:"title"`
	AccountKey  string        `json:"account_key"`
	AccountName string        `json:"account_name"`
	Theme       string        `json:"theme"`
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
	MediaID     string        `json:"media_id,omitempty"`
	Duration    time.Duration `json:"duration"`
}

// Publisher 批量发布器
type Publisher struct {
	tomlCfg   *TOMLConfig
	log       *zap.Logger
	state     *PublishState
	stateFile string
}

// NewPublisher 创建批量发布器
func NewPublisher(tomlCfg *TOMLConfig, log *zap.Logger) *Publisher {
	stateFile := tomlCfg.Advanced.StateFile
	if stateFile == "" {
		stateFile = "./md2nwechat_state.json"
	}

	return &Publisher{
		tomlCfg:   tomlCfg,
		log:       log,
		state:     &PublishState{Completed: []string{}, Failed: []string{}},
		stateFile: stateFile,
	}
}

// LoadState 加载状态
func (p *Publisher) LoadState() error {
	if !p.tomlCfg.Advanced.SaveState {
		return nil
	}

	data, err := os.ReadFile(p.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取状态文件失败: %w", err)
	}

	p.state.Lock()
	defer p.state.Unlock()

	if err := json.Unmarshal(data, p.state); err != nil {
		return fmt.Errorf("解析状态文件失败: %w", err)
	}

	p.log.Info("已加载上次的状态",
		zap.String("last_file", p.state.LastFile),
		zap.Int("completed", len(p.state.Completed)),
		zap.Int("failed", len(p.state.Failed)))

	return nil
}

// SaveState 保存状态
func (p *Publisher) SaveState() error {
	if !p.tomlCfg.Advanced.SaveState {
		return nil
	}

	p.state.Lock()
	p.state.LastUpdate = time.Now()
	data, err := json.MarshalIndent(p.state, "", "  ")
	p.state.Unlock()

	if err != nil {
		return fmt.Errorf("序列化状态失败: %w", err)
	}

	dir := filepath.Dir(p.stateFile)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建状态目录失败: %w", err)
		}
	}

	if err := os.WriteFile(p.stateFile, data, 0644); err != nil {
		return fmt.Errorf("写入状态文件失败: %w", err)
	}

	return nil
}

// ClearState 清除状态
func (p *Publisher) ClearState() error {
	p.state = &PublishState{
		Completed: []string{},
		Failed:    []string{},
	}

	if p.stateFile != "" {
		os.Remove(p.stateFile)
	}

	return nil
}

// Publish 执行批量发布
func (p *Publisher) Publish() (*PublishResult, error) {
	startTime := time.Now()

	// 加载状态
	if err := p.LoadState(); err != nil {
		p.log.Warn("加载状态失败", zap.Error(err))
	}

	// 获取 Markdown 文件列表
	files, err := p.tomlCfg.GetMarkdownFiles()
	if err != nil {
		return nil, fmt.Errorf("获取 Markdown 文件失败: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("没有找到 Markdown 文件")
	}

	// 获取启用的账号
	accounts := p.tomlCfg.GetEnabledAccounts()
	if len(accounts) == 0 {
		return nil, fmt.Errorf("没有启用的公众号账号")
	}

	p.log.Info("开始批量发布",
		zap.Int("file_count", len(files)),
		zap.Int("account_count", len(accounts)),
		zap.Int("theme_count", len(p.tomlCfg.Themes.RotationList)))

	// 初始化结果
	result := &PublishResult{
		TotalFiles: len(files),
		Details:    []PublishDetail{},
		StartTime:  startTime,
	}

	// 遍历文件
	for i, mdFile := range files {
		// 检查是否已完成
		if p.isCompleted(mdFile) {
			p.log.Info("跳过已完成的文件", zap.String("file", mdFile))
			result.SkippedCount++
			continue
		}

		// 轮换账号
		p.state.Lock()
		accountIdx := p.state.AccountIndex % len(accounts)
		p.state.Unlock()

		acc := accounts[accountIdx]

		// 轮换主题
		p.state.Lock()
		themeIdx := p.state.ThemeIndex
		p.state.Unlock()

		theme := p.getThemeForAccount(&acc, themeIdx)

		// 发布单篇文章
		detail := p.publishArticle(mdFile, &acc, theme, i)

		result.Details = append(result.Details, detail)

		if detail.Success {
			result.SuccessCount++
			p.markCompleted(mdFile)
		} else {
			result.FailCount++
			p.markFailed(mdFile)
		}

		// 更新索引
		p.state.Lock()
		p.state.FileIndex = i
		p.state.LastFile = mdFile
		p.state.LastAccount = acc.AppID
		p.state.LastTheme = theme
		p.state.AccountIndex++
		p.state.ThemeIndex++
		p.state.Unlock()

		// 保存状态
		if err := p.SaveState(); err != nil {
			p.log.Warn("保存状态失败", zap.Error(err))
		}

		// 发送间隔
		if i < len(files)-1 && p.tomlCfg.General.SendInterval > 0 {
			p.log.Debug("等待发送间隔", zap.Int("seconds", p.tomlCfg.General.SendInterval))
			time.Sleep(time.Duration(p.tomlCfg.General.SendInterval) * time.Second)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	p.log.Info("批量发布完成",
		zap.Int("success", result.SuccessCount),
		zap.Int("failed", result.FailCount),
		zap.Int("skipped", result.SkippedCount),
		zap.Duration("duration", result.Duration))

	return result, nil
}

// publishArticle 发布单篇文章
func (p *Publisher) publishArticle(mdFile string, acc *AccountConfig, theme string, index int) PublishDetail {
	startTime := time.Now()
	detail := PublishDetail{
		File:        mdFile,
		AccountKey:  acc.AppID,
		AccountName: acc.Name,
		Theme:       theme,
	}

	p.log.Info("正在处理文章",
		zap.Int("index", index+1),
		zap.String("file", filepath.Base(mdFile)),
		zap.String("account", acc.Name),
		zap.String("theme", theme))

	// 创建配置
	cfg := p.createConfig(acc)

	// 读取 Markdown 文件
	markdown, err := os.ReadFile(mdFile)
	if err != nil {
		detail.Error = fmt.Sprintf("读取文件失败: %v", err)
		detail.Duration = time.Since(startTime)
		return detail
	}

	// 提取标题
	title := p.extractTitle(string(markdown), filepath.Base(mdFile))
	detail.Title = title

	// 创建转换器
	conv := converter.NewConverter(cfg, p.log)

	// 构建转换请求
	req := &converter.ConvertRequest{
		Markdown: string(markdown),
		Mode:     converter.ConvertMode(p.tomlCfg.General.ConvertMode),
		Theme:    theme,
		APIKey:  p.tomlCfg.API.APIKey,
	}

	// 执行转换
	convResult := conv.Convert(req)
	if !convResult.Success {
		detail.Error = fmt.Sprintf("转换失败: %s", convResult.Error)
		detail.Duration = time.Since(startTime)
		return detail
	}

	// 获取封面图
	coverPath := p.tomlCfg.GetCoverForArticle(mdFile)
	if coverPath == "" {
		detail.Error = "未找到封面图片"
		detail.Duration = time.Since(startTime)
		return detail
	}

	// 上传封面图
	wechatSvc := wechat.NewService(cfg, p.log)
	coverResult, err := wechatSvc.UploadMaterial(coverPath)
	if err != nil {
		detail.Error = fmt.Sprintf("上传封面失败: %v", err)
		detail.Duration = time.Since(startTime)
		return detail
	}

	// 生成摘要
	digest := generateDigest(convResult.HTML, 120)

	// 创建草稿
	draftSvc := draft.NewService(cfg, p.log)
	draftResult, err := draftSvc.CreateDraft([]draft.Article{
		{
			Title:        title,
			Content:      convResult.HTML,
			Digest:       digest,
			ThumbMediaID: coverResult.MediaID,
			ShowCoverPic: 1,
		},
	})

	if err != nil {
		detail.Error = fmt.Sprintf("创建草稿失败: %v", err)
		detail.Duration = time.Since(startTime)
		return detail
	}

	detail.Success = true
	detail.MediaID = draftResult.MediaID
	detail.Duration = time.Since(startTime)

	p.log.Info("文章发布成功",
		zap.String("title", title),
		zap.String("account", acc.Name),
		zap.String("media_id", maskMediaID(draftResult.MediaID)),
		zap.Duration("duration", detail.Duration))

	return detail
}

// createConfig 创建配置
func (p *Publisher) createConfig(acc *AccountConfig) *config.Config {
	return &config.Config{
		WechatAppID:        acc.AppID,
		WechatSecret:       acc.Secret,
		DefaultConvertMode: p.tomlCfg.General.ConvertMode,
		MD2WechatAPIKey:   p.tomlCfg.API.APIKey,
		MD2WechatBaseURL:   p.tomlCfg.API.APIBaseURL,
		HTTPTimeout:        p.tomlCfg.API.Timeout,
		ImageProvider:      p.tomlCfg.Images.Provider,
		ImageAPIKey:        p.tomlCfg.Images.APIKey,
		ImageAPIBase:       p.tomlCfg.Images.APIBase,
		ImageModel:         p.tomlCfg.Images.Model,
		ImageSize:          p.tomlCfg.Images.Size,
		CompressImages:     p.tomlCfg.Images.Compress,
		MaxImageWidth:      p.tomlCfg.Images.MaxWidth,
		MaxImageSize:       int64(p.tomlCfg.Images.MaxSizeMB) * 1024 * 1024,
	}
}

// getThemeForAccount 获取账号的主题
func (p *Publisher) getThemeForAccount(acc *AccountConfig, index int) string {
	// 如果账号有偏好主题，使用偏好主题
	if len(acc.PreferredThemes) > 0 {
		return acc.PreferredThemes[index%len(acc.PreferredThemes)]
	}

	// 否则使用全局主题轮换
	if len(p.tomlCfg.Themes.RotationList) > 0 {
		return p.tomlCfg.Themes.RotationList[index%len(p.tomlCfg.Themes.RotationList)]
	}

	return "default"
}

// extractTitle 提取标题
func (p *Publisher) extractTitle(content, filename string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}

	// 如果没有找到标题，使用文件名
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// isCompleted 检查是否已完成
func (p *Publisher) isCompleted(file string) bool {
	p.state.RLock()
	defer p.state.RUnlock()

	for _, f := range p.state.Completed {
		if f == file {
			return true
		}
	}
	return false
}

// markCompleted 标记为已完成
func (p *Publisher) markCompleted(file string) {
	p.state.Lock()
	defer p.state.Unlock()

	p.state.Completed = append(p.state.Completed, file)
}

// markFailed 标记为失败
func (p *Publisher) markFailed(file string) {
	p.state.Lock()
	defer p.state.Unlock()

	p.state.Failed = append(p.state.Failed, file)
}

// maskMediaID 遮蔽 media_id
func maskMediaID(id string) string {
	if id == "" || len(id) < 8 {
		return "***"
	}
	return id[:4] + "***" + id[len(id)-4:]
}

// generateDigest 生成摘要
func generateDigest(html string, maxLen int) string {
	if maxLen == 0 {
		maxLen = 120
	}

	// 移除 HTML 标签
	text := stripHTMLTags(html)

	// 截取
	runes := []rune(text)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return text
}

// stripHTMLTags 去除 HTML 标签
func stripHTMLTags(html string) string {
	var result strings.Builder
	inTag := false

	for _, r := range html {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}

	return strings.TrimSpace(result.String())
}

// PrintProgress 打印进度
func (p *Publisher) PrintProgress(result *PublishResult) {
	fmt.Println("\n========== 批量发布报告 ==========")
	fmt.Printf("总文件数: %d\n", result.TotalFiles)
	fmt.Printf("成功: %d\n", result.SuccessCount)
	fmt.Printf("失败: %d\n", result.FailCount)
	fmt.Printf("跳过: %d\n", result.SkippedCount)
	fmt.Printf("耗时: %v\n", result.Duration)
	fmt.Println("\n详细结果:")

	for _, detail := range result.Details {
		status := "✅ 成功"
		if !detail.Success {
			status = "❌ 失败"
		}
		fmt.Printf("  %s | %s -> %s | 主题: %s | %v\n",
			status,
			filepath.Base(detail.File),
			detail.AccountName,
			detail.Theme,
			detail.Duration)
		if detail.Error != "" {
            fmt.Printf("       错误: %s\n", detail.Error)
        }
    }

	fmt.Println("==================================")
}
