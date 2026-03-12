# MD2Nwechat 安装指南

## 第一步：安装 Go 编译环境

### 方法一：手动下载安装（推荐）

1. **下载 Go 安装包**

   访问以下任一地址下载 Go 1.24+ 安装包：

   - 官方地址：https://go.dev/dl/
   - 国内镜像：https://golang.google.cn/dl/

   选择 Windows AMD64 版本（如 `go1.24.0.windows-amd64.msi`）

2. **运行安装程序**

   双击 `.msi` 文件，按默认设置安装

3. **验证安装**

   打开新的命令行窗口，执行：
   ```cmd
   go version
   ```

### 方法二：使用 Chocolatey（需要管理员权限）

以管理员身份打开 PowerShell：
```powershell
choco install golang -y
```

### 方法三：使用 Scoop

```powershell
scoop install go
```

## 第二步：配置 Go 环境

设置 Go 代理（国内用户推荐）：
```cmd
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

## 第三步：编译项目

```cmd
cd e:\CPOPC\MD2Nwechat
go mod tidy
go build -o md2wechat.exe ./cmd/md2wechat
```

## 第四步：配置 API Key

编辑 `md2nwechat.toml` 文件，填入必要的 API Key：

### 已配置的智谱大模型 API Key ✅
```toml
[images]
provider = zhipu
api_key = 810287c9375844c1a15fc546721cd69c.xnkuiMfecV06kE8q
api_base = https://open.bigmodel.cn/api/paas/v4/
model = cogview-3-plus
```

### 需要配置的 md2wechat API Key
```toml
[api]
api_key = 你的_md2wechat_API_Key
```

如需获取 md2wechat API Key：
- 访问：https://www.md2wechat.cn/api-docs
- 联系：通过官网联系获取

## 第五步：准备封面图片

在 `covers/` 文件夹中准备封面图片：
- 支持格式：JPG, PNG, WEBP, GIF
- 命名规则：与 Markdown 文件同名
- 或配置默认封面：`default_cover = ./cover.jpg`

## 第六步：运行测试

```cmd
# 预览模式（不实际发布）
md2wechat.exe batch --dry-run md2nwechat.toml

# 正式发布
md2wechat.exe batch md2nwechat.toml
```

## 常见问题

### Q: Go 下载很慢怎么办？
A: 使用国内镜像 https://golang.google.cn/dl/

### Q: go mod tidy 报错？
A: 设置代理后重试：
```cmd
go env -w GOPROXY=https://goproxy.cn,direct
```

### Q: 编译成功但运行报错？
A: 检查配置文件路径和 API Key 是否正确

## 下一步

安装完成后，请运行：
```cmd
md2wechat.exe batch --dry-run md2nwechat.toml
```

进行预览测试。
