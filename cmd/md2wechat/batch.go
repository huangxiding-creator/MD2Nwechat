package main

import (
	"fmt"
	"os"
	"time"

	"github.com/geekjourneyx/md2wechat-skill/internal/batch"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// batchCmd batch 命令
var batchCmd = &cobra.Command{
	Use:   "batch [config_file]",
	Short: "Batch publish Markdown files to WeChat drafts",
	Long: `Batch publish Markdown files to WeChat Official Account drafts.

This command reads multiple Markdown files from a folder and publishes them
to different WeChat Official Accounts in rotation, using different themes.

Features:
  - Multi-account rotation publishing
  - Theme rotation for each article
  - Resume from interruption (state persistence)
  - Configurable send intervals
  - Detailed progress reporting

Configuration file (TOML format):
  md2nwechat.toml

Example:
  md2wechat batch md2nwechat.toml
  md2wechat batch --dry-run md2nwechat.toml
  md2wechat batch --reset md2nwechat.toml`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runBatch(cmd, args); err != nil {
			responseError(err)
		}
	},
}

// batch 命令参数
var (
	batchConfigFile string
	batchDryRun     bool
	batchReset      bool
	batchVerbose    bool
)

func init() {
	batchCmd.Flags().StringVarP(&batchConfigFile, "config", "c", "md2nwechat.toml", "Config file path")
	batchCmd.Flags().BoolVar(&batchDryRun, "dry-run", false, "Dry run mode (no actual publishing)")
	batchCmd.Flags().BoolVar(&batchReset, "reset", false, "Reset state and start from beginning")
	batchCmd.Flags().BoolVarP(&batchVerbose, "verbose", "v", false, "Verbose output")

	rootCmd.AddCommand(batchCmd)
}

// runBatch 执行批量发布
func runBatch(cmd *cobra.Command, args []string) error {
	// 确定配置文件路径
	configPath := batchConfigFile
	if len(args) > 0 {
		configPath = args[0]
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	fmt.Printf("📋 加载配置文件: %s\n", configPath)

	// 加载 TOML 配置
	tomlCfg, err := batch.LoadTOMLConfig(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 创建日志
	log, _ := zap.NewProduction()
	if batchVerbose || tomlCfg.General.Verbose {
		log, _ = zap.NewDevelopment()
	}

	// 显示配置信息
	printBatchConfig(tomlCfg)

	// Dry run 模式
	if batchDryRun {
		return runBatchDryRun(tomlCfg, log)
	}

	// 创建发布器
	publisher := batch.NewPublisher(tomlCfg, log)

	// 重置状态
	if batchReset {
		fmt.Println("🔄 重置发布状态...")
		if err := publisher.ClearState(); err != nil {
			return fmt.Errorf("重置状态失败: %w", err)
		}
	}

	// 开始时间
	startTime := time.Now()
	fmt.Printf("\n🚀 开始批量发布... (开始时间: %s)\n\n", startTime.Format("2006-01-02 15:04:05"))

	// 执行发布
	result, err := publisher.Publish()
	if err != nil {
		return fmt.Errorf("批量发布失败: %w", err)
	}

	// 打印报告
	publisher.PrintProgress(result)

	return nil
}

// runBatchDryRun Dry run 模式
func runBatchDryRun(tomlCfg *batch.TOMLConfig, log *zap.Logger) error {
	fmt.Println("\n🔍 Dry Run 模式 - 不会实际发布\n")

	// 获取 Markdown 文件
	files, err := tomlCfg.GetMarkdownFiles()
	if err != nil {
		return fmt.Errorf("获取 Markdown 文件失败: %w", err)
	}

	// 获取启用的账号
	accounts := tomlCfg.GetEnabledAccounts()

	// 获取主题列表
	themes := tomlCfg.Themes.RotationList

	fmt.Printf("📁 Markdown 文件夹: %s\n", tomlCfg.General.MarkdownFolder)
	fmt.Printf("📄 文件数量: %d\n", len(files))
	fmt.Printf("👤 启用账号: %d\n", len(accounts))
	fmt.Printf("🎨 主题数量: %d\n", len(themes))

	fmt.Println("\n📄 将要处理的文件:")
	for i, file := range files {
		accIdx := i % len(accounts)
		themeIdx := i % len(themes)
		acc := accounts[accIdx]
		theme := themes[themeIdx]
		if len(acc.PreferredThemes) > 0 {
			theme = acc.PreferredThemes[themeIdx%len(acc.PreferredThemes)]
		}
		fmt.Printf("  %d. %s\n", i+1, file)
		fmt.Printf("     账号: %s | 主题: %s\n", acc.Name, theme)
	}

	fmt.Println("\n👤 启用的账号:")
	for _, acc := range accounts {
		fmt.Printf("  - %s (%s)\n", acc.Name, acc.AppID[:8]+"***")
	}

	fmt.Println("\n🎨 主题轮换列表:")
	for i, theme := range themes {
		fmt.Printf("  %d. %s\n", i+1, theme)
	}

	return nil
}

// printBatchConfig 打印配置信息
func printBatchConfig(tomlCfg *batch.TOMLConfig) {
	fmt.Println("\n========== 配置信息 ==========")
	fmt.Printf("📁 Markdown 文件夹: %s\n", tomlCfg.General.MarkdownFolder)
	fmt.Printf("🖼️ 封面文件夹: %s\n", tomlCfg.General.CoverFolder)
	fmt.Printf("🔄 转换模式: %s\n", tomlCfg.General.ConvertMode)
	fmt.Printf("⏱️ 发送间隔: %d 秒\n", tomlCfg.General.SendInterval)
	fmt.Printf("🔁 最大重试: %d 次\n", tomlCfg.General.MaxRetries)
	fmt.Printf("📝 日志级别: %s\n", tomlCfg.General.LogLevel)

	fmt.Println("\n👤 账号配置:")
	accounts := tomlCfg.GetEnabledAccounts()
	for _, acc := range accounts {
		status := "✅"
		if !acc.Enabled {
			status = "❌"
		}
		fmt.Printf("  %s %s (AppID: %s)\n", status, acc.Name, acc.AppID[:8]+"***")
		if len(acc.PreferredThemes) > 0 {
			fmt.Printf("     偏好主题: %v\n", acc.PreferredThemes)
		}
	}

	fmt.Println("\n🎨 主题轮换:")
	for i, theme := range tomlCfg.Themes.RotationList {
		fmt.Printf("  %d. %s\n", i+1, theme)
	}

	fmt.Println("==============================")
}
