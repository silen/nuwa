# nuwa

`nuwa` 是一个以 Gin 为运行时基础的 Go HTTP 框架核心与基础设施仓库。

更详细的接入和模块使用说明见：

- [docs/usage.md](./docs/usage.md)

当前仓库已经完成从旧 `pkg/*` 混合工具库形态到“根包框架核心 + 清晰边界子包”结构的重组：

- 根包 `nuwa`
  - 引擎构建、运行、恢复、自动路由、响应辅助、状态码和 validator 装配入口。
- `middleware`
  - HTTP 横切能力，如 request id 和 token 中间件。
- `binding`
  - Gin 绑定相关的默认 validator 实现。
- 基础设施子包
  - `config`、`logs`、`database`、`cache/redis`、`mq/kafka`、`httpclient`、`token`、`crypto`、`pool`。
- `internal`
  - 仅供仓库内部使用的实现细节。

## 目录概览

```text
.
├── autoroute.go
├── binding/
├── cache/redis/
├── config/
├── crypto/
├── database/
│   └── clickhouse/
├── engine.go
├── httpclient/
├── internal/
├── logs/
├── middleware/
├── mode.go
├── mq/kafka/
├── pool/
├── recovery.go
├── response.go
├── run.go
├── token/
└── validator.go
```

## 核心用法

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
		nuwa.Success(c, gin.H{"message": "pong"})
	})

	if err := nuwa.RunEngineContext(context.Background(), engine); err != nil {
		log.Fatal(err)
	}
}
```

## 运行约定

- `environment` 用于选择运行模式和配置环境。
  - `dev` -> Gin debug mode
  - `test` -> Gin test mode
  - 其他值 -> Gin release mode
- 配置默认从当前目录向上查找 `config/` 目录，也可以通过 `NUWA_CONFIG_DIR` 显式指定。
- `RunE` / `RunContext` / `RunEngineContext` 会在启动时加载配置并挂载健康检查路由 `/checkHealth`。

## 设计约束

- 不再保留 `Controller` / `Service` / `Model` 基类。
- 不再使用 panic 作为正常响应控制流。
- 不再保留旧 `pkg/*` 路径兼容层。
- 内部辅助实现不再通过“万能工具包”暴露。

## 验证

```text
env GOCACHE=/tmp/nuwa-go-build go test ./...
```
