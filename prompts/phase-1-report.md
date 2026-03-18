# Phase 1 Report

本文件是 `prompts/phase-1.md` 的执行结果，只覆盖 Phase 1：结构审计与迁移蓝图，不进行大规模迁移。

## 1. 当前架构摘要

### 1.1 总体结构

当前仓库可以分成三层，但边界混杂：

1. 根包 `nuwa`
   - 承担 Gin 引擎初始化、运行、panic 恢复、验证器注册、反射路由、响应码、`Controller/Service/Model` 基类。
   - 代表文件：`app.go`、`autorouter.go`、`validator.go`、`code.go`、`controller.go`、`service.go`、`model.go`。

2. `router/handler`
   - 实际承担的是 Gin 中间件，而不是 “handler/controller”。
   - 当前包含请求 ID、token 校验两个 HTTP 中间件。

3. `pkg/*`
   - 同时混放配置、日志、数据库、Redis、Kafka、HTTP 客户端、token、加密、协程池、工具函数。
   - 其中一部分是对外能力，一部分是内部基础设施实现，一部分只是历史遗留工具。

### 1.2 当前职责分布

- HTTP 框架核心集中在根包：
  - `app.go` 同时负责引擎生命周期、运行模式、健康检查、时间初始化、panic 恢复、错误中断。
  - `autorouter.go` 提供基于反射的方法路由。
  - `validator.go` 注册默认校验器并输出中文错误。

- 运行时依赖强依赖全局状态：
  - `pkg/conf` 在 `init()` 中读配置并启动 `WatchConfig`。
  - `pkg/logs` 在 `init()` 中初始化默认 logger。
  - `pkg/pool` 在 `init()` 中创建全局 ants 池。
  - 根包 `app.go` 在 `init()` 中初始化全局 Gin engine。

- 业务模板式抽象仍然存在：
  - `Controller`、`Service`、`Model` 通过嵌入式上下文和日志，提供 `Response/Message/Result` 一类基类能力。
  - 这更像业务脚手架，不像基础设施型框架。

## 2. 当前问题列表

按严重程度排序：

### P0: 初始化副作用过重，导入即触发运行时行为

- `app.go` 在 `init()` 中创建全局 `ginEngine`、注册 validator、设置 Gin mode。
- `pkg/conf/conf.go` 在 `init()` 中查找配置目录、读取配置、反序列化 Redis 配置、启动文件监听。
- `pkg/logs/logger.go`、`pkg/pool/goroutine.go` 也在 `init()` 中初始化全局对象。

问题：

- 包导入和程序启动强耦合，难以测试、难以复用、难以在不同环境显式装配。
- 文件系统、配置、全局引擎、全局池等副作用分散在多个包。
- 框架型仓库不应依赖隐式初始化完成装配。

### P0: `StopRun()` 使用 panic 作为正常控制流

- `controller.go` 的 `Response()`、`Message()` 在输出 JSON 后调用 `StopRun()`。
- `app.go` 的 `RecoverPanic()` 把 `errAbort` 当作正常中断。

问题：

- 正常 HTTP 响应依赖 panic/recover，控制流不透明。
- 该模式把“异常恢复”和“业务返回”混在一起，增加维护和调试成本。
- 后续若中间件栈变化，容易引入响应重复写入或吞错。

### P1: 根包职责过重，不是清晰的框架核心

根包同时承载：

- 引擎生命周期
- 响应码
- panic 恢复
- 校验器
- 自动路由
- 业务模板基类

问题：

- 根包 API 过于松散，难以形成稳定、精简的外部使用面。
- 公共核心与历史包袱没有分层。

### P1: `pkg` 目录混合公共 API、适配层和杂项工具

当前 `pkg` 不是清晰边界，而是历史堆积区：

- `pkg/conf` 既有配置装载，也有响应结构 `ReturnMap`，还混有 Redis/ClickHouse 配置类型。
- `pkg/db` 依赖 `pkg/conf` 和 `pkg/cilickhouse`。
- `pkg/nwHttp`、`pkg/tools` 既有对外函数，也有内部实现细节。

问题：

- 无法区分哪些包应公开稳定，哪些仅应内部使用。
- 未来迁移时容易形成新的耦合扩散。

### P1: `Controller` / `Service` / `Model` 属于业务模板残留

这三个类型的职责主要是：

- 挂载 `context.Context`
- 挂载 logger
- 返回统一响应结构

问题：

- 这是典型业务脚手架抽象，不是 Go 基础设施框架的核心抽象。
- `conf.ReturnMap` 被 `Controller` 和 `Service` 复用，进一步扩大错误边界。

### P2: `router/handler` 语义不清晰

当前目录名叫 `handler`，但文件内容实际是中间件：

- `request_id.go`
- `token.go`

问题：

- 使用者很难从目录名判断职责。
- 应迁移到 `middleware/`，并按职责命名。

### P2: 命名不规范

- `pkg/nwHttp`：非 idiomatic Go 命名，应改为 `httpclient` 或 `httpx`。
- `pkg/cilickhouse`：目录名拼写错误，包名却是 `ck`。
- `router/handler`：职责名不准确。

### P2: `tools` 是大杂烩

`pkg/tools` 当前同时包含：

- JSON 转换
- IP 获取
- MD5
- 随机串
- 小数处理
- panic 包装

问题：

- 无单一职责。
- 很多内容应归属于更明确的包，或进入 `internal`。

## 3. 现有代码分类

### 3.1 框架核心

- `app.go`
- `autorouter.go`
- `validator.go`
- `code.go`

说明：

- 这些文件构成了当前最接近框架核心的能力，但实现上仍夹带大量运行时副作用和历史兼容逻辑。

### 3.2 基础设施适配

- `pkg/db/*`
- `pkg/redis/*`
- `pkg/kafka/*`
- `pkg/token/*`
- `pkg/nwHttp/*`
- `pkg/crypto/*`
- `pkg/pool/*`
- `pkg/logs/*`
- `pkg/conf/*` 中的配置装载部分
- `pkg/cilickhouse/*`

说明：

- 这些能力更适合成为独立子包，由框架核心按需引用，而不是继续藏在 `pkg` 杂糅目录下。

### 3.3 历史包袱

- `controller.go`
- `service.go`
- `model.go`
- `pkg/conf/type.go` 中的 `ReturnMap`
- `pkg/tools/safe.go`
- `pkg/tools/func.go` 中的杂项函数

说明：

- 这些代码明显服务于业务模板式开发习惯，不是基础设施框架的核心形态。

### 3.4 命名问题

- `pkg/nwHttp`
- `pkg/cilickhouse`
- `router/handler`

### 3.5 初始化副作用问题

- `app.go:init`
- `pkg/conf/conf.go:init`
- `pkg/logs/logger.go:init`
- `pkg/pool/goroutine.go:init`
- `app.go:initTimeFunc` 在运行路径中隐式设置全局 carbon 默认值

## 4. 目标架构设计

目标不是业务模板，而是“小而明确的框架核心 + 清晰边界的扩展子包”。

### 4.1 根包职责

根包 `nuwa` 只保留框架核心公共 API：

- 引擎构建与启动
- 运行模式
- 恢复机制
- 自动路由能力
- 响应输出能力
- 响应码
- 默认校验器装配入口

建议根包目标文件：

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
```

### 4.2 子包职责

```text
middleware/
  requestid.go
  token.go

binding/
  validator.go

config/
  loader.go
  redis.go
  clickhouse.go

logger/
  logger.go

database/
  gorm.go
  logger.go
  clickhouse/
    conn.go

cache/redis/
  redis.go

mq/kafka/
  kafka.go

httpclient/
  client.go

token/
  token.go

crypto/
  aes.go

pool/
  ants.go
  goroutine.go

internal/
  abort/
  runtime/
```

### 4.3 设计原则

1. 根包不再持有业务模板基类。
2. 包初始化不做配置读取、网络拨号、引擎装配等重副作用。
3. 配置通过显式加载或显式注入完成。
4. 中间件放入 `middleware/`，适配层进入独立子包。
5. `internal/` 只保存外部不应依赖的实现细节。

## 5. 目录迁移计划

### Phase 2: 根包瘦身与核心框架重组

目标：

- 拆分 `app.go` 为 `engine.go`、`run.go`、`mode.go`、`recovery.go`。
- 保留 `autorouter.go`、`validator.go`、`code.go`，按职责瘦身。
- 新增 `response.go`，把响应输出从 `Controller/Service` 中抽离。
- 删除 `Controller` / `Service` / `Model` 作为对外推荐抽象。
- 清除 `StopRun()` 的 panic 正常控流，改为显式响应结束。

### Phase 3: 基础设施包重构与目录迁移

目标：

- `pkg/conf` -> `config`
- `pkg/logs` -> `logger`
- `pkg/db` -> `database`
- `pkg/cilickhouse` -> `database/clickhouse`
- `pkg/redis` -> `cache/redis`
- `pkg/kafka` -> `mq/kafka`
- `pkg/nwHttp` -> `httpclient`
- `pkg/token` -> `token`
- `pkg/pool` -> `pool`
- `pkg/crypto` -> `crypto`
- `pkg/tools` 按职责拆散或内收 `internal`

### Phase 4: 中间件、绑定、内部实现清理

目标：

- `router/handler/request_id.go` -> `middleware/requestid.go`
- `router/handler/token.go` -> `middleware/token.go`
- 默认 validator 实现迁移到 `binding/`
- 将响应终止、运行时辅助、内部兼容逻辑收拢到 `internal/`
- 清理过渡性别名和脏结构

### Phase 5: 收尾、格式化、测试、最终报告

目标：

- 清理空目录和废弃文件
- 执行 `gofmt`
- 执行 `go test ./...`
- 输出兼容性破坏清单和最终目录树

## 6. 后续阶段明确要删除、迁移、重写的内容

### 6.1 将删除

- `controller.go`
- `service.go`
- `model.go`
- `pkg/conf/system.go` 空文件
- `pkg/conf/type.go` 中的 `ReturnMap`
- `StopRun()` / `errAbort`
- `router/handler` 目录名本身

### 6.2 将迁移

- `router/handler/request_id.go` -> `middleware/requestid.go`
- `router/handler/token.go` -> `middleware/token.go`
- `pkg/conf/*` -> `config/*`
- `pkg/logs/*` -> `logger/*`
- `pkg/db/*` -> `database/*`
- `pkg/cilickhouse/*` -> `database/clickhouse/*`
- `pkg/redis/*` -> `cache/redis/*`
- `pkg/kafka/*` -> `mq/kafka/*`
- `pkg/nwHttp/*` -> `httpclient/*`
- `pkg/token/*` -> `token/*`
- `pkg/pool/*` -> `pool/*`
- `pkg/crypto/*` -> `crypto/*`

### 6.3 将重写或拆分

- `app.go`
  - 拆为核心运行文件，去掉一文件多职责。

- `pkg/conf/conf.go`
  - 改为显式配置加载器，移除 `init()` 读盘和 watcher 副作用。

- `pkg/tools/*`
  - JSON 相关优先并入更明确包。
  - `ServerIP`、`Defer` 等内部辅助函数进入 `internal` 或删除。

- `pkg/nwHttp/http.go`
  - 迁移到 `httpclient` 后按职责重命名和瘦身。

- `pkg/db/gorm.go`
  - 从配置读取耦合中解开，不再直接依赖全局 `conf.Config`。

## 7. 阶段性风险说明

### 7.1 兼容性风险

本次重构允许非兼容变更，但仍需控制节奏：

- 包路径变化会影响现有引用方。
- `Controller/Service/Model` 删除会影响旧代码组织方式。
- `ReturnMap` 删除会影响响应拼装方式。

### 7.2 运行时行为风险

- 去掉 `init()` 后，如果启动装配顺序设计不清晰，可能出现配置未加载、logger 未准备、engine 未初始化的问题。
- 替换 panic 控流后，必须确保响应只写一次、Abort 行为不回退。

### 7.3 中间件行为风险

- `request_id` 当前负责 header 注入、请求日志、服务启动时间透传。
- `token` 当前直接输出统一 JSON 结构。

迁移时必须保留现有外部行为，除非在后续阶段明确宣布行为调整。

### 7.4 基础设施生命周期风险

- Redis client、Kafka writer、ants pool 都使用包级缓存。
- 重构时要明确对象所有权、懒加载策略和关闭策略。

## 8. 结论

当前仓库已经具备一个轻量 Gin 扩展框架的雏形，但根包过重、`pkg` 目录失控、业务模板残留明显、初始化副作用过深。

后续重构的核心方向应是：

1. 根包只保留框架核心。
2. 子包承载明确的基础设施扩展。
3. 删除 `Controller/Service/Model` 和 panic 正常控流。
4. 清除 `pkg` 大杂烩和非规范命名。
5. 把隐式初始化改成显式装配。

到此，Phase 1 的分析、蓝图和迁移清单已经具备，可直接进入 Phase 2。
