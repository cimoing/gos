# 本地可观测环境示例

本文给出 `gos new --with-otel` 生成项目的本地 OpenTelemetry 验证方式。项目模板不默认生成这些基础设施文件，避免把观测平台选择固化进业务项目。

## 最小 otelcol 配置

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

## Jaeger all-in-one 示例

```yaml
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
      - "4318:4318"
    environment:
      COLLECTOR_OTLP_ENABLED: "true"
```

启动后：

```bash
OTEL_ENABLED=true \
OTEL_SERVICE_NAME=traced-api \
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318 \
OTEL_EXPORTER_OTLP_INSECURE=true \
go run ./cmd/api
```

访问接口：

```bash
curl http://127.0.0.1:8080/healthz
```

打开 Jaeger UI：

```text
http://127.0.0.1:16686
```

## 外部 HTTP client tracing

使用 `--with-otel` 生成项目时，`internal/observability` 会包含：

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

该 client 使用 `otelhttp.NewTransport`，会基于传入的 `ctx` 传播 trace context。业务代码应继续使用请求链路中的 `ctx`，不要临时创建 `context.Background()`。

## 数据库调用 tracing

`api-clean --with-otel` 会使用 `github.com/XSAM/otelsql` 打开 database/sql 连接。Repository 和 TxManager 仍然拿到标准库 `*sql.DB`，业务代码不需要为了 tracing 改接口。

当 `DB_DSN` 指向可用 MySQL 且 `OTEL_ENABLED=true` 时，请求链路中的数据库 Query、Exec、事务和 Rows 操作会产生 span。`api-minimal` 没有数据库层，因此不包含这部分代码。

## 注意事项

```text
1. OTEL_EXPORTER_OTLP_ENDPOINT 使用 host:port，不要附加 /v1/traces。
2. 容器内访问宿主机 Collector 时，endpoint 可能需要改成 host.docker.internal:4318。
3. 本地 Jaeger all-in-one 适合开发验证，不等同于生产部署方案。
4. 日志仍由 stdout/stderr 输出，trace 只负责链路数据。
```
