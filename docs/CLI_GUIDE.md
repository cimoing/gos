# gos CLI 使用指南

本文从基础到深入说明 `gos` 的使用方式。它关注“如何生成项目和代码”；生成出来的业务代码如何运行、扩展和维护，请继续阅读 `docs/GENERATED_PROJECT_GUIDE.md`。

## 1. 基本概念

`gos` 是一个 Go 后端项目脚手架 CLI。它的职责是生成可读、可改、可测试的项目结构和常用代码骨架，而不是在运行时接管业务代码。

`gos` 自身基于 Cobra 组织顶层命令；生成项目的 `cmd/api` 入口同样基于 Cobra，并支持 `serve`、`schedule`、`queue` 和由 `gos make:command` 生成的自定义命令。

当前支持的能力：

```bash
gos new <project> [--module=<module>] [--template=api-clean|api-minimal] [--with-otel] [--force] [--dry-run]
gos make:usecase <module>/<action> [--force] [--dry-run]
gos make:handler <module> [--module=<module-path>] [--register] [--openapi] [--force] [--dry-run]
gos make:model <module> [--fields=name:string[:unique,nullable,size=320,default=value,sql=TEXT,json=name]] [--openapi] [--force] [--dry-run]
gos make:repository <module> [--module=<module-path>] [--db=mysql] [--table=<table>] [--fields=name:string[:unique,nullable,size=320,default=value,sql=TEXT,json=name]] [--with-migration] [--migration-dir=migrations] [--register] [--openapi] [--force] [--dry-run]
gos make:migration <name> [--dir=migrations] [--force] [--dry-run]
gos make:test <usecase|handler|repository> <name> [--module=<module-path>] [--force] [--dry-run]
gos make:command <name> [--module=<module-path>] [--register] [--force] [--dry-run]
gos version
gos completion <bash|zsh|fish|powershell>
```

## 2. 安装与运行

在脚手架仓库中可以直接运行：

```bash
go run ./cmd/gos help
go run ./cmd/gos version
go run ./cmd/gos completion bash
```

也可以先构建本地二进制：

```bash
go build -o bin/gos ./cmd/gos
```

Windows PowerShell 中可以这样调用：

```powershell
.\bin\gos.exe help
```

如果系统没有 `make` 命令，直接使用 `go`、`docker compose` 和 `gos` 命令即可。本项目不要求本地必须安装 `make`。

## 3. 快速开始

创建一个 API 项目：

```bash
go run ./cmd/gos new myapp --module=example.com/myapp
```

进入生成项目：

```bash
cd myapp
go test ./...
go run ./cmd/api
go run ./cmd/api schedule
go run ./cmd/api queue
```

默认健康检查接口：

```bash
curl http://127.0.0.1:8080/healthz
```

生成一个完整业务模块的常见顺序：

```bash
gos make:usecase order/create
gos make:handler order --register --openapi
gos make:repository order --fields=number:string:unique,size=64,total:int64,paid:bool,created_at:time:default=now --with-migration --register
gos make:command sync-orders --register
go test ./...
go build ./cmd/api
```

## 4. new 命令

`gos new` 用于创建新项目。

```bash
gos new myapp --module=example.com/myapp --template=api-clean
```

参数说明：

```text
--module      生成项目的 Go module path。未指定时默认使用项目名。
--template    项目模板。可选 api-clean 或 api-minimal，默认 api-clean。
--with-otel   生成可选 OpenTelemetry tracing 支持，默认不启用。
--dry-run     只展示将要生成的文件，不写入磁盘。
--force       覆盖已存在文件。谨慎使用。
```

模板选择：

```text
api-clean     默认模板，包含分层 HTTP、配置、响应、错误映射、中间件、MySQL、事务、Docker 和 CI。
api-minimal   极简 HTTP 模板，只包含 cmd/api、环境配置、健康检查路由和基础测试。
```

快速生成极简项目：

```bash
gos new tiny-api --module=example.com/tiny-api --template=api-minimal
```

生成带 OpenTelemetry 支持的项目：

```bash
gos new traced-api --module=example.com/traced-api --with-otel
```

`--with-otel` 会生成 `internal/observability/otel.go`，引入 `otelhttp` HTTP tracing middleware，并增加 OTLP HTTP trace exporter 依赖。运行时仍由环境变量控制，默认 `OTEL_ENABLED=false`。

更完整的配置、代码落点和验证方式见 `docs/OPEN_TELEMETRY.md`。

建议第一次生成前使用：

```bash
gos new myapp --module=example.com/myapp --dry-run
```

确认文件列表符合预期后再去掉 `--dry-run`。

## 5. make:usecase

`make:usecase` 生成应用用例和单元测试骨架。

```bash
gos make:usecase user/register
```

生成文件：

```text
internal/usecase/user/register.go
internal/usecase/user/register_test.go
```

生成代码默认包含：

```text
1. Input DTO
2. Output DTO
3. Usecase struct
4. NewXUsecase 构造函数
5. Execute(ctx, input) 方法
6. 基础单元测试
```

最佳实践：

```text
1. Usecase 负责业务流程编排，不直接处理 HTTP 请求。
2. Usecase 的依赖用接口表达，接口定义在使用方。
3. 事务边界放在 Usecase，而不是 Repository。
4. 测试中使用 fake 或 mock 隔离数据库、Redis、外部 API。
```

## 6. make:handler

`make:handler` 生成标准库 `net/http` Handler 和测试骨架。

```bash
gos make:handler user
```

生成文件：

```text
internal/interfaces/http/handler/user_handler.go
internal/interfaces/http/handler/user_handler_test.go
```

自动注册到路由：

```bash
gos make:handler user --register
```

同时追加 OpenAPI path：

```bash
gos make:handler user --register --openapi
```

说明：

```text
1. --register 只在识别 api-clean 标准 router.go 结构时自动更新。
2. 非标准 router.go 会降级为 skipped，并提示手动注册代码。
3. --openapi 只在识别标准 api/openapi.yaml 结构时追加 path，并生成 tag、列表/创建响应、创建 requestBody、CreateXRequest schema 和标准错误响应引用。
4. Handler 层只做协议适配、参数解析、调用 Usecase、响应转换。
```

最佳实践：

```text
1. Handler 不写核心业务逻辑。
2. Handler 不直接拼 SQL。
3. Handler 中只返回统一响应结构。
4. 错误统一交给 httperror 映射。
5. 新接口先更新 api/openapi.yaml，再实现 Handler。
```

## 7. make:model

`make:model` 生成 Domain Entity。

```bash
gos make:model invoice --fields=number:string:json=invoice_number,total:int64,created_at:time --openapi
```

生成文件：

```text
internal/domain/invoice/entity.go
api/openapi.yaml（使用 --openapi 时更新）
```

适用场景：

```text
1. 只需要领域实体，暂时不生成 Repository。
2. 先建模，再逐步补充持久化实现。
3. 已有数据库访问方式，不希望使用当前 MySQL Repository 模板。
4. 使用 --openapi 时，同步向 api/openapi.yaml 的 components.schemas 追加实体 schema。
```

## 8. make:repository

`make:repository` 生成 Domain Repository 契约、MySQL Repository 实现、测试骨架，并可选生成迁移文件和注册到依赖组装。

```bash
gos make:repository invoice --fields=number:string:unique,size=64,total:int64,paid:bool,created_at:time:default=now --with-migration --register --openapi
```

生成文件：

```text
internal/domain/invoice/entity.go
internal/domain/invoice/repository.go
internal/infrastructure/persistence/mysql/invoice_repository.go
internal/infrastructure/persistence/mysql/invoice_repository_test.go
internal/infrastructure/persistence/mysql/invoice_repository_integration_test.go
migrations/<timestamp>_create_invoices_table.up.sql
migrations/<timestamp>_create_invoices_table.down.sql
internal/app/assembly.go
api/openapi.yaml（使用 --openapi 时更新）
```

参数说明：

```text
--db              当前支持 mysql。
--table           指定表名。不指定时由模块名推导，例如 category -> categories。
--fields          字段 DSL。
--with-migration  同步生成 up/down SQL 文件。
--migration-dir   迁移目录，默认 migrations。
--register        自动注册到 internal/app/assembly.go。
--openapi         根据 --fields 向 api/openapi.yaml 追加实体 schema。
```

### 8.1 字段 DSL

基础格式：

```text
name:type
```

多个字段用逗号分隔：

```bash
--fields=name:string,age:int,created_at:time
```

支持类型：

```text
string -> string / VARCHAR(255)
int    -> int / INT
int64  -> int64 / BIGINT
bool   -> bool / BOOLEAN
time   -> time.Time / TIMESTAMP
```

支持选项：

```text
nullable       允许 NULL
required       明确 NOT NULL
unique         生成唯一索引
index          生成普通索引
size=N         string 字段使用 VARCHAR(N)
default=value  生成默认值
sql=TYPE       指定 SQL 类型，例如 DECIMAL(10,2)
json=name      指定 JSON 标签名
```

`--openapi` 会复用字段 DSL：

```text
1. string -> type: string，size=N 会生成 maxLength。
2. int -> type: integer, format: int32。
3. int64 -> type: integer, format: int64。
4. bool -> type: boolean。
5. time -> type: string, format: date-time。
6. nullable 会生成 nullable: true，并从 required 列表中移除。
7. json=<name> 会作为 schema property 名。
```

示例：

```bash
gos make:repository customer --fields=email:string:unique,size=320,json=email_address,age:int:default=18,deleted_at:time:nullable,index,balance:int64:sql=BIGINT
```

注意：

```text
1. id 是保留字段，不能出现在 --fields 中。
2. default=null 只能用于 nullable 字段。
3. sql= 只允许安全字符集合，避免把任意 SQL 注入迁移模板。
4. 生成的 Repository 是基础 CRUD 骨架，复杂查询应按业务显式添加方法。
```

## 9. make:migration

生成空迁移文件：

```bash
gos make:migration create_users_table
```

生成：

```text
migrations/<timestamp>_create_users_table.up.sql
migrations/<timestamp>_create_users_table.down.sql
```

最佳实践：

```text
1. up.sql 只写正向变更。
2. down.sql 写可回滚变更。
3. 执行迁移使用生成项目内置的 `go run ./cmd/api migrate up` 和 `go run ./cmd/api migrate down`。
4. 不要把不可逆的数据修复伪装成普通结构迁移。
```

## 10. 生成项目内 migrate

执行正向迁移：

```bash
DB_DSN="user:pass@tcp(127.0.0.1:3306)/app?parseTime=true" go run ./cmd/api migrate up
```

回滚最近 1 个迁移：

```bash
go run ./cmd/api migrate down --dsn="user:pass@tcp(127.0.0.1:3306)/app?parseTime=true"
```

回滚多个迁移：

```bash
go run ./cmd/api migrate down --steps=3
go run ./cmd/api migrate down --all
```

说明：

```text
1. 执行前会自动创建 schema_migrations 记录表。
2. up 只执行记录表中不存在的版本。
3. down 按已执行版本倒序查找 .down.sql 并删除对应记录。
4. 可通过 --table=<name> 自定义迁移记录表名。
```

## 11. make:test

单独补测试骨架：

```bash
gos make:test usecase order/create
gos make:test handler order
gos make:test repository order
```

适用于：

```text
1. 先手写了实现，后补测试骨架。
2. 旧代码逐步迁移到脚手架推荐结构。
3. 只需要测试模板，不需要重新生成主代码。
```

## 12. make:command

`make:command` 生成可从 `cmd/api` 执行的 Cobra 命令脚本。

```bash
gos make:command sync-orders
```

生成文件：

```text
internal/command/sync_orders.go
internal/command/sync_orders_test.go
```

使用 `--register` 会自动注册到标准 `cmd/api/main.go` 的 Cobra root command：

```bash
gos make:command sync-orders --register
go run ./cmd/api sync-orders
```

生成项目默认内置子命令：

```text
serve       启动 HTTP API。go run ./cmd/api 默认等同 serve。
schedule    启动定时任务 worker，默认通过 internal/worker.Scheduler 管理生命周期。
queue       启动队列消费 worker，默认通过 internal/worker.QueueWorker 管理生命周期。
```

注意：

```text
1. --register 只在识别标准 Cobra cmd/api/main.go marker 时自动更新。
2. 命令名会规范化，例如 sync_orders、sync-orders、sync/orders 都会注册为 sync-orders。
3. 生成命令默认返回 *cobra.Command，只包含 context 检查和日志，占位业务逻辑应手动补充。
```

## 13. version

查看当前 `gos` 版本信息：

```bash
gos version
```

输出包含：

```text
gos <version>
commit <commit>
built <build-date>
```

本地开发默认显示 `dev`、`none`、`unknown`。发布构建时可以通过 ldflags 注入：

```bash
go build -ldflags "-X github.com/cimoing/gos/internal/command.Version=v0.1.0 -X github.com/cimoing/gos/internal/command.Commit=abc1234 -X github.com/cimoing/gos/internal/command.BuildDate=2026-06-09T00:00:00Z" -o bin/gos ./cmd/gos
```

## 14. completion

生成 shell completion：

```bash
gos completion bash
gos completion zsh
gos completion fish
gos completion powershell
```

常见用法：

```bash
gos completion bash > gos.bash
gos completion zsh > _gos
gos completion fish > gos.fish
```

PowerShell：

```powershell
gos completion powershell > gos.ps1
```

发布流程建议见 `docs/RELEASE.md`。

## 15. dry-run、force 与冲突处理

`--dry-run` 不写入磁盘，只展示计划生成的文件：

```bash
gos make:repository invoice --fields=number:string --dry-run
```

`--force` 会覆盖已存在文件：

```bash
gos make:handler invoice --force
```

建议：

```text
1. 首次在已有项目中使用生成命令时先 dry-run。
2. 对业务代码谨慎使用 --force。
3. 自动注册失败时优先按提示手动注册，不要为了注册强行覆盖非标准文件。
4. 生成代码可以自由修改，后续再次生成前先确认差异。
```

## 16. 推荐开发流程

新增一个业务能力时，推荐顺序：

```text
1. 更新 api/openapi.yaml，明确接口契约。
2. gos make:usecase <module>/<action>。
3. 先完善 Usecase 测试。
4. 定义 Usecase 需要的 Repository 或外部能力接口。
5. gos make:repository <module> --with-migration --register。
6. 在 Usecase 中注入 Repository 和 TxManager。
7. gos make:handler <module> --register --openapi。
8. Handler 调用 Usecase，统一返回 response。
9. 需要后台入口时 gos make:command <name> --register。
10. go test ./...。
11. 有数据库时运行 integration 测试。
```

如果是数据模型优先：

```text
1. gos make:model <module> --fields=...
2. 补充领域规则和构造函数。
3. gos make:repository <module> --fields=... --with-migration。
4. 按业务补充 Repository 方法。
```

## 17. Docker 与集成测试

生成项目包含测试数据库 Compose 文件：

```bash
docker compose -f deployments/docker/docker-compose.test.yml up -d
```

运行 Repository 集成测试：

```bash
TEST_DATABASE_DSN='root:password@tcp(127.0.0.1:3307)/myapp_test?parseTime=true' go test -tags=integration ./internal/infrastructure/persistence/mysql
```

Windows PowerShell：

```powershell
$env:TEST_DATABASE_DSN='root:password@tcp(127.0.0.1:3307)/myapp_test?parseTime=true'
go test -tags=integration ./internal/infrastructure/persistence/mysql
```

## 18. 常见问题

`make` 不存在：

```text
直接使用 go、gos 和 docker compose 命令。Makefile 只是便利入口。
```

`--register` 显示 skipped：

```text
说明目标文件不是标准 api-clean marker 结构。生成的主文件仍然有效，按命令输出提示手动注册即可。
```

生成项目启动时报 database is not configured：

```text
通常是业务代码调用了 TxManager 或 Repository，但 DB_DSN 为空。需要设置 DB_DRIVER=mysql 和 DB_DSN，或在无数据库模式下避免调用数据库能力。
```

生成项目启动时报 parse XXX as bool/int：

```text
说明环境变量格式不合法。布尔值使用 true/false，整数值使用十进制数字，例如 REDIS_DB=0。
```

生成项目启动时报 parse XXX as duration：

```text
说明 duration 环境变量格式不合法。使用 Go duration 格式，例如 5s、30s、1m。
```

Repository 集成测试被跳过：

```text
这是预期行为。设置 TEST_DATABASE_DSN 后再使用 -tags=integration 运行。
```

## 19. CLI 最佳实践

```text
1. 生成器用于启动代码，不用于长期覆盖业务代码。
2. 先 dry-run，再正式生成。
3. 先契约和测试，再实现业务逻辑。
4. Usecase 控制事务边界，Repository 只做数据访问。
5. 自动注册是便利能力，不应替代对 app/router 组装逻辑的理解。
6. 字段 DSL 适合基础 CRUD，复杂表结构应手动审查迁移 SQL。
7. 每次生成后运行 gofmt、go test、go build。
```
