# Nuwa Non-Compatible Refactor Prompt

用于 Cursor / Codex / Claude 的分阶段执行版 prompt。

目标：将 `github.com/silen/nuwa` 从“混合工具库”重构为“符合 Go 第三方基础设施/框架项目规范”的仓库，结构风格参考 Gin：

https://github.com/gin-gonic/gin

## 使用方式

将本文件整体交给 agent 执行，要求它严格按 Phase 顺序推进，不允许跳阶段。

如果你想降低一次性大改失败率，可以按下面方式逐段执行：

1. 先只执行 `Phase 1`
2. 验收后再执行 `Phase 2`
3. 依次推进到 `Phase 5`

---

你是资深 Go 框架工程师、开源维护者、长期基础设施项目 owner。请对仓库 `github.com/silen/nuwa` 做一次“非兼容性重构”。

本次任务不是局部修补，而是结构级重组。目标是把当前仓库重构成一个更符合 Go 基础设施/框架项目规范的代码库，目录与职责组织风格参考 Gin。

## 全局原则

本次重构明确允许：

1. 不保留旧包路径兼容
2. 不保留旧 API 命名兼容
3. 允许移动、拆分、合并、删除文件和目录
4. 允许删除不合理抽象
5. 允许重写测试

但必须保证：

1. 代码是生产可用的 Go 代码
2. 风格 idiomatic Go
3. 功能语义尽量保持一致，除非旧设计本身明显错误
4. 优先标准库和现有依赖
5. 所有代码必须 `gofmt`
6. 最终必须执行 `go test ./...`

## 架构定位

这个项目不是业务系统模板，不要改造成：

1. `cmd + internal + usecase + repository + service` 的业务分层脚手架
2. DDD / CQRS / Clean Architecture 教条模板
3. Java 风格 base class 框架

它应该被重构为：

1. 一个 Go 基础设施 / 框架型仓库
2. 根包承载框架核心能力
3. 子包承载清晰边界的扩展能力
4. `internal` 只放内部实现细节
5. 对外 API 简洁、稳定、低心智负担

## 必须解决的核心问题

1. 根包 `nuwa` 职责过重，混合了 app lifecycle、engine、validator、autoroute、response、code、controller/service/model
2. `pkg` 目录混合公共 API、基础设施实现和杂项工具
3. 存在不符合 Go 风格的设计，如 `Controller` / `Service` / `Model`
4. 存在 `StopRun()` 这类 panic 控流
5. 存在隐式初始化和较重的 `init()` 副作用
6. 存在全局 engine、全局启动时间、全局 map 等可变状态
7. 中间件放在 `router/handler`，语义不清晰
8. 包命名不规范，如 `nwHttp`、`cilickhouse`
9. `tools` 为大杂烩包，必须拆解

## 目标结构

目标结构接近下面形式，可以微调，但原则不能变：

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
    requestid.go
    token.go

  binding/
    validator.go

  config/
    loader.go
    types.go
    constants.go

  logger/
    logger.go

  database/
    gorm.go
    gorm_logger.go
    clickhouse.go

  cache/
    redis/
      client.go

  mq/
    kafka/
      client.go

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
    ...
```

## 现有代码范围

请重点分析并改造以下内容：

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

## 输出与执行规则

你必须严格按下面的 Phase 顺序执行。

每个 Phase 都必须：

1. 先说明目标
2. 再实施代码修改
3. 输出该阶段改动摘要
4. 输出风险和未完成项
5. 做阶段性验证

除非 Phase 明确要求，否则不要提前做后续阶段的大规模改动。

---

## Phase 1: 架构审计与重构蓝图

### 目标

完成对当前仓库的结构审计，并产出可执行的迁移蓝图，但此阶段只允许做极少量准备性改动，不做大规模迁移。

### 必做事项

1. 阅读当前核心文件与子包
2. 输出当前架构摘要
3. 输出问题清单，按严重程度排序
4. 区分以下几类内容：
   - 框架核心
   - 基础设施适配
   - 历史包袱
   - 命名问题
   - 初始化副作用问题
5. 输出目标架构设计
6. 输出目录迁移计划
7. 明确列出哪些内容将在后续阶段删除、迁移、重写

### 禁止事项

1. 不要在本阶段进行大规模文件搬迁
2. 不要在本阶段重写主要逻辑
3. 不要提前重构所有包

### 可接受的小改动

1. 补充极少量注释或文档文件
2. 增加重构跟踪文件

### 本阶段输出要求

必须输出：

1. 当前架构摘要
2. 当前问题列表
3. 目标架构设计
4. 目录迁移计划
5. 阶段性风险说明

### 本阶段验收标准

1. 当前结构问题被说清楚
2. 目标目录和职责边界清晰
3. 后续阶段任务可直接执行

---

## Phase 2: 根包瘦身与核心框架重组

### 目标

把根包重构成类似 Gin 的框架核心层，只保留真正属于框架核心的能力。

### 必做事项

1. 将根包按能力拆分为清晰文件，例如：
   - `nuwa.go`
   - `engine.go`
   - `run.go`
   - `mode.go`
   - `recovery.go`
   - `autoroute.go`
   - `validator.go`
   - `response.go`
   - `code.go`
2. 删除 `Controller` / `Service` / `Model` 这类基类设计，或将其能力重构为更 Go 风格的函数或轻量类型
3. 删除 `StopRun()` 及相关 panic 控流
4. 重写响应辅助逻辑，避免依赖基类方法结束流程
5. 减少或去除全局 engine 与隐式初始化
6. 将模式、恢复、启动逻辑拆开

### 必须遵守

1. 不允许继续保留 panic 作为正常分支控制流
2. 不允许保留明显不合理的基类模式
3. 不要为了兼容旧 API 再包一层旧结构

### 阶段性验证

1. 根包文件职责清晰
2. 原核心测试尽量修复到可运行
3. 无循环依赖

### 本阶段输出要求

1. 新的根包职责说明
2. 删除了哪些旧抽象
3. 主动破坏兼容点清单
4. 阶段测试结果

---

## Phase 3: 基础设施包重构与目录迁移

### 目标

将 `pkg` 里的基础设施能力按职责拆成清晰子包，消除命名问题和大杂烩结构。

### 必做事项

1. 重构配置模块：
   - `pkg/conf` -> `config`
   - 明确 loader、types、constants 的职责
   - 降低 `init()` 副作用
2. 重构日志模块：
   - `pkg/logs` -> `logger`
   - 保留上下文感知能力
3. 重构数据库模块：
   - `pkg/db` -> `database`
   - `pkg/cilickhouse` 改名或并入 `database`
4. 重构缓存模块：
   - `pkg/redis` -> `cache/redis`
5. 重构 MQ 模块：
   - `pkg/kafka` -> `mq/kafka`
6. 重构 HTTP 客户端：
   - `pkg/nwHttp` -> `httpclient`
7. 重构 token：
   - `pkg/token` -> `token`
8. 保留合理的 `crypto`、`pool` 等包，但位置和命名要统一
9. 拆掉 `pkg/tools`
   - 将剩余能力分流到明确公开包或 `internal`

### 必须遵守

1. 每个目录都必须能回答“职责是什么”
2. 不能再保留 `tools` 大杂烩包
3. 不允许继续保留错误命名
4. 包 API 需要简洁明确

### 阶段性验证

1. 新目录结构成立
2. import 路径修复完成
3. 各子包测试通过或已被合理重写

### 本阶段输出要求

1. 迁移后的目录树
2. 子包职责说明
3. 命名修正清单
4. 阶段测试结果

---

## Phase 4: 中间件、绑定、内部实现清理

### 目标

整理横切能力，把中间件和内部辅助实现归位，进一步降低耦合。

### 必做事项

1. 将 `router/handler/*` 重构为 `middleware/*`
2. 检查 `request id` 和 `token` 中间件边界：
   - 保持无敏感信息泄露
   - 保持上下文传递清晰
3. 如果 validator 适合拆为 `binding/validator.go`，则完成拆分
4. 将不适合公开暴露的辅助实现迁移到 `internal`
5. 清理遗留桥接逻辑和历史兼容写法

### 必须遵守

1. `middleware` 目录语义必须清晰
2. `internal` 只放不应暴露的实现细节
3. 不要制造新的“万能公共包”

### 阶段性验证

1. 中间件测试可运行
2. 包边界清晰
3. 无明显多余桥接代码

### 本阶段输出要求

1. 中间件和内部包设计说明
2. 清理掉的遗留结构清单
3. 阶段测试结果

---

## Phase 5: 收尾、格式化、测试、重构报告

### 目标

完成收尾验证，并输出完整的重构交付说明。

### 必做事项

1. 全量 `gofmt`
2. 执行 `go test ./...`
3. 修复测试失败
4. 清理无用文件、无用 import、无用桥接层
5. 补充必要的 `doc.go` 或 README 更新

### 最终输出必须包含

1. 当前架构摘要
2. 当前问题列表
3. 目标架构说明
4. 重构后的目录树
5. 关键设计决策
6. 破坏兼容点清单
7. 测试结果摘要
8. 尚可继续优化的项

### 最终验收标准

1. 目录结构清晰
2. 根包职责明确
3. 基础设施包边界明确
4. 无 panic 正常控流
5. 无 `tools` 大杂烩
6. 无错误命名包
7. 代码通过格式化
8. 测试通过

---

## 强约束总结

请始终牢记：

1. 这是非兼容性重构，不要做兼容迁移包袱
2. 这是 Go 框架 / 基础设施仓库，不是业务系统模板
3. 结构风格参考 Gin 的根包 + 子包组织
4. 删除 `Controller` / `Service` / `Model`
5. 删除 panic 控流
6. 拆掉 `pkg/tools`
7. 修正 `nwHttp` / `cilickhouse`
8. 显式初始化，减少全局状态
9. 直接改代码并跑测试
10. 每个阶段都要给出阶段性结果，不允许只停留在建议层
