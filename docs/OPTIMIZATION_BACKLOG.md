# 优化 Backlog

本文记录当前项目后续优化项，避免在连续迭代中遗忘。已完成的阶段性进度仍以 `DEVELOPMENT_PLAN.md` 为准。

## 高优先级

```text
暂无。
```

## 已完成优化

```text
1. 后台任务/队列 worker 骨架
   schedule、queue 已从 cmd/api 中的简单 ticker 占位下沉到 internal/worker，提供统一的启动、停止、错误日志和 panic recover。

2. HTTP 生产默认值
   生成项目已配置 ReadHeaderTimeout、ReadTimeout、WriteTimeout、IdleTimeout 和 MaxHeaderBytes，并支持通过环境变量严格解析。

3. 配置校验文档化
   docs/CONFIG_REFERENCE.md 已整理环境变量、类型、默认值、适用模板和非法值行为。

4. 模板依赖刷新流程
   docs/TEMPLATE_DEPENDENCIES.md 已记录 go.mod.tmpl/go.sum.tmpl 刷新流程和检查清单。

5. 请求体大小限制
   生成项目已支持 HTTP_MAX_BODY_BYTES，并通过标准库 http.MaxBytesHandler 限制请求体大小。

6. gos version 基础能力
   gos version 已输出版本、commit 和构建时间，Makefile build 支持 ldflags 注入。

7. CLI 发布体验完善
   gos completion 已支持 bash、zsh、fish、powershell，docs/RELEASE.md 已记录构建、版本注入、completion 和发布前检查。

8. 外部 HTTP client tracing 示例
   --with-otel 模板已生成 observability.NewHTTPClient/NewHTTPTransport，使用 otelhttp.Transport 传播 trace context。

9. 本地可观测环境示例
   docs/LOCAL_OBSERVABILITY.md 已提供 otelcol debug exporter 和 Jaeger all-in-one 示例。

10. 数据库 tracing 可选方案
   api-clean --with-otel 已使用 github.com/XSAM/otelsql 包裹 database/sql，Repository/TxManager 继续使用 *sql.DB，不改变业务代码调用方式。
```

## 中优先级

```text
1. 安全默认值增强
   CORS 生产配置建议、敏感日志字段脱敏、panic 输出控制、健康检查和错误响应细节边界。
```

## 低优先级

```text
1. --with-log-file
   可选生成 lumberjack 文件轮转方案，默认仍保持 stdout 日志。

2. 更多模板能力开关
   例如 --with-docker、--with-ci、--with-auth、--with-redis。需要控制组合复杂度，避免测试矩阵膨胀。

3. OpenAPI 深度增强
   更完整的 schema、错误响应、请求体生成和契约校验。
```

## 当前建议顺序

```text
1. 安全默认值增强
2. OpenAPI 深度增强
3. --with-log-file
```
