# Phase 3 Prompt

你是资深 Go 框架工程师、开源维护者、长期基础设施项目 owner。请对仓库 `github.com/silen/nuwa` 执行“非兼容性重构”的第 3 阶段。

参考结构风格：

https://github.com/gin-gonic/gin

## 全局背景

这是 Go 基础设施 / 框架仓库重构。默认 `Phase 1` 和 `Phase 2` 已完成，根包已完成瘦身和核心能力重组。

## 本阶段目标

将 `pkg` 中的基础设施能力按职责拆成清晰子包，消除命名问题和大杂烩结构。

## 要解决的核心问题

1. `pkg` 混合公共 API、基础设施实现和杂项工具
2. 包命名不规范，如 `nwHttp`、`cilickhouse`
3. `tools` 是大杂烩包
4. 基础设施能力边界不清晰

## 目标结构方向

```text
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

## 本阶段必须完成的事项

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
8. 保留合理的 `crypto`、`pool`，但位置和命名要统一
9. 拆掉 `pkg/tools`
   - 将剩余能力分流到明确公开包或 `internal`
10. 修复 import 路径和测试

## 本阶段必须遵守

1. 每个目录都必须有单一明确职责
2. 不能再保留 `tools` 大杂烩包
3. 不允许继续保留错误命名
4. 不要为兼容旧路径保留大量桥接层
5. 不要过度设计接口层

## 本阶段输出要求

1. 迁移后的目录树
2. 各子包职责说明
3. 命名修正清单
4. 主动破坏兼容点清单
5. 阶段测试结果

## 阶段性验证

1. 新目录结构成立
2. import 路径修复完成
3. 各子包测试通过或已被合理重写

## 执行要求

请直接在仓库中实施迁移。
优先完成目录边界和命名清理，不要把此阶段变成新的架构设计讨论。
