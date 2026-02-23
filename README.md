# asctl

管理 App Store Connect 应用元数据的命令行工具。

从 App Store Connect 下载应用的描述、关键字、推广文本等元数据到本地纯文本文件，方便编辑和版本控制；发布时将本地元数据同步到新版本。

## 功能

- **init** — 从 App Store Connect 最新版本下载元数据，按平台和语言保存到本地目录
- **release** — 创建新版本或更新已有版本，将本地元数据同步到 App Store Connect

## 安装

```bash
go install github.com/rainbend/appstoreconnect-cli@latest
```

或从源码构建：

```bash
git clone https://github.com/rainbend/appstoreconnect-cli.git
cd appstoreconnect-cli
make build    # 输出到 bin/asctl
make install  # 安装到 $GOPATH/bin/asctl
```

## 前置准备

需要一个 App Store Connect API 密钥（.p8 文件）。参考 [创建 API 密钥](https://developer.apple.com/documentation/appstoreconnectapi/creating-api-keys-for-app-store-connect-api) 获取以下信息：

- **Key ID** — API 密钥 ID
- **Issuer ID** — 颁发者 ID
- **Private Key** — 下载的 `.p8` 私钥文件路径

认证参数可通过命令行 flag、环境变量或项目根目录的 `.env` 文件提供：

| Flag | 环境变量 | 说明 |
|------|---------|------|
| `--key-id` | `ASC_KEY_ID` | API Key ID |
| `--issuer-id` | `ASC_ISSUER_ID` | Issuer ID |
| `--private-key` | `ASC_PRIVATE_KEY` | .p8 私钥文件路径（默认 `~/.config/appstoreconnect/AuthKey_{KEY_ID}.p8`） |

## 使用

### 初始化元数据

从 App Store Connect 下载最新版本的所有本地化元数据到本地：

```bash
asctl init --app-id <APP_ID> --platform ios
```

执行后会在 `.metadata/` 目录下生成如下结构：

```
.metadata/
└── ios/
    ├── en-US/
    │   ├── description.txt
    │   ├── keywords.txt
    │   ├── promotional_text.txt
    │   └── whats_new.txt
    └── zh-Hans/
        ├── description.txt
        ├── keywords.txt
        ├── promotional_text.txt
        └── whats_new.txt
```

每个字段对应一个纯文本文件，直接编辑即可。

### 发布版本

将本地元数据同步到 App Store Connect，如果版本不存在会自动创建：

```bash
asctl release --app-id <APP_ID> --version 1.2.0 --platform ios
```

使用 `--whats-new` 一次性为所有语言设置更新说明：

```bash
asctl release --app-id <APP_ID> --version 1.2.0 --whats-new "Bug fixes and improvements"
```

如果不指定 `--whats-new`，则从各语言目录下的 `whats_new.txt` 读取。

### 参数说明

**全局参数**

| 参数 | 环境变量 | 说明 |
|------|---------|------|
| `--key-id` | `ASC_KEY_ID` | API Key ID（必填） |
| `--issuer-id` | `ASC_ISSUER_ID` | Issuer ID（必填） |
| `--private-key` | `ASC_PRIVATE_KEY` | .p8 私钥文件路径（默认 `~/.config/appstoreconnect/AuthKey_{KEY_ID}.p8`） |

**init 命令**

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--app-id` | `-a` | `ASC_APP_ID` | App Store Connect App ID（必填） |
| `--platform` | `-p` | `ios` | 平台：`ios` 或 `macos` |

**release 命令**

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--app-id` | `-a` | `ASC_APP_ID` | App Store Connect App ID（必填） |
| `--version` | `-v` | — | 版本号，如 `1.2.0`（必填） |
| `--platform` | `-p` | `ios` | 平台：`ios` 或 `macos` |
| `--whats-new` | — | — | 更新说明，覆盖所有语言的 `whats_new.txt` |

## 典型工作流

```bash
# 1. 设置环境变量（或创建 .env 文件）
export ASC_KEY_ID="your-key-id"
export ASC_ISSUER_ID="your-issuer-id"
export ASC_PRIVATE_KEY="/path/to/AuthKey.p8"
export ASC_APP_ID="123456789"

# 2. 首次初始化，拉取现有元数据
asctl init

# 3. 编辑本地元数据文件
vim .metadata/ios/en-US/description.txt

# 4. 发布新版本
asctl release --version 2.0.0 --whats-new "全新设计"
```

## License

MIT
