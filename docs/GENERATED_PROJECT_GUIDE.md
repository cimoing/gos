# 生成项目代码使用指南

本文说明 `gos new` 生成的 `api-clean` 项目如何使用、扩展和维护。它不仅说明命令，更重点解释生成出来的代码应该怎么接着写。

## 1. 项目结构总览

生成项目默认结构：

```text
cmd/api                          API 服务入口
api/openapi.yaml                 HTTP API 契约
internal/command                 gos 生成的命令脚本
internal/app                     应用组装和生命周期
internal/config                  环境变量配置读取
internal/domain                  领域实体和领域接口
internal/usecase                 应用用例
internal/interfaces/http         HTTP 路由、Handler、中间件、错误映射
internal/infrastructure/cache    缓存接口和 memory/file/memcache/redis 实现
internal/infrastructure/database database/sql 连接和事务管理
internal/infrastructure/lock     Redis 分布式锁实现
internal/infrastructure/persistence/mysql
                                  MySQL Repository 实现
internal/infrastructure/redisclient
                                  Redis client 构造
internal/logging                 slog 日志初始化和可选 trace/span 字段注入
internal/worker                  schedule/queue 后台 worker 生命周期骨架
internal/pkg/apperror            应用错误
internal/pkg/response            统一响应结构
deployments/docker               Dockerfile 和 Compose
migrations                       SQL 迁移文件
```

依赖方向：

```text
HTTP Handler -> Usecase -> Domain
MySQL Repository -> Domain
App -> 组装 Config、DB、Repository、Usecase、Handler
```

核心原则：

```text
1. Handler 负责协议适配。
2. Usecase 负责业务流程和事务边界。
3. Domain 表达业务实体、业务规则和接口。
4. Infrastructure 实现数据库、队列、外部 API 等细节。
5. App 层手写依赖组装，保持显式可读。
```

## 2. 启动项目

直接运行：

```bash
go run ./cmd/api
go run ./cmd/api serve
```

`go run ./cmd/api` 默认等同于 `serve`。

后台入口：

```bash
go run ./cmd/api schedule
go run ./cmd/api queue
```

生成项目的命令入口基于 Cobra 实现。`schedule` 和 `queue` 复用 `internal/worker` 中的后台 worker 骨架，具备统一的启动/停止日志、周期执行、错误记录和 panic recover。业务占位逻辑分别在 `cmd/api/main.go` 的 `runScheduledJobs` 与 `processQueueOnce` 中扩展。

构建：

```bash
go build ./cmd/api
```

测试：

```bash
go test ./...
go vet ./...
```

Docker 本地环境：

```bash
docker compose -f deployments/docker/docker-compose.yml up --build
```

健康检查：

```bash
curl http://127.0.0.1:8080/healthz
```

## 2.1 命令脚本

新增可执行 Cobra 命令：

```bash
gos make:command sync-orders --register
go run ./cmd/api sync-orders
```

生成文件：

```text
internal/command/sync_orders.go
internal/command/sync_orders_test.go
```

`--register` 会更新标准 `cmd/api/main.go` 中的 Cobra root command。若项目已经重写过入口文件，自动注册可能跳过，此时按命令输出提示手动添加 import 和 `rootCmd.AddCommand(...)` 即可。

## 2.2 后台 Worker

后台 worker 骨架位于：

```text
internal/worker
```

默认包含：

```text
Scheduler     周期执行一组命名任务。
QueueWorker   周期调用队列消费占位函数。
runSafely     捕获 panic 并转换为错误日志。
```

`schedule` 默认使用 `worker.NewScheduler`，`queue` 默认使用 `worker.NewQueueWorker`。接入真实任务时，优先保留 `internal/worker` 的生命周期控制，在 `runScheduledJobs`、`processQueueOnce` 或独立业务包中替换具体逻辑。

## 3. 配置使用

配置从环境变量读取到 `internal/config.Config`。业务代码不要直接读取环境变量，应通过构造函数传入配置或依赖。

配置解析会在启动阶段暴露明显错误。布尔、整数和 duration 配置不会静默吞掉非法值，例如 `REDIS_DB=abc`、`DB_ENABLE_NESTED_TRANSACTION=maybe`、`HTTP_READ_TIMEOUT=abc` 或 `OTEL_ENABLED=maybe` 会导致 `config.Load` 返回错误。

常用配置：

```text
APP_NAME
APP_ENV
HTTP_ADDR
HTTP_READ_HEADER_TIMEOUT
HTTP_READ_TIMEOUT
HTTP_WRITE_TIMEOUT
HTTP_IDLE_TIMEOUT
HTTP_MAX_HEADER_BYTES
HTTP_MAX_BODY_BYTES
CORS_ALLOWED_ORIGINS
CORS_ALLOWED_METHODS
CORS_ALLOWED_HEADERS
CORS_ALLOW_CREDENTIALS
CORS_MAX_AGE
DB_DRIVER
DB_DSN
DB_ENABLE_NESTED_TRANSACTION
REDIS_ADDR
REDIS_PASSWORD
REDIS_DB
CACHE_BACKEND
CACHE_FILE_DIR
CACHE_MEMCACHE_SERVERS
CACHE_DEFAULT_TTL
LOCK_REDIS_KEY_PREFIX
LOCK_DEFAULT_TTL
LOG_LEVEL
OTEL_ENABLED
OTEL_SERVICE_NAME
OTEL_EXPORTER_OTLP_ENDPOINT
OTEL_EXPORTER_OTLP_INSECURE
```

HTTP server 默认生产值：

```text
HTTP_READ_HEADER_TIMEOUT=5s
HTTP_READ_TIMEOUT=15s
HTTP_WRITE_TIMEOUT=30s
HTTP_IDLE_TIMEOUT=60s
HTTP_MAX_HEADER_BYTES=1048576
HTTP_MAX_BODY_BYTES=10485760
```

这些配置会进入 `http.Server` 或 router wrapper，分别控制请求头读取、请求体读取、响应写入、空闲连接保持、最大请求头大小和最大请求体大小。`HTTP_MAX_BODY_BYTES=0` 可关闭请求体大小限制。

CORS 默认配置：

```text
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Authorization,Content-Type,X-Request-ID
CORS_ALLOW_CREDENTIALS=false
CORS_MAX_AGE=600
```

生产环境建议把 `CORS_ALLOWED_ORIGINS` 设置为明确域名列表。若启用 `CORS_ALLOW_CREDENTIALS=true`，不要使用宽泛来源。

数据库配置：

```bash
DB_DRIVER=mysql
DB_DSN='root:password@tcp(127.0.0.1:3306)/myapp?parseTime=true'
DB_ENABLE_NESTED_TRANSACTION=false
```

`DB_DSN` 为空时不会建立数据库连接，便于零依赖启动和运行纯单元测试。一旦业务路径调用 Repository 或 TxManager，就应该配置数据库。

缓存配置：

```bash
CACHE_BACKEND=memory
CACHE_FILE_DIR=.cache
CACHE_MEMCACHE_SERVERS=127.0.0.1:11211
CACHE_DEFAULT_TTL=5m
```

`internal/infrastructure/cache` 提供统一 `Store` 接口，支持 `memory`、`file`、`memcache`、`redis` 四种后端。`memory` 适合本地开发和单进程临时缓存；`file` 使用 key hash 落盘，适合轻量持久化；`memcache` 和 `redis` 适合多实例共享缓存。

Redis 分布式锁配置：

```bash
LOCK_REDIS_KEY_PREFIX=myapp:lock:
LOCK_DEFAULT_TTL=30s
```

`internal/infrastructure/lock` 提供 Redis `Locker`，使用 `SET NX EX/PX` 语义获取锁，并用 Lua 脚本保证 `Release` 与 `Refresh` 只作用于当前 token。锁适合短临界区、定时任务互斥和队列消费者互斥；长任务应主动 `Refresh` 或拆分执行粒度。

使用 `gos new --with-otel` 生成的项目会包含 OpenTelemetry tracing 支持。默认不开启：

```bash
OTEL_ENABLED=false
OTEL_SERVICE_NAME=myapp
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
OTEL_EXPORTER_OTLP_INSECURE=true
```

开启后，启动阶段会初始化 OTLP HTTP trace exporter，HTTP router 会通过 `otelhttp` 生成请求 span。`api-clean --with-otel` 还会通过 `github.com/XSAM/otelsql` 包裹 database/sql，让 Repository 和 TxManager 的数据库操作进入同一条 trace。未使用 `--with-otel` 时不会生成相关依赖和代码。

OpenTelemetry 的生成代码落点、运行配置、Collector 示例和排障清单见 `docs/OPEN_TELEMETRY.md`。

最佳实践：

```text
1. 启动阶段读取配置。
2. 配置通过结构体传递。
3. Usecase 不读取环境变量。
4. 非字符串配置使用严格解析，非法值应尽早失败。
5. 测试中显式构造配置，不依赖机器环境。
```

## 4. 应用组装

依赖组装入口在：

```text
internal/app/assembly.go
```

默认生成：

```go
db, err := database.Open(ctx, cfg.Database)
transactions := database.NewTxManager(db, database.TxOptions{
	EnableNestedTransaction: cfg.Database.EnableNestedTransaction,
})
redisClient := redisclient.New(cfg.Redis)
cacheStore, err := cache.NewStore(cache.Options{
	Backend:         cfg.Cache.Backend,
	FileDir:         cfg.Cache.FileDir,
	MemcacheServers: cfg.Cache.MemcacheServers,
	RedisClient:     redisClient,
	DefaultTTL:      cfg.Cache.DefaultTTL,
})
locker := lock.NewRedisLocker(redisClient, lock.Options{
	KeyPrefix:  cfg.Lock.RedisKeyPrefix,
	DefaultTTL: cfg.Lock.DefaultTTL,
})
```

当使用：

```bash
gos make:repository invoice --register
```

会向 `Dependencies` 自动补充：

```go
InvoiceRepository *mysqlrepo.InvoiceRepository
```

并在 `BuildDependencies` 中构造：

```go
invoiceRepository := mysqlrepo.NewInvoiceRepository(db)
```

后续接入 Usecase 时，推荐继续手写显式组装：

```go
createInvoice := invoiceusecase.NewCreateUsecase(
	deps.InvoiceRepository,
	deps.Transactions,
)
invoiceHandler := handler.NewInvoiceHandler(createInvoice)
```

最佳实践：

```text
1. 构造函数注入优先。
2. 不使用运行时服务容器。
3. App 层可以依赖所有层，用于最终组装。
4. Usecase 不 import MySQL Repository。
5. Handler 不直接 new Repository。
6. 需要缓存或分布式锁时，通过 `Dependencies.Cache`、`Dependencies.Locker` 注入 Usecase，不在业务代码中直接创建客户端。
```

## 5. HTTP 层

路由入口：

```text
internal/interfaces/http/router.go
```

默认路由：

```text
GET /healthz
```

默认中间件链：

```text
RequestID
Recover
CORS
AccessLog
Timeout
```

生成 Handler：

```bash
gos make:handler invoice --register --openapi
```

生成代码默认包含：

```go
func (h *InvoiceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /invoices", h.List)
}
```

Handler 推荐职责：

```text
1. 读取 path/query/body。
2. 做轻量协议层校验。
3. 调用 Usecase。
4. 将结果转换为 response.JSON。
5. 将错误交给 httperror 映射。
```

Handler 不推荐：

```text
1. 直接操作数据库。
2. 写复杂业务规则。
3. 开启事务。
4. 返回领域实体的内部敏感字段。
```

## 6. 统一响应与错误

统一响应包：

```text
internal/pkg/response
```

典型成功响应：

```go
response.JSON(w, http.StatusOK, response.Success(data))
```

错误类型：

```text
internal/pkg/apperror
```

HTTP 错误映射：

```text
internal/interfaces/http/httperror
```

推荐方式：

```go
if err != nil {
	httperror.Write(w, err)
	return
}
```

最佳实践：

```text
1. Domain/Usecase 返回业务错误，不返回 HTTP 状态码。
2. HTTP 层负责把错误映射成状态码和响应体。
3. 不把数据库原始错误直接暴露给客户端。
4. 日志可以记录内部错误，响应体保持稳定结构。
```

## 7. Usecase 层

Usecase 位于：

```text
internal/usecase/<module>
```

生成：

```bash
gos make:usecase invoice/create
```

推荐结构：

```go
type CreateInput struct {
	Number string
	Total  int64
}

type CreateOutput struct {
	ID int64
}

type InvoiceRepository interface {
	Save(ctx context.Context, invoice *invoice.Invoice) error
}

type CreateUsecase struct {
	invoices InvoiceRepository
	tx       interface {
		WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
	}
}
```

Usecase 执行：

```go
func (uc *CreateUsecase) Execute(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	inv := &invoice.Invoice{
		Number: input.Number,
		Total:  input.Total,
	}

	if err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		return uc.invoices.Save(ctx, inv)
	}); err != nil {
		return nil, err
	}

	return &CreateOutput{ID: inv.ID}, nil
}
```

最佳实践：

```text
1. Usecase 输入输出使用 DTO，不直接绑定 HTTP request/response。
2. 依赖接口定义在使用方，便于测试。
3. 事务边界在 Usecase。
4. 单元测试使用 fake Repository。
5. 复杂业务规则下沉到 Domain。
```

## 8. Domain 层

Domain 位于：

```text
internal/domain/<module>
```

生成 Entity：

```bash
gos make:model invoice --fields=number:string,total:int64,created_at:time
```

或由 Repository 命令一并生成：

```bash
gos make:repository invoice --fields=number:string,total:int64 --with-migration
```

推荐在 Entity 上补充构造函数和业务校验：

```go
func NewInvoice(number string, total int64) (*Invoice, error) {
	if number == "" {
		return nil, ErrInvalidNumber
	}
	if total <= 0 {
		return nil, ErrInvalidTotal
	}
	return &Invoice{
		Number: number,
		Total:  total,
	}, nil
}
```

最佳实践：

```text
1. Domain 不 import net/http。
2. Domain 不 import database/sql。
3. Domain 不读取环境变量。
4. 领域错误使用稳定变量，便于 errors.Is。
5. 领域实体不要塞入协议层字段。
```

## 9. Repository 层

MySQL Repository 位于：

```text
internal/infrastructure/persistence/mysql
```

生成：

```bash
gos make:repository invoice --fields=number:string:unique,size=64,total:int64,paid:bool,created_at:time:default=now --with-migration --register
```

生成方法：

```text
FindByID(ctx, id)
Save(ctx, entity)
DeleteByID(ctx, id)
```

`Save` 语义：

```text
1. entity.ID == 0 时 INSERT。
2. entity.ID != 0 时 UPDATE。
3. INSERT 成功后回填 entity.ID。
```

Repository 内部通过：

```go
executor := database.ExecutorFromContext(ctx, r.db)
```

自动判断当前 context 中是否已有事务。有事务时使用 `*sql.Tx`，没有事务时使用 `*sql.DB`。

最佳实践：

```text
1. Repository 只做持久化，不控制事务边界。
2. 简单 CRUD 可以使用生成代码。
3. 复杂查询手写明确方法，例如 FindPendingByCustomerID。
4. SQL 迁移文件必须人工审查后再进入生产。
5. 集成测试覆盖真实数据库行为。
```

## 10. 事务管理

事务管理位于：

```text
internal/infrastructure/database/transaction.go
```

Usecase 中使用：

```go
err := txManager.WithinTx(ctx, func(ctx context.Context) error {
	if err := orderRepo.Save(ctx, order); err != nil {
		return err
	}
	if err := inventoryRepo.Save(ctx, inventory); err != nil {
		return err
	}
	return nil
})
```

Repository 不需要接收 `*sql.Tx` 参数，因为它会从 `ctx` 中获取当前事务 executor。

### 10.1 默认嵌套行为

默认配置：

```text
DB_ENABLE_NESTED_TRANSACTION=false
```

行为：

```text
1. 外层 WithinTx 开启数据库事务。
2. 内层 WithinTx 发现 ctx 中已有事务。
3. 内层不会再次开启事务，也不会创建 savepoint。
4. 所有写入由最外层 commit 或 rollback。
```

### 10.2 开启 savepoint 嵌套事务

配置：

```text
DB_ENABLE_NESTED_TRANSACTION=true
```

行为：

```text
1. 外层 WithinTx 开启数据库事务。
2. 内层 WithinTx 创建 SAVEPOINT。
3. 内层成功时 RELEASE SAVEPOINT。
4. 内层失败时 ROLLBACK TO SAVEPOINT。
5. 最终 commit 仍由最外层事务决定。
```

适合场景：

```text
1. 一个大事务中有可局部回滚的子步骤。
2. 子步骤失败后仍希望继续执行后续逻辑。
3. 团队明确理解 savepoint 与真正独立事务的区别。
```

不适合场景：

```text
1. 希望内层成功后立即永久提交。
2. 业务逻辑不清楚哪些错误可以局部回滚。
3. 需要跨数据库或跨服务事务。
```

最佳实践：

```text
1. 默认关闭嵌套事务。
2. 优先保持事务边界简单。
3. 开启 savepoint 后要为局部回滚路径补测试。
4. 不要在 Handler 或 Repository 中开启事务。
```

## 11. 数据库迁移

生成迁移：

```bash
gos make:migration create_invoices_table
```

或 Repository 同步生成：

```bash
gos make:repository invoice --fields=number:string,total:int64 --with-migration
```

推荐流程：

```text
1. 生成迁移文件。
2. 人工审查 SQL 类型、索引、默认值。
3. 在本地测试库执行。
4. 运行 Repository integration test。
5. 再提交代码。
```

注意：

```text
脚手架生成迁移文件，但不强制绑定具体迁移工具。团队可以选择 golang-migrate、Atlas 或内部发布系统。
```

## 12. 测试体系

普通测试：

```bash
go test ./...
```

Repository 集成测试：

```bash
docker compose -f deployments/docker/docker-compose.test.yml up -d
TEST_DATABASE_DSN='root:password@tcp(127.0.0.1:3307)/myapp_test?parseTime=true' go test -tags=integration ./internal/infrastructure/persistence/mysql
```

PowerShell：

```powershell
$env:TEST_DATABASE_DSN='root:password@tcp(127.0.0.1:3307)/myapp_test?parseTime=true'
go test -tags=integration ./internal/infrastructure/persistence/mysql
```

测试分层建议：

```text
1. Domain 规则用单元测试。
2. Usecase 用 fake 依赖做单元测试。
3. Handler 用 httptest 验证状态码和响应结构。
4. Repository 用真实数据库做 integration test。
5. 跨模块主流程再补 E2E。
```

## 13. OpenAPI 契约

契约文件：

```text
api/openapi.yaml
```

推荐流程：

```text
1. 先定义 path、request、response。
2. 再生成或编写 Handler。
3. Handler 测试对齐契约。
4. API 变更时同步更新 OpenAPI。
```

`gos make:handler <module> --openapi` 会追加基础 list/create path，并包含 tag、列表成功响应、创建 requestBody、CreateXRequest schema 和标准错误响应引用。`gos make:model <module> --openapi` 与 `gos make:repository <module> --openapi` 会根据字段 DSL 向 `components.schemas` 追加实体 schema。复杂请求字段和更细的业务错误码仍应按业务手动完善。

## 14. 日志与中间件

日志初始化位于：

```text
internal/logging
```

启动阶段会读取 `LOG_LEVEL` 创建 JSON logger，并通过 `slog.SetDefault` 设为默认 logger。支持级别：

```text
debug
info
warn
warning
error
```

使用 `--with-otel` 生成并启用 tracing 时，日志 handler 会从 `context.Context` 中读取当前 span，并自动补充 `trace_id` 和 `span_id` 字段。生成 logger 默认会对 `password`、`token`、`authorization`、`secret`、`dsn` 等常见敏感字段键做脱敏。

默认中间件：

```text
RequestID   生成或透传 X-Request-ID。
Recover     捕获 panic，避免进程直接崩溃；响应不暴露 panic 内容，日志只记录 panic 类型。
CORS        处理跨域，支持通过 CORS_* 环境变量配置。
AccessLog   记录请求方法、路径、耗时、状态。
Timeout     为请求设置超时。
```

最佳实践：

```text
1. 日志中保留 request_id。
2. 不记录密码、Token、密钥；确需记录上下文字段时使用结构化字段并依赖 logger 默认脱敏。
3. Recover 只兜底，不代替错误处理。
4. Timeout 要结合业务场景调整。
5. CORS 生产环境不要使用过宽配置。
```

## 15. CI

生成项目包含：

```text
.github/workflows/ci.yml
```

默认检查：

```text
1. gofmt
2. go vet ./...
3. go test ./...
4. go build ./cmd/api
5. 如果存在 MySQL Repository，则运行 integration test
```

建议在团队项目中继续补充：

```text
1. golangci-lint
2. OpenAPI 校验
3. Docker 镜像构建
4. 数据库迁移 dry-run
5. 安全扫描
```

## 16. 生成代码修改原则

生成代码不是黑盒，可以自由修改。推荐原则：

```text
1. 先保留生成代码的分层边界。
2. 修改前补测试。
3. 复杂业务不要塞进 Handler。
4. 复杂 SQL 不要过度抽象成通用方法。
5. app/assembly.go 是显式组装点，改动应清晰。
6. 再次运行生成命令前先用 git diff 查看本地改动。
```

## 17. 模块开发最佳实践

新增“创建发票”能力的完整示例：

```bash
gos make:repository invoice --fields=number:string:unique,size=64,total:int64,paid:bool,created_at:time:default=now --with-migration --register
gos make:usecase invoice/create
gos make:handler invoice --register --openapi
go test ./...
```

然后手动完善：

```text
1. internal/domain/invoice/entity.go 增加 NewInvoice 和领域错误。
2. internal/usecase/invoice/create.go 注入 Repository 和 TxManager。
3. internal/usecase/invoice/create_test.go 使用 fake Repository 覆盖成功和失败路径。
4. internal/interfaces/http/handler/invoice_handler.go 解析请求并调用 Usecase。
5. api/openapi.yaml 补充 requestBody、领域 schema 和业务错误码。
6. migrations/*.up.sql 人工审查字段类型和索引。
7. 运行 integration test 验证 Repository。
```

## 18. 不推荐做法

```text
1. 在 Handler 中直接写 SQL。
2. 在 Repository 中悄悄开启事务。
3. 在 Domain 中 import HTTP、SQL、Redis SDK。
4. 在 Usecase 中读取 os.Getenv。
5. 为每个 struct 机械生成接口。
6. 用 --force 覆盖已经手动改过的业务代码。
7. 把生成的基础 CRUD 当成复杂业务的最终实现。
```

## 19. 没有脚本目录时如何工作

当前模板不依赖 `scripts/*.ps1` 或 shell 脚本。常用操作直接使用：

```bash
go run ./cmd/api
go test ./...
go vet ./...
go build ./cmd/api
docker compose -f deployments/docker/docker-compose.yml up --build
docker compose -f deployments/docker/docker-compose.test.yml up -d
```

`Makefile` 只是可选便利入口；没有 `make` 的系统不受影响。
