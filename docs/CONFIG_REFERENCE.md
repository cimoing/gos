# 生成项目配置参考

本文整理 `gos new` 生成项目的环境变量、默认值和非法值行为，便于上线前检查。

## 通用配置

| 变量 | 类型 | 默认值 | 模板 | 说明 |
| --- | --- | --- | --- | --- |
| `APP_NAME` | string | 项目名 kebab-case | api-clean, api-minimal | 服务名。 |
| `APP_ENV` | string | `local` | api-clean | 运行环境标识。 |
| `HTTP_ADDR` | string | `:8080` | api-clean, api-minimal | HTTP 监听地址。 |
| `HTTP_READ_HEADER_TIMEOUT` | duration | `5s` | api-clean, api-minimal | 读取请求头超时。 |
| `HTTP_READ_TIMEOUT` | duration | `15s` | api-clean, api-minimal | 读取完整请求超时。 |
| `HTTP_WRITE_TIMEOUT` | duration | `30s` | api-clean, api-minimal | 写响应超时。 |
| `HTTP_IDLE_TIMEOUT` | duration | `60s` | api-clean, api-minimal | keep-alive 空闲连接超时。 |
| `HTTP_MAX_HEADER_BYTES` | int | `1048576` | api-clean, api-minimal | 最大请求头大小。 |
| `HTTP_MAX_BODY_BYTES` | int64 | `10485760` | api-clean, api-minimal | 最大请求体大小，设置为 `0` 可关闭限制。 |
| `LOG_LEVEL` | enum | `info` | api-clean, api-minimal | 支持 `debug`、`info`、`warn`、`warning`、`error`。 |

## api-clean 数据配置

| 变量 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `DB_DRIVER` | string | `mysql` | 数据库 driver。 |
| `DB_DSN` | string | 空字符串 | 为空时不建立数据库连接。 |
| `DB_ENABLE_NESTED_TRANSACTION` | bool | `false` | 是否开启 savepoint 嵌套事务。 |
| `REDIS_ADDR` | string | `127.0.0.1:6379` | Redis 地址，占位配置。 |
| `REDIS_PASSWORD` | string | 空字符串 | Redis 密码，占位配置。 |
| `REDIS_DB` | int | `0` | Redis DB，占位配置。 |

## OpenTelemetry 配置

以下配置仅在使用 `gos new --with-otel` 时生成。

| 变量 | 类型 | 默认值 | 模板 | 说明 |
| --- | --- | --- | --- | --- |
| `OTEL_ENABLED` | bool | `false` | api-clean, api-minimal | 是否启用 tracing。 |
| `OTEL_SERVICE_NAME` | string | 项目名 kebab-case | api-clean, api-minimal | trace resource 中的 `service.name`。 |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | string | `localhost:4318` | api-clean, api-minimal | OTLP HTTP endpoint，不包含 `/v1/traces`。 |
| `OTEL_EXPORTER_OTLP_INSECURE` | bool | `true` | api-clean, api-minimal | 是否使用非 TLS 连接。 |

## 严格解析规则

```text
1. bool 使用 strconv.ParseBool，支持 true/false、1/0 等 Go 标准格式。
2. int 使用 strconv.Atoi，只接受十进制整数字符串。
3. int64 使用 strconv.ParseInt，只接受十进制整数字符串。
4. duration 使用 time.ParseDuration，例如 5s、30s、1m。
5. 非字符串配置解析失败时，config.Load 返回错误，应用启动失败。
6. string 配置为空字符串时使用默认值；确实需要空值的配置会显式以空字符串作为默认值，例如 DB_DSN。
```

常见错误：

```text
REDIS_DB=abc                         -> parse REDIS_DB as int
DB_ENABLE_NESTED_TRANSACTION=maybe   -> parse DB_ENABLE_NESTED_TRANSACTION as bool
HTTP_READ_TIMEOUT=abc                -> parse HTTP_READ_TIMEOUT as duration
HTTP_MAX_BODY_BYTES=abc              -> parse HTTP_MAX_BODY_BYTES as int64
OTEL_ENABLED=maybe                   -> parse OTEL_ENABLED as bool
```

## 上线检查建议

```text
1. 明确 APP_ENV、APP_NAME 和 OTEL_SERVICE_NAME。
2. 生产环境设置合理的 HTTP timeout，不直接沿用本地压测之外的临时值。
3. 如果业务依赖数据库，必须设置 DB_DSN。
4. 不把密码、Token、密钥写入日志字段。
5. 启用 OpenTelemetry 前确认 Collector endpoint、网络和 TLS/insecure 配置。
```
