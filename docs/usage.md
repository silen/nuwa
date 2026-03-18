# nuwa 使用文档

本文面向当前仓库代码结构，说明 `nuwa` 的主要模块职责、启动方式和常用能力用法。

## 1. 架构概览

`nuwa` 当前分为三层：

- 根包 `nuwa`
  - HTTP 运行时核心。
  - 负责引擎创建、运行、恢复、自动路由、统一响应、状态码和 validator 装配入口。
- 横切能力
  - `binding`：Gin 请求绑定相关的默认 validator 实现。
  - `middleware`：HTTP 请求链上的 request id、token 等中间件。
- 基础设施子包
  - `config`、`logs`、`database`、`cache/redis`、`mq/kafka`、`httpclient`、`token`、`crypto`、`pool`。

建议把它理解为“一个基于 Gin 的轻量运行时核心 + 一组配套基础设施包”，不是业务脚手架。

## 2. 安装

```bash
go get github.com/silen/nuwa
```

如果你直接在当前仓库开发：

```bash
go test ./...
```

在受限环境里运行测试时，可以显式设置 Go build cache：

```bash
env GOCACHE=/tmp/nuwa-go-build go test ./...
```

## 3. 目录与职责

```text
.
├── engine.go          # 创建和管理 gin.Engine
├── run.go             # 启动、监听、优雅关闭
├── recovery.go        # panic 恢复中间件
├── response.go        # 统一响应输出
├── autoroute.go       # 反射式自动路由
├── validator.go       # 根包 validator 装配入口
├── binding/           # validator 具体实现
├── middleware/        # request id / token 中间件
├── config/            # 配置加载
├── logs/              # 默认 logger
├── cache/redis/       # Redis 客户端
├── database/          # GORM / ClickHouse 连接
├── httpclient/        # HTTP 客户端封装
├── token/             # token 存储与校验
├── pool/              # goroutine / ants 池封装
└── crypto/            # AES 等加解密能力
```

## 4. 最小可运行示例

```go
package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/silen/nuwa"
	"github.com/silen/nuwa/middleware"
)

func main() {
	engine := nuwa.New()

	engine.Use(middleware.RequestID())
	engine.GET("/ping", func(c *gin.Context) {
		nuwa.Success(c, gin.H{
			"message": "pong",
		})
	})

	if err := nuwa.RunEngineContext(context.Background(), engine); err != nil {
		log.Fatal(err)
	}
}
```

`nuwa.New()` 会完成这些初始化：

- 根据环境变量设置 Gin mode。
- 安装默认恢复中间件 `Recovery()`。
- 安装默认 validator。
- 注册默认 404 响应。

## 5. 环境与配置

### 5.1 运行模式

`environment` 控制运行模式：

- `dev`：Gin debug mode
- `test`：Gin test mode
- 其他值：Gin release mode

对应 API：

- `nuwa.Mode()`
- `nuwa.SetMode(mode string)`

### 5.2 配置加载规则

启动时，`RunE` / `RunContext` / `RunEngineContext` 会调用 `config.Load()`。

配置查找规则：

1. 如果设置了 `NUWA_CONFIG_DIR`，直接使用该目录。
2. 否则从当前工作目录开始，逐级向上查找 `config/` 目录。

环境配置文件名由 `environment` 决定：

- `dev` -> `config/dev.yaml`
- `test` -> `config/test.yaml`
- `prod` -> `config/prod.yaml`

如果设置 `NUWA_CONFIG_WATCH=true`，会开启配置文件监听。

### 5.3 最小配置示例

`run.go` 当前会读取 `server.host` 和 `server.port`，`cache/redis` 会依赖 `redis` 配置。

```yaml
server:
  host: 0.0.0.0
  port: "8080"

redis:
  host: 127.0.0.1:6379
  password: ""
  pool_size: 10
  db: 0

mysql:
  prefix: ""

clickhouse:
  address: 127.0.0.1:8123
  username: default
  password: ""
```

## 6. 启动方式

根包提供三种启动方式：

- `nuwa.Run()`
  - 启动共享 engine，失败时调用 `logs.Fatal` 退出进程。
- `nuwa.RunE()`
  - 启动共享 engine，返回错误给调用方。
- `nuwa.RunContext(ctx)`
  - 启动共享 engine，并使用外部 `context.Context` 控制退出。
- `nuwa.RunEngineContext(ctx, engine)`
  - 启动指定的 `*gin.Engine`。

运行时行为：

- 自动挂载 `/checkHealth`
- 使用 `http.Server`
- 响应 `SIGTERM` / `SIGINT`
- 默认 5 秒优雅关闭超时

## 7. 共享 Engine 与自定义 Engine

根包维护一个默认共享 engine：

- `nuwa.Default()`
- `nuwa.Engine()`
- `nuwa.SetDefaultEngine(engine)`

建议：

- 简单项目可以直接用共享 engine。
- 需要明确装配链路时，优先使用 `nuwa.New()` 自己创建 engine，再通过 `RunEngineContext` 启动。

## 8. 统一响应与参数绑定

根包提供统一响应结构：

```go
type Response struct {
    Message string `json:"message"`
    Status  int    `json:"status"`
    Data    any    `json:"data,omitempty"`
}
```

常用辅助函数：

- `nuwa.JSON(c, response)`
- `nuwa.Success(c, data)`
- `nuwa.Message(c, message, code)`
- `nuwa.Bind(c, &req)`

### 8.1 成功响应

```go
func Ping(c *gin.Context) {
	nuwa.Success(c, gin.H{"ok": true})
}
```

输出示例：

```json
{
  "message": "ok",
  "status": 200,
  "data": {
    "ok": true
  }
}
```

### 8.2 失败响应

```go
func Fail(c *gin.Context) {
	nuwa.Message(c, "参数错误", nuwa.PARAMETER_ERROR)
}
```

### 8.3 绑定请求参数

```go
type CreateUserRequest struct {
	Name string `json:"name" binding:"required"`
}

func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if !nuwa.Bind(c, &req) {
		return
	}

	nuwa.Success(c, req)
}
```

默认 validator 使用中文错误翻译，位于 `binding/validator.go`。

### 8.4 状态码常量

- `nuwa.SUCCESSFUL`
- `nuwa.PARAMETER_ERROR`
- `nuwa.UNAUTHORIZED`
- `nuwa.NOT_FOUND`
- `nuwa.EXCEPTION`
- `nuwa.SYSTEM_ERROR`

## 9. 中间件

### 9.1 `middleware.RequestID()`

作用：

- 读取或生成 `X-Request-Id`
- 回写响应头 `X-Request-Id`
- 写入运行时头 `serverIp`、`start-server-time`
- 记录请求日志
- 自动脱敏敏感 header，如 `Authorization`、`Cookie`、`token`

示例：

```go
engine.Use(middleware.RequestID())
```

### 9.2 `middleware.Token()`

作用：

- 从请求头读取 `token`
- 调用 `token.NewToken(c).Check(...)` 校验
- 校验成功后写入：
  - `c.Set("userInfo", res)`
  - `c.Set("token", tokenString)`

示例：

```go
engine.Use(middleware.Token())
```

注意：

- 当前固定读取请求头 `token`，不是 `Authorization: Bearer ...`。
- 中间件只做 HTTP 层校验，token 的存储与读取由 `token` 子包负责。

## 10. 自动路由

根包提供四个反射式路由入口：

- `nuwa.AutoRoute(handler)`
- `nuwa.AutoRouteAny(handler)`
- `nuwa.REST(handler)`
- `nuwa.RESTAny(handler)`

区别：

- `AutoRoute` / `AutoRouteAny`
  - 路径第一段解析为方法名。
- `REST` / `RESTAny`
  - 使用 HTTP method 作为方法名。
- `Any` 版本允许除 `*gin.Context` 外再接收基础类型参数。

### 10.1 `AutoRoute`

```go
type UserHandler struct{}

func (UserHandler) Show(c *gin.Context) {
	nuwa.Success(c, gin.H{"route": "show"})
}

engine.GET("/user/*path", nuwa.AutoRoute(UserHandler{}))
```

- 访问 `/user/show` 会调用 `Show`
- 如果方法不存在，返回 404

### 10.2 `AutoRouteAny`

```go
type UserHandler struct{}

func (UserHandler) Show(c *gin.Context, id int) {
	nuwa.Success(c, gin.H{"id": id})
}

engine.GET("/user/*path", nuwa.AutoRouteAny(UserHandler{}))
```

- 访问 `/user/show/123` 会调用 `Show(c, 123)`
- 当前支持基础类型参数：`int`、`int32`、`int64`、`uint`、`uint32`、`uint64`、`float32`、`float64`、`string`

## 11. 日志

默认日志包是 `logs`，基于 `logrus`。

常用 API：

- `logs.WithContext(ctx)`
- `logs.Logger()`
- `logs.SetLogger(logger)`
- `logs.SetLevel(level)`
- `logs.SetOutput(writer)`
- `logs.Info(...)`
- `logs.Error(...)`

如果 `context.Context` 里带有 `X-Request-Id`，默认 formatter 会把 request id 输出到日志行中。

## 12. Redis 与 Token

### 12.1 Redis

```go
client, err := redis.NewRedis(ctx)
if err != nil {
	return err
}
```

说明：

- 依赖 `config.Load()` 已经初始化 `config.Redis`
- 支持通过可选参数覆盖 DB：

```go
client, err := redis.NewRedis(ctx, 3)
```

### 12.2 Token

```go
tk := token.NewToken(ctx)

tokenString, err := tk.Get(map[string]any{
	"user_id": 1,
}, time.Hour)

userInfo, err := tk.Check(tokenString)
err = tk.Del(tokenString)
```

说明：

- token 默认存放在 Redis DB `3`
- `Check("")` 会返回 `token.ErrTokenNotFound`

## 13. HTTP 客户端

`httpclient` 基于 `fasthttp`，支持常见请求方式、JSON body、上传文件和超时控制。

```go
client := httpclient.NewHTTP(ctx).
	SetHeader(map[string]string{
		"X-App": "demo",
	}).
	SetTimeout(5 * time.Second)

	var out struct {
		Status int `json:"status"`
	}

	_, err := client.PostJson("http://example.com/api", map[string]any{
		"name": "nuwa",
	}, &out)
```

可用方法：

- `Get`
- `Post`
- `PostJson`
- `Send`
- `UploadFile`
- `SetHeader`
- `SetContentType`
- `SetTimeout`

辅助函数：

- `httpclient.CompressStr(str)`

## 14. 数据库

### 14.1 MySQL / SQL Server

`database` 包对 GORM 做了连接封装。

```go
db, err := database.Mysql(ctx, database.Config{
	Master:   "root:pass@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local",
	PrintSql: true,
})
```

也支持：

- `database.SQLServer(ctx, cfg)`

如果提供 `Slave`，会通过 `gorm.io/plugin/dbresolver` 注册读写分离。

### 14.2 ClickHouse

原生 SQL 连接：

```go
conn, err := clickhouse.Open(ctx, clickhouse.Config{
	Addr:     []string{"127.0.0.1:8123"},
	Database: "default",
	Username: "default",
	Password: "",
})
```

GORM 连接：

```go
db, err := database.ClickHouse(ctx, clickhouse.Config{
	Addr:     []string{"127.0.0.1:8123"},
	Database: "default",
	Username: "default",
	Password: "",
	PrintSQL: true,
})
```

## 15. Kafka

```go
reader := kafka.KafkaReader("127.0.0.1:9092|127.0.0.1:9093", "topic-demo", "group-a", 3000)
writer := kafka.KafkaWriter("127.0.0.1:9092|127.0.0.1:9093", "topic-demo")
```

说明：

- broker 使用 `|` 分隔
- `KafkaWriter` 内部按 `kafkaURL + topic` 复用 writer

## 16. 并发池

### 16.1 默认 ants 池

```go
pool.New().Submit(func() {
	// do work
})
```

### 16.2 批量并发执行

```go
items := []int{1, 2, 3}
err := pool.ExecTaskByGoroutine(context.Background(), items, func(v int) {
	// handle v
}, 20)
```

如果任务任一项报错后需要提前结束：

```go
err := pool.ExecTaskByGoroutineErrorEnd(context.Background(), items, func(v int) error {
	return nil
}, 20)
```

### 16.3 安全 goroutine

```go
pool.Go(ctx, func() {
	// do work
})
```

## 17. 加解密

`crypto` 当前提供 AES-CBC 加解密：

```go
ciphertext, err := crypto.EncryptAES(key, "hello")
plaintext, err := crypto.DecryptAES(key, ciphertext)
```

注意：

- `key` 长度必须满足 AES 要求
- 输出密文是十六进制字符串

## 18. 常见注意事项

1. `RunEngineContext` 启动前会自动调用 `config.Load()`，所以需要保证配置目录可用。
2. `middleware.Token()` 当前依赖 Redis，因此在未初始化 Redis 配置时会返回系统错误。
3. `RequestID` 中间件会向响应头写入 `serverIp` 和 `start-server-time`。
4. `pool` 和 `logs` 目前仍保留默认全局对象模式，适合快速接入，但对多实例装配并不完全隔离。
5. `AutoRouteAny` / `RESTAny` 的参数解析仅支持基础类型。

## 19. 推荐接入顺序

对于一个新 HTTP 服务，建议按下面顺序接入：

1. 准备 `config/dev.yaml`
2. `engine := nuwa.New()`
3. 挂载 `middleware.RequestID()`
4. 注册业务路由
5. 如需鉴权，再挂载 `middleware.Token()`
6. 通过 `nuwa.Success` / `nuwa.Message` 统一输出响应
7. 用 `nuwa.RunEngineContext` 启动

## 20. 相关文档

- [README](../README.md)
- [Prompts Overview](../prompts/README.md)
