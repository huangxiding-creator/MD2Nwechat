# MD2Nwechat - Markdown 批量推送到微信公众号草稿箱工具

## 项目简介

MD2Nwechat 是一个命令行工具，用于将本地 Markdown 文件批量推送到多个微信公众号草稿箱。它基于 [md2wechat-skill](https://github.com/geekjourneyx/md2wechat-skill) 项目开发，扩展了以下核心功能:

- 批量处理 Markdown 文件
- 多公众号账号轮换
- 主题样式自动轮换
- 支持断点续传（状态持久化)
- 宠善的错误处理和重试机制

## 快速开始

### 1. 配置微信公众号

创建配置文件 `md2nwechat.toml`:

```toml
[general]
markdown_folder = "./markdown"      # Markdown 文件夹
cover_folder = "./covers"           # 封面图文件夹
default_cover = "./cover.jpg"     # 默认封面
convert_mode = "api"                    # 转换模式：api/ai
send_interval = 3                         # 发送间隔（秒）
max_retries = 3                         # 最大重试次数
continue_on_error = true                # 处理失败后继续
save_state = true                        # 保存状态（断点续传）
state_file = "./md2nwechat_state.json"

[account.zongbaozhisheng]
name = "总包之声"
appid = wx937e6194af9dabd8
secret = d5ba9b16e9bd109b0cd906244edb74e0
enabled = true
preferred_themes = ["autumn-warm", "elegant-gold"]

[account.gongchengbao]
name = "工程豹"
appid = wxab4386f9c12d3636
secret = ff9f895ed854f5df15ff88ec8b69f818
enabled = true
preferred_themes = ["ocean-calm", "bytedance"]

[account.zongbaoshuo]
name = "总包说"
appid = wx580e54435d78a2b7
secret = a3a294fbe9805b4cf65927d1282ab076
enabled = true
preferred_themes = ["spring-fresh", "chinese"]

[themes]
rotation_list = ["autumn-warm", "spring-fresh", "ocean-calm", "elegant-gold", "minimal-blue"]

[api]
api_key = "your-md2wechat-api-key"  # 可选
api_base_url = "https://www.md2wechat.cn"

[images]
provider = "modelscope"
api_key = "your-image-api-key"
compress = true
max_width = 1920
max_size_mb = 5
```

```toml

### 2. 运行批量发布

```bash
# 查看帮助
./md2wechat batch --help

# 预览模式（不实际发布）
./md2wechat batch --dry-run

# 正式发布
./md2wechat batch md2nwechat.toml

# 重置状态重新发布
./md2wechat batch --reset md2nwechat.toml
```

### 3. 查看发布报告

执行完成后会生成详细的发布报告，包括：
- 成功/失败数量
- 每篇文章的处理时间
- 草稿 Media ID

## 项目结构

```
MD2Nwechat/
├── cmd/md2wechat/       # 命令行入口
│   ├── main.go           # 主程序
│   ├── batch.go          # 批量发布命令
│   ├── convert.go        # 转换命令
│   └── ...
├── internal/
│   ├── batch/             # 批量发布模块
│   │   ├── config_toml.go  # TOML 配置解析
│   │   ├── publisher.go   # 批量发布器
│   │   └── ...
│   ├── config/             # 配置管理
│   ├── converter/          # Markdown 转换器
│   ├── draft/              # 草稿服务
│   ├── wechat/             # 微信 API
│   └── ...
├── markdown/             # Markdown 文件目录
├── covers/               # 封面图目录
├── md2wechat.toml        # 配置文件
└── README.md
```

## 公众号配置

### 添加新公众号

在 `md2nwechat.toml` 中添加新账号：

```toml
[account.new_account]
name = "新账号名称"
appid = "wx..."
secret = "..."
enabled = true
preferred_themes = ["theme1", "theme2"]
```

### 启用/禁用账号

设置 `enabled = true/false`

## 主题配置

### 可用主题

#### API 模式主题（38个）
- 基础主题：default, bytedance, apple, sports, chinese, cyber
- Minimal 系列：minimal-gold/green/blue/orange/red/navy/gray/sky
- Focus 系列：focus-gold/green/blue/orange/red/navy/gray/sky
- Elegant 系列：elegant-gold/green/blue/orange/red/navy/gray/sky
- Bold 系列：bold-gold/green/blue/orange/red/navy/gray/sky

#### AI 模式主题（3个）
- autumn-warm（秋日暖光）
- spring-fresh（春日清新）
- ocean-calm（深海静谧）

### 配置主题轮换

```toml
[themes]
rotation_list = ["autumn-warm", "spring-fresh", "ocean-calm"]
```

## 高级配置

### 断点续传

工具会自动保存发布状态到 `md2nwechat_state.json`，中断后可继续发布。

### 发送间隔

避免触发微信 API 限流：

```toml
[general]
send_interval = 3  # 秒
```

### 错误处理

```toml
[general]
continue_on_error = true  # 遇到错误继续处理
max_retries = 3           # 最大重试次数
```

## 编译安装

### 从源码编译

```bash
# 克隆项目
git clone https://github.com/geekjourneyx/md2wechat-skill.git
cd md2wechat-skill

# 编译
go build -o md2wechat ./cmd/md2wechat

# 安装到系统
sudo mv md2wechat /usr/local/bin/
```

### 下载预编译版本

访问 [Releases 页面](https://github.com/geekjourneyx/md2wechat-skill/releases) 下载最新版本。

## 常见问题

### Q: 如何查看发布进度？

A: 工具会在控制台实时显示进度，完成后生成详细报告。

### Q: 中断后如何继续？

A: 直接重新运行 `md2wechat batch` 命令，工具会自动从上次中断的位置继续。

### Q: 如何重新发布所有文章？

A: 使用 `--reset` 参数或删除状态文件 `md2nwechat_state.json`。

### Q: 支持哪些图片格式？

A: 支持 JPG, PNG, WEBP, GIF 格式的封面图。

## 许可证

MIT License

## 作者

基于 [md2wechat-skill](https://github.com/geekjourneyx/md2wechat-skill) 开发
