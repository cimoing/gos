# OpenTelemetry 支持说明

本文说明 `gos new --with-otel` 生成的 OpenTelemetry tracing 能力如何使用、代码落点在哪里，以及上线前需要确认哪些配置。

## 1. 能力定位

OpenTelemetry 支持是可选生成能力：

```bash
gos new traced-api --module=example.com/traced-api --with-otel
```

默认生成项目不包含 OpenTelemetry 依赖，也不会生成 `internal/observability` 目录。只有显式使用 `--with-otel` 时才会增加 tracing 相关代码和依赖。

当前支持范围：

```text
1. 启动阶段初始化 OpenTelemetry TracerProvider。
2. 使用 OTLP HTTP exporter 上报 trace。
3. HTTP router 通过 otelhttp 自动生成请求 span。
4. 使用 W3C Trace Context 和 Baggage propagator。
5. 提供基于 otelhttp.Transport 的外部 HTTP client helper。
6. api-clean 使用 otelsql 包裹 database/sql，自动生成数据库访问 span。
7. 通过环境变量控制是否启用，默认关闭。
```

当前不包含：

```text
1. metrics 指标采集。
2. logs 日志上报。
3. Redis、消息队列自动 tracing。
4. Collector、Jaeger、Tempo 等基础设施部署文件。
```

## 2. 运行配置

使用 `--with-otel` 生成后，项目会在 `.env.example` 中包含：

```env
OTEL_ENABLED=false
OTEL_SERVICE_NAME=myapp
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
OTEL_EXPORTER_OTLP_INSECURE=true
```

配置含义：

```text
OTEL_ENABLED
  是否启用 tracing。默认 false。

OTEL_SERVICE_NAME
  trace 中的 service.name。建议使用稳定服务名，例如 order-api。

OTEL_EXPORTER_OTLP_ENDPOINT
  OTLP HTTP endpoint。默认 localhost:4318，不包含 /v1/traces。

OTEL_EXPORTER_OTLP_INSECURE
  是否使用非 TLS 连接。开发环境通常为 true，生产环境按网关或 Collector 配置决定。
```

本地开启示例：

```bash
OTEL_ENABLED=true \
OTEL_SERVICE_NAME=traced-api \
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318 \
OTEL_EXPORTER_OTLP_INSECURE=true \
go run ./cmd/api
```

PowerShell 示例：

```powershell
$env:OTEL_ENABLED='true'
$env:OTEL_SERVICE_NAME='traced-api'
$env:OTEL_EXPORTER_OTLP_ENDPOINT='localhost:4318'
$env:OTEL_EXPORTER_OTLP_INSECURE='true'
go run ./cmd/api
```

## 3. 生成代码落点

`api-clean` 模板：

```text
internal/config/config.go
internal/logging/logging.go
internal/observability/otel.go
internal/observability/http_client.go
internal/infrastructure/database/database.go
internal/app/app.go
internal/interfaces/http/router.go
cmd/api/main.go
go.mod
go.sum
.env.example
```

`api-minimal` 模板：

```text
internal/config/config.go
internal/logging/logging.go
internal/observability/otel.go
internal/observability/http_client.go
internal/interfaces/http/router.go
cmd/api/main.go
go.mod
go.sum
```

核心职责：

```text
internal/config
  读取 OTEL_* 环境变量，生成 ObservabilityConfig。

internal/logging
  启用 OTEL 时从 context 中读取当前 span，并向 slog 记录补充 trace_id/span_id。

internal/observability
  创建 OTLP HTTP trace exporter、TracerProvider、resource、propagator 和 traced HTTP client helper。

internal/infrastructure/database
  api-clean 启用 OTEL 时使用 otelsql 打开 database/sql 连接，Repository 和 TxManager 继续使用 *sql.DB。

internal/app 或 cmd/api
  服务启动时初始化 OpenTelemetry，并在服务退出时 shutdown。

internal/interfaces/http/router
  启用时用 otelhttp.NewHandler 包裹 HTTP handler。
```

## 4. 请求链路

启用后，一次 HTTP 请求的链路大致如下：

```text
client
  -> net/http server
  -> otelhttp middleware
  -> generated middleware chain
  -> handler
  -> usecase
  -> repository or external dependency
```

当前自动产生的 span 主要来自 `otelhttp` 对 HTTP server 的包裹。`api-clean --with-otel` 还会通过 `otelsql` 为 database/sql 操作生成 span。Usecase、队列消费等更细粒度 span 需要业务代码按需补充。

外部 HTTP 调用可以使用生成的 traced client：

```go
client := observability.NewHTTPClient(30 * time.Second)

req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
if err != nil {
	return err
}

resp, err := client.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()
```

该 client 使用 `otelhttp.NewTransport`，会从请求 `ctx` 传播 trace context。

数据库 tracing 在 `api-clean --with-otel` 中自动接入：

```go
db, err := otelsql.Open(cfg.Driver, cfg.DSN, ...)
```

业务代码仍然使用标准库 `*sql.DB`、`*sql.Tx` 和现有 `TxManager`，不需要修改 Repository 接口。`OTEL_ENABLED=false` 时没有 exporter，instrumentation 会使用全局 no-op provider；`OTEL_ENABLED=true` 且 Collector 可用时，Query、Exec、Prepare、Tx、Rows 等 database/sql 操作会进入同一条 trace。

`api-minimal` 默认没有数据库层，因此不会生成 DB tracing 代码。

手动创建 span 示例：

```go
tracer := otel.Tracer("order-usecase")

ctx, span := tracer.Start(ctx, "CreateOrder")
defer span.End()

if err := doWork(ctx); err != nil {
	span.RecordError(err)
	return err
}
```

如果在 Usecase 中补充 span，应继续使用传入的 `ctx`，不要新建 `context.Background()`，这样才能保留 trace 传播关系。

## 5. Collector 示例

项目不内置 Collector 配置，团队可以按自身观测平台选择。下面是一个最小 OTLP HTTP receiver 示例，仅用于说明接入形态：

```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:

exporters:
  debug:
    verbosity: normal

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
```

生产环境一般会把 `debug` exporter 换成 Jaeger、Tempo、OTLP upstream、云厂商 APM 或内部可观测平台。本地 Jaeger 示例见 `docs/LOCAL_OBSERVABILITY.md`。

## 6. 与日志的关系

当前 OTEL 能力只覆盖 tracing，不替代日志。

生成项目现有日志特点：

```text
1. 使用 slog 输出结构化日志。
2. 通过 LOG_LEVEL 控制 debug/info/warn/error 级别。
3. HTTP AccessLog 中间件记录请求方法、路径、状态和耗时。
4. RequestID 中间件负责生成或透传 X-Request-ID。
5. 启用 OTEL 时日志自动补充 trace_id/span_id。
6. 默认写 stdout/stderr，不负责落本地文件。
7. 容器或平台侧负责采集、保存、检索和轮转日志。
```

推荐组合：

```text
1. 应用 stdout 输出 JSON 日志。
2. OpenTelemetry 负责 trace，上报到 Collector 或 APM。
3. 日志中保留 request_id，启用 OTEL 时包含 trace_id/span_id。
4. 日志保存、索引、告警交给 Loki、ELK、OpenSearch、云日志或平台采集器。
```

后续如果需要更强的日志链路关联，可以增加可选能力：

```text
1. 增加 --with-log-file 生成 lumberjack 文件轮转方案。
2. 提供 otelcol + Loki/Tempo 的 docker compose 示例。
3. 为业务 logger 增加统一的脱敏字段处理。
```

## 7. 验证方式

脚手架自身验证：

```bash
go test ./...
```

生成项目验证：

```bash
gos new traced-api --module=example.com/traced-api --with-otel
cd traced-api
go test ./...
go build ./cmd/api
```

运行时验证：

```bash
OTEL_ENABLED=true OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318 go run ./cmd/api
curl http://127.0.0.1:8080/healthz
```

如果 Collector 正常接收，应能看到服务名为 `OTEL_SERVICE_NAME` 的 HTTP server span。

## 8. 常见问题

未使用 `--with-otel`，为什么没有 `OTEL_*` 配置？

```text
这是预期行为。默认模板保持轻量，不引入 OpenTelemetry 依赖和代码。
```

设置了 `OTEL_ENABLED=true` 但看不到 trace？

```text
1. 确认 Collector 正在监听 OTLP HTTP 端口，通常是 4318。
2. 确认 OTEL_EXPORTER_OTLP_ENDPOINT 不要带 /v1/traces。
3. 确认网络、容器端口映射和 TLS/insecure 配置匹配。
4. 访问至少一个 HTTP 路由，例如 /healthz。
```

是否会影响默认启动？

```text
不会。即使生成了 OTEL 支持，默认 OTEL_ENABLED=false，不会初始化 exporter。
```

是否会保存日志？

```text
不会。当前只上报 trace。日志仍按应用日志方案输出到 stdout/stderr，由部署平台采集保存。
```

## 9. 后续演进建议

优先级较高的增强：

```text
1. 提供更完整的本地 otelcol + Tempo/Jaeger compose 示例。
2. 增加可选 metrics exporter 和 database/sql 连接池指标采集。
```

保持为可选能力：

```text
OpenTelemetry 依赖不应进入默认最小模板路径。只有项目明确需要观测能力时，通过 --with-otel 显式生成。
```
