# Moon Readwise Server

Readwise API 兼容服务，用于接收 Moon Reader 的高亮同步数据，存储到 SQLite，并支持 JSON/CSV 导出。

## 功能

- 兼容 Readwise API v2 接口
- Moon Reader 高亮同步
- SQLite 本地存储
- JSON/CSV 数据导出
- Token 认证 (支持时序攻击防护)
- 安全 HTTP 头 (X-Content-Type-Options, X-Frame-Options 等)
- 请求体大小限制 (1MB)
- 优雅退出

## 快速开始

### 本地开发

```bash
# 编译
make

# 复制配置文件
cp config.yaml.example config.yaml
# 编辑 config.yaml 修改 token

# 启动服务
make run
```

### 生产部署 (systemd)

```bash
# 编译并安装
sudo make install

# 修改 token
sudo vim /etc/moon-readwise/moon-readwise.env

# 启动服务
sudo systemctl start moon-readwise
sudo systemctl enable moon-readwise

# 查看状态
sudo systemctl status moon-readwise
```

## 配置

### 配置文件 (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080

auth:
  token: "your-secret-token"

database:
  path: "./data/highlights.db"
```

### 环境变量

环境变量优先级高于配置文件：

| 变量 | 说明 |
|------|------|
| `MOON_AUTH_TOKEN` | 认证 Token |

### 命令行参数

```
moon-readwise [-c config] [-d] [-v]

  -c string   配置文件路径 (默认 "config.yaml")
  -d          开启调试模式
  -v          打印版本号
```

## CLI 客户端

内置命令行客户端用于测试：

```bash
# 编译客户端
make client

# 使用示例
./moon-cli hl list
./moon-cli hl create --title "书名" --text "高亮内容" --author "作者"
./moon-cli hl get 1
./moon-cli hl update 1 --text "新内容" --note "新备注"
./moon-cli hl delete 1
./moon-cli hl delete --all
./moon-cli books list
./moon-cli books get 1
./moon-cli export json
./moon-cli export csv
```

全局参数：

```
-server string   服务器地址 (默认 "http://localhost:8080")
-token string    认证 Token (必须，或设置 MOON_AUTH_TOKEN 环境变量)
```

### 命令结构

```
moon-cli [全局选项] <命令> [命令选项]

命令：
  hl|highlights <操作>  管理高亮笔记 (hl 为 highlights 的缩写)
    list               列出所有高亮
    create             创建高亮 (--title, --text 必需)
    get <ID>           获取指定高亮
    update <ID>        更新指定高亮 (--text, --note, --title, --author, --source, --color)
    delete <ID>        删除指定高亮
    delete --all       删除所有高亮

  books <操作>          管理书籍
    list               列出所有书籍
    get <ID>           获取指定书籍

  export <格式>         导出数据
    json               导出为 JSON
    csv                导出为 CSV

  help                 显示帮助信息
```

## Moon Reader 配置

在 Moon Reader 中设置：

- **URL**: `http://<服务器IP>:8080/api/moon/v2/highlights/`
- **Token**: 你的认证 Token

## Makefile 命令

| 命令 | 说明 |
|------|------|
| `make` | 编译 server 和 client |
| `make server` | 仅编译 server |
| `make client` | 仅编译 client |
| `make run` | 以调试模式启动 |
| `make test` | 运行测试 |
| `make install` | 安装到系统 (需要 sudo) |
| `make uninstall` | 从系统卸载 (需要 sudo) |
| `make clean` | 清理编译产物 |

## 安全特性

- Token 认证 (支持时序攻击防护)
- 请求体大小限制 (1MB)
- 安全 HTTP 头:
  - X-Content-Type-Options: nosniff
  - X-Frame-Options: DENY
  - X-XSS-Protection: 1; mode=block
  - Content-Security-Policy: default-src 'none'
  - Cache-Control: no-store
- Content-Type 验证 (POST/PUT/PATCH 请求)
- 环境变量 Token 覆盖
- 优雅退出

## 项目结构

```
moon2wise/
├── cmd/
│   ├── server/main.go          # 服务端入口
│   └── client/main.go          # CLI 客户端
├── internal/
│   ├── api/
│   │   ├── handlers.go         # HTTP 处理器
│   │   ├── middleware.go       # 认证中间件
│   │   ├── ratelimit.go        # 限流中间件
│   │   ├── security.go         # 安全头中间件
│   │   ├── router.go           # 路由配置
│   │   └── highlights_test.go  # 高亮测试
│   └── storage/
│       ├── models.go           # 数据模型
│       ├── sqlite.go           # SQLite 存储
│       └── sqlite_test.go      # 存储测试
├── deploy/
│   ├── moon-readwise.service   # systemd 服务文件
│   └── moon-readwise.env       # 环境变量模板
├── config.yaml.example         # 配置文件模板
├── Makefile
└── README.md
```
