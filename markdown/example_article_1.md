# 示例文章：微信公众号批量发布工具介绍

这是一篇示例文章，用于演示 MD2Nwechat 批量推送功能。

## 功能特点

MD2Nwechat 是一个强大的命令行工具，专门用于将本地 Markdown 文件批量推送到微信公众号草稿箱。

### 核心功能

- **批量处理**: 支持一次性处理多个 Markdown 文件
- **多账号轮换**: 自动在多个公众号之间轮换发布
- **主题轮换**: 每篇文章自动使用不同的排版主题
- **断点续传**: 支持中断后继续发布
- **完善日志**: 详细的发布日志和报告

## 使用方法

1. 配置 `md2nwechat.toml` 文件
2. 运行 `md2wechat batch` 命令
3. 查看发布报告

## 代码示例

```go
// 创建批量发布器
publisher := batch.NewPublisher(cfg, log)

// 执行发布
result, err := publisher.Publish()
if err != nil {
    log.Error("发布失败", zap.Error(err))
}
```

## 总结

MD2Nwechat 让公众号内容发布变得更简单高效！
