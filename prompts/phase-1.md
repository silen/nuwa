# Phase 1 Prompt

你是资深 Go 框架工程师、开源维护者、长期基础设施项目 owner。请对仓库 `github.com/silen/nuwa` 执行“非兼容性重构”的第 1 阶段。

参考结构风格：

https://github.com/gin-gonic/gin

## 全局背景

这个项目要被重构成 Go 基础设施 / 框架型仓库，不是业务系统模板。

本次重构明确允许：

1. 不保留旧包路径兼容
2. 不保留旧 API 命名兼容
3. 允许移动、拆分、合并、删除文件和目录
4. 允许删除不合理抽象
5. 允许重写测试

但必须保证：

1. 代码生产可用
2. 风格 idiomatic Go
3. 优先标准库和现有依赖
4. 不做业务模板化改造
5. 最终整体任务需要 `gofmt` 和 `go test ./...`

## 架构定位

不要把它改造成：

1. `cmd + internal + usecase + repository + service` 业务脚手架
2. DDD / CQRS / Clean Architecture 教条模板
3. Java 风格 base class 框架

它应被重构为：

1. 根包承载框架核心能力
2. 子包承载清晰边界的扩展能力
3. `internal` 只放内部实现细节
4. 对外 API 简洁明确

## 当前重点问题

1. 根包 `nuwa` 职责过重
2. `pkg` 目录混合公共 API、基础设施实现和杂项工具
3. 存在 `Controller` / `Service` / `Model`
4. 存在 `StopRun()` 这类 panic 控流
5. 存在较重 `init()` 副作用和全局状态
6. `router/handler` 语义不清晰
7. 包命名不规范，如 `nwHttp`、`cilickhouse`
8. `tools` 是大杂烩包

## 目标结构方向

目标结构大致如下，可微调，但原则不能变：

```text
/
  doc.go
  nuwa.go
  engine.go
  run.go
  mode.go
  recovery.go
  autoroute.go
  validator.go
  response.go
  code.go

  middleware/
  binding/
  config/
  logger/
  database/
  cache/redis/
  mq/kafka/
  httpclient/
  token/
  crypto/
  pool/
  internal/
```

## 本阶段目标

完成结构审计和迁移蓝图设计。本阶段只允许做极少量准备性改动，不做大规模迁移。

## 本阶段必须分析的范围

- `app.go`
- `autorouter.go`
- `validator.go`
- `code.go`
- `controller.go`
- `service.go`
- `model.go`
- `router/handler/*`
- `pkg/conf/*`
- `pkg/logs/*`
- `pkg/db/*`
- `pkg/redis/*`
- `pkg/kafka/*`
- `pkg/token/*`
- `pkg/nwHttp/*`
- `pkg/tools/*`
- `pkg/pool/*`
- `pkg/crypto/*`
- `pkg/cilickhouse/*`

## 本阶段必须完成的输出

1. 当前架构摘要
2. 当前问题列表，按严重程度排序
3. 按类别划分现有代码：
   - 框架核心
   - 基础设施适配
   - 历史包袱
   - 命名问题
   - 初始化副作用问题
4. 目标架构设计
5. 目录迁移计划
6. 明确列出将在后续阶段删除、迁移、重写的内容
7. 阶段性风险说明

## 本阶段禁止事项

1. 不要进行大规模文件搬迁
2. 不要重写主要逻辑
3. 不要提前全面重构所有包

## 本阶段可接受的小改动

1. 增加极少量文档
2. 增加重构跟踪文件

## 本阶段验收标准

1. 当前结构问题被说清楚
2. 目标目录和职责边界清晰
3. 后续阶段任务可以直接执行

## 执行要求

请直接在仓库内工作。
先做分析，再给出结构蓝图。
除非必要，不要修改业务逻辑。
