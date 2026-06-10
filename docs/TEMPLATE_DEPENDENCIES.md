# 模板依赖刷新流程

本文说明如何维护生成项目模板中的 `go.mod.tmpl` 和 `go.sum.tmpl`，避免依赖升级时遗漏间接依赖或校验和。

## 依赖来源

当前模板依赖主要来自：

```text
api-clean
  github.com/bradfitz/gomemcache  Memcache cache backend
  github.com/go-sql-driver/mysql
  github.com/redis/go-redis/v9     Redis cache backend and distributed lock
  github.com/spf13/cobra
  github.com/XSAM/otelsql         仅 --with-otel，用于 database/sql tracing
  go.opentelemetry.io/*            仅 --with-otel

api-minimal
  github.com/spf13/cobra
  go.opentelemetry.io/*            仅 --with-otel

gos CLI
  github.com/spf13/cobra
```

## 刷新原则

```text
1. 默认模板不应引入 OpenTelemetry 依赖。
2. --with-otel 的依赖只放在模板条件块中。
3. api-clean 和 api-minimal 使用相同 Cobra 版本。
4. gos CLI 和生成项目优先使用相同 Cobra 主版本。
5. 依赖刷新后必须运行生成项目矩阵验证。
```

## 推荐流程

1. 生成临时项目：

```bash
go run ./cmd/gos new .tmp/deps-clean --module=example.com/deps-clean --template=api-clean --force
go run ./cmd/gos new .tmp/deps-clean-otel --module=example.com/deps-clean-otel --template=api-clean --with-otel --force
go run ./cmd/gos new .tmp/deps-minimal --module=example.com/deps-minimal --template=api-minimal --force
go run ./cmd/gos new .tmp/deps-minimal-otel --module=example.com/deps-minimal-otel --template=api-minimal --with-otel --force
```

2. 在每个临时项目中运行：

```bash
go mod tidy
go test ./...
go build ./cmd/api
```

3. 将整理后的依赖同步回模板：

```text
1. 默认项目中的 go.mod/go.sum 内容同步到无条件模板区域。
2. OTEL 项目比默认项目新增的依赖同步到 {{ if .WithOpenTelemetry }} 条件块。
3. 不要把 OTEL-only 依赖泄漏到默认模板。
4. 保持 api-clean 和 api-minimal 的 OTEL 条件块版本一致，除非有明确差异。
5. api-clean 独有的 database/sql tracing 依赖只同步到 api-clean 的 OTEL 条件块。
```

4. 在脚手架仓库运行：

```bash
go test ./...
```

当前 `internal/scaffold` 测试会生成四种项目组合，并分别运行 `go test ./...` 与 `go build ./cmd/api`。

## 检查清单

```text
1. 默认 api-clean go.mod 不包含 go.opentelemetry.io。
2. 默认 api-minimal go.mod 不包含 go.opentelemetry.io。
3. api-clean --with-otel 项目包含 otelsql、otelhttp、otel、otlptracehttp、otel/sdk。
4. api-minimal --with-otel 项目包含 otelhttp、otel、otlptracehttp、otel/sdk，不包含 otelsql。
5. go.sum.tmpl 包含所有必要校验和。
6. go test ./... 通过。
7. 文档中的依赖版本与模板一致。
```

## 后续自动化方向

```text
1. 增加专门的依赖刷新脚本，自动 diff 默认项目和 OTEL 项目依赖。
2. 在 CI 中保留生成项目矩阵验证。
3. 为模板依赖版本增加集中记录，减少 README、模板和文档之间的漂移。
```
