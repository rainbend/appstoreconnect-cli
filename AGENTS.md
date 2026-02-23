## AppStoreConnect-Cli

管理 App Store Connect 应用元数据的命令行工具，使用 Go 实现。

### 功能需求

1. `init` 命令：从 App Store Connect 最新版本下载推广文本、描述、关键字和新增内容，按平台（iOS/macOS）和语言区分，保存到本地 `.metadata/` 目录
2. `release` 命令：创建新版本或更新已有版本，推广文本、描述、关键字从本地 `.metadata/` 读取，新增内容（whats_new）每次发布前更新；版本已存在时触发更新而非报错
3. 认证方式：JWT (ES256)，参考 https://developer.apple.com/documentation/appstoreconnectapi/creating-api-keys-for-app-store-connect-api

### 项目结构

```
├── main.go                          # 入口
├── cmd/
│   ├── root.go                      # 根命令，全局认证参数 (--key-id, --issuer-id, --private-key)
│   ├── init.go                      # init 子命令：下载元数据
│   └── release.go                   # release 子命令：创建/更新版本
├── internal/
│   ├── api/
│   │   ├── client.go                # HTTP 客户端，JWT token 生成与缓存
│   │   ├── types.go                 # App Store Connect API JSON:API 类型定义
│   │   ├── versions.go              # 版本相关 API (List, Create)
│   │   └── localizations.go         # 本地化相关 API (List, Create, Update)
│   └── metadata/
│       └── metadata.go              # 本地 metadata 文件读写 (Read, Write, ListLocales)
└── .metadata/                       # 运行时生成的元数据目录（隐藏目录）
    └── {platform}/                  # ios 或 macos
        └── {locale}/                # 如 en-US, zh-Hans
            ├── description.txt
            ├── keywords.txt
            ├── promotional_text.txt
            └── whats_new.txt
```

### 技术栈

- Go 1.23+
- github.com/spf13/cobra — CLI 框架
- github.com/golang-jwt/jwt/v5 — JWT 签名 (ES256)
- App Store Connect API v1 (JSON:API 格式)，Base URL: `https://api.appstoreconnect.apple.com/v1`

### 关键设计

- **认证**：通过 .p8 私钥文件生成 ES256 JWT，token 缓存 15 分钟自动刷新。三个参数支持命令行 flag 和环境变量 (`ASC_KEY_ID`, `ASC_ISSUER_ID`, `ASC_PRIVATE_KEY`)
- **API 客户端** (`internal/api/client.go`)：`Client.do(method, path, body)` 是核心请求方法，自动附加 Bearer token，处理 JSON 序列化和错误响应
- **类型系统** (`internal/api/types.go`)：使用 Go 泛型 `ListResponse[T]` / `SingleResponse[T]` 解析 JSON:API 响应；请求体通过 `Create*Request` / `Update*Request` 结构体构建
- **平台映射**：CLI 使用 `ios`/`macos`，API 使用 `IOS`/`MAC_OS`，通过 `ParsePlatform()` 转换
- **元数据存储**：每个字段一个纯文本文件，方便直接编辑和版本控制
- **release 流程**：先查版本是否存在 → 不存在则创建 → 读取本地 metadata → 对比远端已有 localization → 存在则 PATCH 更新，不存在则 POST 创建
- `--whats-new` flag 可一次性覆盖所有语言的 whats_new，否则从各语言的 `whats_new.txt` 读取

### 编码规范

- 使用 `internal/` 包限制内部实现不被外部引用
- API 方法按资源拆分文件（versions.go, localizations.go），新增资源类型时新建文件
- 错误处理使用 `fmt.Errorf("context: %w", err)` 包装
- CLI 输出使用 `fmt.Printf` 打印进度，警告用 `fmt.Fprintf(os.Stderr, ...)`
