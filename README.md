# nuwa

nuwa 是一个高性能的 Go Web 框架，基于 Gin 框架构建，采用 MVC 架构模式。该框架专注于提供企业级的性能优化、可扩展性和可维护性。

## 特性

- **MVC 架构**：清晰的三层架构（Controller、Service、Model）
- **自动路由**：基于反射的自动路由机制
- **统一错误处理**：标准化的错误码和响应格式
- **配置管理**：使用 Viper 进行配置管理，支持热更新
- **结构化日志**：集成 Logrus 提供结构化日志，支持请求追踪ID
- **参数验证**：集成中文验证信息的参数验证器
- **性能优化**：多项性能优化措施，包括缓存、连接池、压缩等

## 性能优化

### 1. 缓存机制
- 基于 Redis 的分布式缓存系统
- 支持缓存穿透保护的 GetOrSet 辅助函数
- 可配置的连接池参数

### 2. 数据库连接池优化
- 为 MySQL、SQL Server 和 ClickHouse 提供连接池配置
- 合理的默认连接池参数
- 连接池监控和告警机制

### 3. 响应压缩
- Gzip 响应压缩中间件
- 智能识别可压缩的内容类型
- 减少网络传输量

### 4. 静态资源缓存控制
- 为不同类型静态资源设置合适的缓存策略
- ETag 和 Last-Modified 支持
- 304 Not Modified 响应

### 5. 性能监控和健康检查
- 请求性能监控中间件
- 详细的性能指标收集
- 健康检查和指标查询端点
- 限流中间件

### 6. 连接池监控和告警
- 实时连接池状态监控
- 多种告警阈值设置
- 灵活的告警回调机制

## 快速开始

```go
package main

import (
    "github.com/silen/nuwa"
)

func main() {
    nuwa.Run()
}
```

## 配置

框架支持多环境配置，通过 `environment` 环境变量控制：
- `dev`：开发环境
- `test`：测试环境
- `prod`：生产环境

配置文件位于 `config/` 目录下，支持 YAML 格式。

## 中间件

框架内置了多种中间件：
- `PerformanceMonitor()`：性能监控
- `MetricsMiddleware()`：指标收集
- `CompressionMiddleware()`：响应压缩
- `RequestIDMiddleware()`：请求ID追踪

## 路由

支持多种路由方式：
- `AutoRoute()`：自动路由
- `AutoRouteAny()`：任意方法自动路由
- `REST()`：RESTful 路由
- `RESTAny()`：任意方法RESTful路由

## 日志

框架提供结构化日志系统，支持：
- 请求追踪ID
- 上下文日志记录
- 多级别日志输出
- 自定义日志格式

## 许可证

MIT