# Go 后端脚手架项目开发计划

制定日期：2026-06-03

## 1. 计划目标

本计划基于 `README.md` 中的整体设计文档制定，目标是将项目从设计文档推进为一个可用、可测试、可迭代的 Go 后端工程脚手架。

项目优先级如下：

```text
1. 先完成可运行的 CLI 工具
2. 再完成 gos new 项目生成能力
3. 再补齐 Usecase、Handler、Repository、Migration 等代码生成器
4. 再增强模板、测试、CI、本地开发工具
5. 最后扩展多模板和高级能力
```

核心原则：

```text
1. Go 原生优先
2. 轻框架、重规范
3. 接口优先
4. 测试优先
5. 生成代码可读、可改、可测试
6. 不引入重型运行时框架
```

---

## 当前实现快照（2026-06-05）

当前项目已具备可用 MVP 能力，并已推进到工程能力增强阶段。

已实现：

```text
1. gos new <project>
2. gos make:usecase <module>/<action>
3. gos make:handler <module>，支持 --register 和 --openapi
4. gos make:model <module>，支持 --fields、--force、--dry-run
5. gos make:repository <module>，支持 --table、--fields、--with-migration、--register
6. gos make:migration <name>
7. gos make:test <usecase|handler|repository> <name>
8. api-clean 模板
9. MySQL driver 注册、database/sql 连接入口、事务管理和依赖组装
10. MySQL Repository 集成测试模板
11. 测试数据库 Docker Compose 工作流
12. GitHub Actions CI 模板
13. CORS 中间件
14. 基于 enableNestedTransaction 的嵌套事务 savepoint
15. 项目内 .tmp 临时目录用于端到端验证和 Go cache
16. CLI 使用指南 docs/CLI_GUIDE.md
17. 生成项目代码使用指南 docs/GENERATED_PROJECT_GUIDE.md
18. api-clean 生成项目 README 详细使用文档
19. 多模板发现与校验
20. api-minimal 极简 HTTP 项目模板
21. 基于 Cobra 的 cmd/api 子命令模板：serve、schedule、queue
22. 基于 Cobra 的 gos make:command <name>，支持 --register、--force、--dry-run
23. gos new --with-otel 可选 OpenTelemetry tracing 支持
24. 生成项目 internal/logging 日志初始化，LOG_LEVEL 生效
25. OpenTelemetry 启用时日志自动注入 trace_id/span_id
26. 脚手架自身 CLI 基于 Cobra 组织命令
27. api-clean/api-minimal 与 OTEL 开关组合的生成项目矩阵编译验证
28. 生成项目 bool/int 环境变量严格解析，非法配置启动即失败
29. 生成项目 internal/worker 后台任务/队列生命周期骨架
30. 生成项目 HTTP server 生产默认值支持配置化 timeout 和 MaxHeaderBytes
31. docs/CONFIG_REFERENCE.md 配置参考表
32. docs/TEMPLATE_DEPENDENCIES.md 模板依赖刷新流程
33. gos version 支持版本、commit、构建时间输出
34. 生成项目 HTTP_MAX_BODY_BYTES 请求体大小限制
35. gos completion 支持 bash、zsh、fish、powershell
36. docs/RELEASE.md 发布说明
37. --with-otel 生成 observability.NewHTTPClient/NewHTTPTransport
38. docs/LOCAL_OBSERVABILITY.md 本地可观测环境示例
39. api-clean --with-otel 使用 otelsql 支持 database/sql tracing
40. 生成项目安全默认值增强：CORS 配置化、日志敏感字段脱敏、panic 输出边界
41. OpenAPI 基础深化：复用响应组件、列表响应 schema、错误响应引用和示例
```

仍未完成：

```text
1. README/DEVELOPMENT_PLAN 旧设计段落的进一步归档整理
2. docs/OPTIMIZATION_BACKLOG.md 中记录的后续工程增强项
```

---

## 2. 版本路线图

### 2.1 v0.1.0 MVP

目标：完成最小可用版本，支持初始化一个标准 Go API 项目。

范围：

```text
1. 初始化脚手架自身工程
2. 实现 gos CLI 入口
3. 实现 gos new <project>
4. 内置 api-clean 项目模板
5. 生成基础目录结构
6. 生成 go.mod、README.md、Makefile、.gitignore、.env.example
7. 生成 cmd/api、internal/config、internal/app、internal/interfaces/http
8. 生成统一响应、基础错误处理、基础用户注册示例
9. 生成 Docker Compose 开发环境
10. 支持文件冲突检测
11. 生成后执行 gofmt
12. 基础单元测试覆盖核心生成流程
```

验收标准：

```text
1. go test ./... 通过
2. go run ./cmd/gos new myapp 可以生成项目
3. 生成后的 myapp 可以执行 go test ./...
4. 生成后的 myapp 可以执行 go run ./cmd/api
5. 文件已存在时默认不会静默覆盖
6. README 中能说明基本使用方式
```

### 2.2 v0.2.0 代码生成增强

目标：支持常用业务代码骨架生成。

范围：

```text
1. gos make:usecase <module>/<action>
2. gos make:handler <module>
3. gos make:repository <module> --db=mysql
4. gos make:model <module>
5. gos make:migration <name>
6. gos make:test <module>/<action>
7. 命名转换工具完善
8. 模板上下文完善
9. 支持 dry-run
10. 支持 --force 覆盖
```

验收标准：

```text
1. 每个 make 命令都有单元测试
2. 生成文件路径、包名、类型名符合 Go 命名习惯
3. 生成后的代码通过 gofmt
4. 冲突、覆盖、dry-run 行为明确且可测试
5. 生成 Usecase 时同步生成测试骨架
```

### 2.3 v0.3.0 工程能力增强

目标：让生成项目具备较完整的后端基础能力。

范围：

```text
1. RequestID 中间件
2. Recover 中间件
3. AccessLog 中间件
4. Timeout 中间件
5. CORS 中间件
6. 统一响应包
7. 统一错误映射示例
8. slog 日志模板
9. 优雅关闭模板
10. 事务管理接口模板
11. MySQL Repository 示例
12. Redis 配置示例
13. migrations 目录与迁移命令
```

验收标准：

```text
1. 生成项目具备清晰的分层示例
2. HTTP Handler 不包含核心业务逻辑
3. Usecase 不依赖 HTTP、Gin、数据库具体实现
4. Repository 实现位于 infrastructure 层
5. app 层负责依赖组装
6. Makefile 覆盖 run、test、vet、build、migrate、docker 命令
```

### 2.4 v0.4.0 测试与契约增强

目标：完善测试优先和接口优先工作流。

范围：

```text
1. OpenAPI 模板
2. Handler 测试模板
3. Repository 集成测试模板
4. Usecase 测试模板增强
5. 测试数据库初始化脚本
6. GitHub Actions CI 模板
7. golangci-lint 配置模板
8. 示例 E2E 测试结构
```

验收标准：

```text
1. 生成项目默认包含 api/openapi.yaml
2. 示例 Usecase 测试可直接运行
3. 示例 Handler 测试验证状态码和响应结构
4. CI 可以执行 go vet、go test、go build
5. 测试目录职责清晰
```

### 2.5 v0.5.0 多模板支持

目标：支持不同类型后端项目。

范围：

```text
1. api-basic 模板
2. grpc-service 模板
3. worker 模板
4. monolith 模板
5. --template 参数
6. 模板发现与校验
7. 模板级配置
8. 不同 HTTP 框架可选
9. 不同数据库方案可选
```

验收标准：

```text
1. gos new myapp --template=api-clean 可用
2. gos new worker-app --template=worker 可用
3. 未知模板会给出清晰错误
4. 模板之间共享通用生成能力
5. 每个模板都有最小可运行测试
```

### 2.6 v0.6.0 高级能力

目标：补充企业级项目常见能力，但保持可选。

范围：

```text
1. Wire 支持
2. Queue 模板
3. Scheduler 模板
4. Auth 模板
5. 权限模板
6. 多租户模板
7. 插件机制探索
```

验收标准：

```text
1. 高级能力均为显式开启
2. 默认模板仍保持轻量
3. 不引入重型运行时容器
4. 复杂能力有独立文档和测试
```

---

## 3. 模块开发计划

### 3.1 CLI 层

目录：

```text
cmd/gos
internal/command
```

任务：

```text
1. 初始化 CLI 入口
2. 解析子命令和参数
3. 实现 new 命令
4. 实现 make:* 命令
5. 统一命令输出格式
6. 统一错误返回
```

建议顺序：

```text
1. main.go
2. command.NewCommand
3. command.MakeUsecaseCommand
4. command.MakeHandlerCommand
5. command.MakeRepositoryCommand
6. command.MakeMigrationCommand
```

### 3.2 Generator 层

目录：

```text
internal/generator
```

任务：

```text
1. 定义 Generator 接口
2. 定义 Context
3. 定义 FileSpec
4. 加载模板
5. 渲染模板
6. 写入文件
7. 执行 gofmt
8. 返回生成结果
```

核心接口草案：

```go
type Generator interface {
	Generate(ctx context.Context, input Input) (*Result, error)
}

type FileSpec struct {
	Path     string
	Content  []byte
	Template string
}

type Result struct {
	Created []string
	Skipped []string
	Updated []string
}
```

### 3.3 Template 层

目录：

```text
internal/template
templates
```

任务：

```text
1. 使用 embed 打包内置模板
2. 使用 text/template 渲染
3. 支持模板变量
4. 支持模板路径解析
5. 支持模板存在性检查
6. 支持未来扩展外部模板
```

模板变量优先支持：

```text
ProjectName
ModuleName
ActionName
PackageName
TypeName
SnakeName
KebabName
CamelName
PascalName
Timestamp
HTTPFramework
DatabaseDriver
```

### 3.4 Naming 层

目录：

```text
internal/naming
```

任务：

```text
1. snake_case
2. kebab-case
3. camelCase
4. PascalCase
5. 复数转换
6. 模块路径解析
```

验收重点：

```text
1. user/register -> module=user action=register
2. user_profile -> UserProfile
3. user-profile -> UserProfile
4. users -> User 或 Users 的行为清晰可控
```

### 3.5 Filesystem 层

目录：

```text
internal/filesystem
```

任务：

```text
1. 创建目录
2. 写入文件
3. 冲突检测
4. --force 覆盖
5. dry-run
6. gofmt 格式化
7. 路径安全检查
```

验收重点：

```text
1. 默认不覆盖已有文件
2. --force 明确覆盖
3. dry-run 不写入磁盘
4. 不允许写出目标项目目录
```

### 3.6 Project 层

目录：

```text
internal/project
```

任务：

```text
1. 检测当前是否为 Go 项目
2. 读取脚手架配置
3. 推断 module path
4. 检查目标目录
5. 支持项目级配置文件
```

建议配置文件：

```text
.gos.yaml
```

---

## 4. 推荐开发顺序

第一轮只追求跑通主链路：

```text
1. go mod init
2. cmd/gos/main.go
3. internal/command/new.go
4. internal/generator 基础结构
5. internal/template 基础渲染
6. internal/filesystem 基础写入
7. templates/api-clean 最小模板
8. gos new myapp 跑通
9. 生成后的 myapp go test ./... 通过
```

第二轮补齐生成体验：

```text
1. 冲突检测
2. gofmt
3. 友好输出
4. dry-run
5. --force
6. 命名转换
7. 单元测试
```

第三轮补齐业务代码生成：

```text
1. make:usecase
2. make:handler
3. make:repository
4. make:migration
5. make:test
```

第四轮增强模板质量：

```text
1. 中间件
2. 日志
3. 错误映射
4. OpenAPI
5. Docker Compose
6. CI
```

---

## 5. 近期迭代计划

### Sprint 0：项目骨架

产出：

```text
1. go.mod
2. cmd/gos/main.go
3. internal/command
4. internal/generator
5. internal/template
6. internal/filesystem
7. internal/naming
8. README 使用说明补充
```

验收：

```text
1. go test ./... 通过
2. go run ./cmd/gos --help 有输出
```

### Sprint 1：new 命令 MVP

产出：

```text
1. gos new <project>
2. api-clean 最小模板
3. 项目目录生成
4. go.mod 生成
5. Makefile 生成
6. .env.example 生成
```

验收：

```text
1. go run ./cmd/gos new myapp 成功
2. 生成项目结构符合设计文档第 5 章
3. 生成项目 go test ./... 通过
```

### Sprint 2：生成器基础能力

产出：

```text
1. 模板上下文
2. 文件冲突检测
3. --force
4. --dry-run
5. gofmt
6. 生成结果摘要
```

验收：

```text
1. 已有文件不会被静默覆盖
2. dry-run 不产生文件
3. Go 文件自动格式化
```

### Sprint 3：make 命令

产出：

```text
1. make:usecase
2. make:handler
3. make:repository
4. make:migration
5. make:test
```

验收：

```text
1. 每个命令有测试
2. 每个命令生成文件路径正确
3. 生成代码可编译
```

### Sprint 4：项目模板增强

产出：

```text
1. config
2. app
3. router
4. middleware
5. response
6. errors
7. transaction
8. Docker Compose
9. GitHub Actions
```

验收：

```text
1. 生成项目可以本地启动
2. 生成项目测试通过
3. 生成项目文档清晰
```

---

## 6. 测试计划

### 6.1 脚手架自身测试

必须覆盖：

```text
1. 命令参数解析
2. 命名转换
3. 模板渲染
4. 文件写入
5. 冲突检测
6. dry-run
7. force 覆盖
8. migration 时间戳生成
9. 项目结构生成
```

推荐测试命令：

```bash
go test ./...
go vet ./...
```

### 6.2 生成项目测试

必须覆盖：

```text
1. go test ./...
2. go vet ./...
3. go build ./cmd/api
4. Handler 示例测试
5. Usecase 示例测试
```

### 6.3 快照测试

适合用于：

```text
1. 项目结构快照
2. 关键模板输出快照
3. README 输出快照
4. Makefile 输出快照
```

注意：快照测试只验证结构和文本稳定性，核心逻辑仍应写普通单元测试。

---

## 7. 发布计划

### 7.1 发布前检查

```text
1. gofmt ./...
2. go test ./...
3. go vet ./...
4. 生成示例项目
5. 测试示例项目
6. 检查 README
7. 检查 CHANGELOG
```

### 7.2 版本语义

```text
v0.1.x 修复 MVP 问题
v0.2.x 增强 make 命令
v0.3.x 增强工程模板
v0.4.x 增强测试和契约
v0.5.x 引入多模板
v0.6.x 引入高级能力
```

### 7.3 首个可用版本发布标准

```text
1. gos new 可用
2. api-clean 模板可用
3. 生成项目可编译
4. 生成项目可测试
5. 基础文档完整
6. 不存在明显破坏性生成行为
```

---

## 8. 风险与决策

### 8.1 主要风险

```text
1. 模板过重，导致维护成本升高
2. 抽象过早，导致生成器复杂
3. 支持选项过多，导致测试矩阵膨胀
4. 生成代码隐藏依赖，违背 Go 原生原则
5. 默认技术选型引发不必要争议
```

### 8.2 控制策略

```text
1. MVP 只支持 api-clean
2. 默认只放必要模板
3. 高级能力保持可选
4. 每个新选项必须有测试
5. 生成代码优先可读性
6. 不在 v0.1.0 引入插件机制
```

### 8.3 待决策事项

```text
1. CLI 库选择：标准库 flag、cobra、urfave/cli
2. 默认 HTTP 框架：Gin 或 Chi
3. 默认数据库方案：sqlx、sqlc 或 GORM
4. 是否在 v0.1.0 内置示例用户模块
5. 是否默认生成 OpenAPI
6. 是否默认生成 Docker Compose
```

建议默认：

```text
1. CLI：cobra，便于多命令扩展
2. HTTP：Gin，降低入门成本
3. DB：sqlx，兼顾显式 SQL 和轻量封装
4. 示例模块：保留用户注册最小示例
5. OpenAPI：v0.4.0 再增强
6. Docker Compose：v0.1.0 提供基础版本
```

---

## 9. 完成定义

单个功能完成必须满足：

```text
1. 有明确命令或模板入口
2. 有单元测试或生成项目验证
3. 错误信息清晰
4. 文档已更新
5. 生成代码 gofmt 通过
6. 不破坏已有命令
```

单个版本完成必须满足：

```text
1. 版本范围内任务完成
2. go test ./... 通过
3. 生成示例项目验证通过
4. README 与使用示例更新
5. CHANGELOG 更新
```

---

## 10. 当前下一步

建议立即执行：

```text
1. 初始化 go.mod
2. 创建 cmd/gos/main.go
3. 创建 internal/command/new.go
4. 创建 internal/generator 基础类型
5. 创建 internal/template 渲染器
6. 创建 internal/filesystem 写入器
7. 创建 templates/api-clean 最小模板
8. 实现 gos new <project> 最短路径
```

第一阶段不要急于做完整模板，先让一条生成链路跑通。跑通后再逐步把 README 中的完整设计拆进模板和生成器。

---

## 11. 执行进度

### 2026-06-03

已完成：

```text
1. 初始化 Go 模块
2. 创建 cmd/gos CLI 入口
3. 实现 gos help
4. 实现 gos new <project>
5. 支持 --module、--template、--force、--dry-run
6. 支持参数顺序兼容，例如 gos new myapp --module=example.com/myapp
7. 创建 generator、template、filesystem、naming、scaffold 基础模块
8. 创建 api-clean 最小模板
9. 生成项目包含 go.mod、README.md、Makefile、.gitignore、.env.example
10. 生成项目包含 cmd/api、config、app、router、response、usecase 示例
11. 文件冲突默认报错
12. Go 文件生成后自动格式化
13. 增加命名转换、文件写入、命令 dry-run、项目生成单元测试
14. 增加项目级 Makefile
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. gos new 可生成临时项目
3. 生成项目 go test ./... 通过
4. 生成项目 go build ./cmd/api 通过
```

下一步：

```text
1. 扩展字段 DSL 的默认值、自定义 SQL 类型和 JSON 标签
2. 设计 make:handler 自动追加 OpenAPI path 的方案
3. 将 assembly 注册从 nil DB 占位升级为真实 DB 接入方案
4. 增加 CI / lint 模板
5. 设计多模板支持
```

### 2026-06-03 追加进度

已完成：

```text
1. 实现 gos make:usecase <module>/<action>
2. 实现 gos make:migration <name>
3. make:usecase 默认生成 Usecase 和单元测试骨架
4. make:migration 默认生成 up/down SQL 文件
5. 生成命令支持 --force 和 --dry-run
6. api-clean 模板增加 PowerShell 脚本：scripts/dev.ps1、scripts/test.ps1、scripts/build.ps1
7. api-clean 模板增加 Dockerfile 与 docker-compose.yml
8. 生成项目 README 改为优先展示 go 命令、PowerShell 脚本和 Docker 命令
9. Makefile 保留为可选便利入口，不作为必需运行条件
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 在生成项目内执行 make:usecase 成功
4. 在生成项目内执行 make:migration 成功
5. 生成项目 go test ./... 通过
6. 生成项目 go build ./cmd/api 通过
```

### 2026-06-03 字段 DSL 增强进度

已完成：

```text
1. make:repository --fields 支持 json=<name>
2. 生成 Domain Entity 时自动增加 json tag
3. make:repository --fields 支持 default=<value>
4. string/int/int64/bool/time 字段默认值按类型校验并生成安全 SQL 字面量
5. make:repository --fields 支持 sql=<TYPE>
6. 自定义 SQL 类型限制为安全字符集合，支持 DECIMAL(10,2) 等常见类型
7. 字段选项解析支持 SQL 类型中的括号与逗号
8. nullable 字段无显式 default 时继续不生成 DEFAULT
9. nullable 字段有显式 default 时生成 NULL DEFAULT <value>
10. 命令帮助文本补充字段 DSL 选项说明
11. default=null 仅允许用于 nullable 字段，避免生成矛盾 SQL
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 使用项目内 .tmp/go-build 作为 Go 构建缓存通过验证
3. 临时 CLI 编译到 .tmp/gos-e2e.exe 成功
4. 临时 CLI 可生成 api-clean 项目
5. 在生成项目内执行高级 make:repository --fields 成功
6. 生成项目 go test ./... 通过
```

### 2026-06-03 Handler OpenAPI 追加进度

已完成：

```text
1. gos make:handler 增加 --openapi 参数
2. --openapi 会更新 api/openapi.yaml
3. 标准 api-clean OpenAPI 文件会新增对应 path
4. 新增 path 默认生成 GET 操作
5. operationId 使用 list<TypeName>s 形式
6. 响应 schema 复用 SuccessResponse
7. 非标准 OpenAPI 文件不会被强行修改，会记录为 skipped
8. --openapi 与 --register 可组合使用
9. 命令帮助文本补充 --openapi
10. 失败时输出手动补 OpenAPI path 的提示
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 编译到 .tmp/gos-e2e.exe 成功
3. 临时 CLI 可生成 api-clean 项目
4. 在生成项目内执行 make:handler payment --register --openapi 成功
5. router.go 包含 paymentHandler.RegisterRoutes(mux)
6. api/openapi.yaml 包含 /payments
7. api/openapi.yaml 包含 operationId: listPayments
8. 生成项目 go test ./... 通过
9. 生成项目 go build ./cmd/api 通过
```

### 2026-06-04 MySQL Driver 接入进度

已完成：

```text
1. api-clean 生成项目 go.mod 固定 github.com/go-sql-driver/mysql v1.10.0
2. go.mod 显式记录 filippo.io/edwards25519 v1.2.0 indirect 依赖
3. api-clean 生成项目新增 go.sum
4. go.sum 包含 MySQL driver 及其传递依赖校验值
5. database 包新增 mysql.go，通过 blank import 注册 MySQL driver
6. README 说明 DB_DSN 为空时不建连，DB_DRIVER=mysql + DB_DSN 时启用 MySQL
7. database 包新增单元测试覆盖空 DSN 跳过建连
8. database 包新增单元测试覆盖有 DSN 但缺少 driver 的配置错误
9. 生成项目无需先手动 go get/go mod tidy 即可 go test/go build
```

验证结果：

```text
1. 通过 pkg.go.dev 确认 github.com/go-sql-driver/mysql v1.10.0
2. 使用 go mod download 校验 mysql v1.10.0 和 filippo.io/edwards25519 v1.2.0
3. 脚手架自身 go test ./... 通过
4. 临时 CLI 编译到 .tmp/gos-driver-check.exe 成功
5. 临时 CLI 可生成 api-clean 项目
6. 生成项目包含 go.mod、go.sum、database/mysql.go、database/database_test.go
7. 在生成项目内执行 make:handler payment --register --openapi 成功
8. 在生成项目内执行 make:repository invoice --register --with-migration 成功
9. 生成项目 go test ./... 通过
10. 生成项目 go build ./cmd/api 通过
```

### 2026-06-04 CI / Lint 模板进度

已完成：

```text
1. api-clean 模板新增 .github/workflows/ci.yml
2. CI 使用 actions/checkout@v5
3. CI 使用 actions/setup-go@v6
4. setup-go 通过 go-version-file: go.mod 读取 Go 版本
5. CI 开启默认 Go module cache
6. CI 执行 gofmt 格式化检查
7. CI 执行 go vet ./...
8. CI 执行 go test ./...
9. CI 执行 go build ./cmd/api
10. CI 权限收敛为 contents: read
11. api-clean 模板新增 scripts/lint.ps1
12. lint.ps1 执行 gofmt -l . 和 go vet ./...
13. README 增加 CI 与 lint 命令说明
14. go:embed 显式包含隐藏目录 .github/workflows/ci.yml.tmpl
```

验证结果：

```text
1. 通过 GitHub releases/Marketplace 信息确认 actions/checkout v5 与 actions/setup-go v6
2. 脚手架自身 go test ./... 通过
3. 临时 CLI 编译到 .tmp/gos-ci-check.exe 成功
4. 临时 CLI 可生成 api-clean 项目
5. 生成项目包含 .github/workflows/ci.yml
6. 生成项目包含 scripts/lint.ps1
7. scripts/lint.ps1 在生成项目内通过
8. 生成项目 go test ./... 通过
9. 生成项目 go build ./cmd/api 通过
10. 新增 handler/repository 后 scripts/lint.ps1 通过
11. 新增 handler/repository 后生成项目 go test ./... 通过
12. 新增 handler/repository 后生成项目 go build ./cmd/api 通过
```

### 2026-06-04 文档当前状态整理进度

已完成：

```text
1. README 增加当前实现状态章节
2. README 明确当前已实现命令列表
3. README 标注多模板仍属于后续路线图
4. README 的 CLI 命令设计章节改为当前真实命令
5. README 的 make:handler 示例补充 --register 和 --openapi
6. README 的 make:repository 示例补充 --fields、--with-migration 和 --register
7. README 的 CI 示例更新为 actions/checkout@v5 和 actions/setup-go@v6
8. README 的 CI 示例改为 go-version-file: go.mod
9. DEVELOPMENT_PLAN 增加当前实现快照
10. DEVELOPMENT_PLAN 标注 NewOrderRepository(nil) 为历史记录，当前已升级为 NewOrderRepository(db)
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. README 中不再出现过时的 actions/checkout@v4
3. README 中不再出现过时的 actions/setup-go@v5
4. README 中不再把 --db=sqlc、--http=gin、--with-openapi 写成当前 new 命令参数
5. DEVELOPMENT_PLAN 顶部可直接看到当前已实现和未完成项
```

### 2026-06-04 取消 PowerShell 脚本生成进度

已完成：

```text
1. 删除 api-clean 模板中的 scripts/dev.ps1.tmpl
2. 删除 api-clean 模板中的 scripts/test.ps1.tmpl
3. 删除 api-clean 模板中的 scripts/build.ps1.tmpl
4. 删除 api-clean 模板中的 scripts/lint.ps1.tmpl
5. 生成项目 README 不再展示 PowerShell scripts 章节
6. 项目结构快照测试移除 scripts/*.ps1
7. 项目结构快照测试新增 scripts 目录不存在断言
8. README 当前实现状态移除 PowerShell 脚本描述
9. DEVELOPMENT_PLAN 当前实现快照移除 scripts/lint.ps1 描述
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 编译到 .tmp/gos-no-scripts-check.exe 成功
3. 临时 CLI 可生成 api-clean 项目
4. 生成项目不包含 scripts 目录
5. 生成项目仍包含 .github/workflows/ci.yml
6. 生成项目 go test ./... 通过
7. 生成项目 go build ./cmd/api 通过
```

### 2026-06-04 CORS 中间件进度

已完成：

```text
1. api-clean 模板新增 internal/interfaces/http/middleware/cors.go
2. 新增 CORSOptions 配置结构
3. 默认允许所有 Origin
4. 默认允许 GET、POST、PUT、PATCH、DELETE、OPTIONS
5. 默认允许 Authorization、Content-Type、X-Request-ID 请求头
6. 支持 AllowCredentials
7. 支持 Access-Control-Max-Age
8. OPTIONS 预检请求返回 204 并跳过后续 handler
9. Router 默认接入 middleware.CORS(middleware.CORSOptions{})
10. middleware 测试覆盖普通跨域响应头
11. middleware 测试覆盖预检请求短路
12. README 当前实现状态加入 CORS
13. DEVELOPMENT_PLAN 当前实现快照将 CORS 标为已完成
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 编译到 .tmp/gos-cors-check.exe 成功
3. 临时 CLI 可生成 api-clean 项目
4. 生成项目包含 internal/interfaces/http/middleware/cors.go
5. router.go 包含 middleware.CORS(middleware.CORSOptions{})
6. 生成项目 go test ./... 通过
7. 生成项目 go build ./cmd/api 通过
8. 新增 handler/repository 后生成项目 go test ./... 通过
9. 新增 handler/repository 后生成项目 go build ./cmd/api 通过
```

### 2026-06-04 事务管理模板进度

已完成：

```text
1. api-clean 模板新增 internal/infrastructure/database/transaction.go
2. 新增 Executor 接口，兼容 *sql.DB 与 *sql.Tx 的常用查询方法
3. 新增 TxManager
4. 新增 NewTxManager(db)
5. 新增 TxManager.WithinTx(ctx, fn)
6. WithinTx 支持嵌套事务场景复用已有事务上下文
7. WithinTx 在 fn 返回错误时 rollback
8. WithinTx 在 panic 时 rollback 后继续 panic
9. WithinTx 在成功时 commit
10. 新增 ExecutorFromContext(ctx, db)
11. app.Dependencies 增加 Transactions *database.TxManager
12. BuildDependencies 默认创建 database.NewTxManager(db)
13. make:repository 生成的 MySQL Repository 改为使用 database.ExecutorFromContext(ctx, r.db)
14. Repository 在普通上下文中使用 *sql.DB
15. Repository 在事务上下文中使用 *sql.Tx
16. database 测试覆盖缺少 DB 的事务错误
17. database 测试覆盖 nil transaction function 错误
18. README 当前实现状态加入事务管理
19. DEVELOPMENT_PLAN 当前实现快照将事务模板标为已完成
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 编译到 .tmp/gos-tx-check.exe 成功
3. 临时 CLI 可生成 api-clean 项目
4. 生成项目包含 internal/infrastructure/database/transaction.go
5. assembly.go 包含 Transactions *database.TxManager
6. assembly.go 包含 database.NewTxManager(db)
7. make:repository 生成的 repository 包含 ExecutorFromContext
8. 生成项目 go test ./... 通过
9. 生成项目 go build ./cmd/api 通过
```

### 2026-06-04 Savepoint 嵌套事务后续特性规划（已于后续实现）

已规划：

```text
1. 将事务内再次开启事务时使用 SAVEPOINT 作为后续特性
2. 通过 enableNestedTransaction 配置项控制是否启用
3. enableNestedTransaction=false 时保持当前行为，内层 WithinTx 复用外层事务
4. enableNestedTransaction=true 时，事务上下文内再次调用 WithinTx 创建 SAVEPOINT
5. 内层事务成功时执行 RELEASE SAVEPOINT
6. 内层事务失败时执行 ROLLBACK TO SAVEPOINT
7. 内层 savepoint 不代表最终持久化，最终仍由最外层事务 commit 决定
8. README 事务管理设计章节补充当前行为与 savepoint 后续计划
9. DEVELOPMENT_PLAN 当前未完成项加入 savepoint 嵌套事务
```

后续实现注意事项：

```text
1. 需要设计配置位置，例如 DatabaseConfig.EnableNestedTransaction 或 TransactionConfig.EnableNestedTransaction
2. 需要定义 savepoint 命名策略，避免嵌套冲突
3. 需要明确 panic 时 savepoint rollback 语义
4. 需要覆盖 MySQL 集成测试
5. 需要避免让开发者误以为内层 release 等同于最终 commit
```

### 2026-06-04 Savepoint 嵌套事务实现进度

已完成：

```text
1. api-clean .env.example 新增 DB_ENABLE_NESTED_TRANSACTION=false
2. DatabaseConfig 新增 EnableNestedTransaction bool
3. config.Load 通过 getEnvBool("DB_ENABLE_NESTED_TRANSACTION", false) 读取配置
4. 新增 getEnvBool 辅助函数
5. TxManager 新增 TxOptions
6. NewTxManager 支持可选 TxOptions 参数
7. BuildDependencies 将 cfg.Database.EnableNestedTransaction 注入 TxManager
8. 事务上下文从直接存 *sql.Tx 升级为 txState
9. txState 负责保存当前事务和 savepoint 序号
10. enableNestedTransaction=false 时嵌套 WithinTx 继续复用外层事务
11. enableNestedTransaction=true 时嵌套 WithinTx 创建 SAVEPOINT
12. 内层成功时执行 RELEASE SAVEPOINT
13. 内层返回错误时执行 ROLLBACK TO SAVEPOINT
14. 内层 panic 时执行 ROLLBACK TO SAVEPOINT 后继续 panic
15. savepoint 名称按 sp_<n> 生成，避免同一事务内冲突
16. ExecutorFromContext 兼容新的 txState
17. database 测试覆盖 TxOptions 保存
18. README 事务管理设计章节改为已实现说明
19. DEVELOPMENT_PLAN 当前实现快照将 savepoint 标为已完成
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 编译到 .tmp/gos-savepoint-check.exe 成功
3. 临时 CLI 可生成 api-clean 项目
4. 生成项目 .env.example 包含 DB_ENABLE_NESTED_TRANSACTION=false
5. 生成项目 config.go 包含 EnableNestedTransaction 和 getEnvBool
6. 生成项目 transaction.go 包含 SAVEPOINT、ROLLBACK TO SAVEPOINT、RELEASE SAVEPOINT
7. 生成项目 assembly.go 注入 cfg.Database.EnableNestedTransaction
8. make:repository 生成的 repository 继续使用 ExecutorFromContext
9. 生成项目 go test ./... 通过
10. 生成项目 go build ./cmd/api 通过
```

### 2026-06-04 嵌套事务配置命名优化进度

已完成：

```text
1. 将配置语义从 savepoint 实现细节调整为 nested transaction 业务语义
2. DatabaseConfig.EnableSavepoint 重命名为 DatabaseConfig.EnableNestedTransaction
3. TxOptions.EnableSavepoint 重命名为 TxOptions.EnableNestedTransaction
4. 环境变量 DB_ENABLE_SAVEPOINT 重命名为 DB_ENABLE_NESTED_TRANSACTION
5. .env.example 使用 DB_ENABLE_NESTED_TRANSACTION=false
6. BuildDependencies 注入 cfg.Database.EnableNestedTransaction
7. README 使用 enableNestedTransaction 描述配置项
8. README 同时说明对应环境变量为 DB_ENABLE_NESTED_TRANSACTION
9. 项目生成快照测试更新为新命名
10. DEVELOPMENT_PLAN 同步新命名
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 编译到 .tmp/gos-nested-name-check.exe 成功
3. 临时 CLI 可生成 api-clean 项目
4. 生成项目 .env.example 包含 DB_ENABLE_NESTED_TRANSACTION=false
5. 生成项目 config.go 包含 EnableNestedTransaction
6. 生成项目 assembly.go 注入 cfg.Database.EnableNestedTransaction
7. 生成项目不再包含 EnableSavepoint 或 DB_ENABLE_SAVEPOINT
8. 生成项目 go test ./... 通过
9. 生成项目 go build ./cmd/api 通过
```

### 2026-06-04 make:model 实现进度

已完成：

```text
1. 新增 gos make:model <module>
2. make:model 支持 --fields
3. make:model 复用 make:repository 字段 DSL
4. make:model 支持 json tag 生成
5. make:model 支持 time 字段 import "time"
6. make:model 支持 --force
7. make:model 支持 --dry-run
8. 默认生成 internal/domain/<module>/entity.go
9. 已存在 entity.go 时默认报冲突，不静默覆盖
10. gos help 增加 make:model
11. 命令层增加 make:model dry-run 测试
12. 生成器增加 model 生成、字段校验、冲突测试
13. README 当前命令列表加入 make:model
14. DEVELOPMENT_PLAN 当前实现快照将 make:model 标为已完成
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 编译到 .tmp/gos-model-check.exe 成功
3. 临时 CLI 可生成 api-clean 项目
4. 在生成项目内执行 make:model invoice --fields=... 成功
5. 生成 entity.go 包含 json tag 和 time.Time 字段
6. 重复执行 make:model invoice 默认返回文件已存在错误
7. 生成项目 go test ./... 通过
8. 生成项目 go build ./cmd/api 通过
```

### 2026-06-04 Repository 集成测试模板进度

已完成：

```text
1. make:repository 额外生成 *_repository_integration_test.go
2. 集成测试文件使用 //go:build integration
3. 默认 go test ./... 不运行集成测试
4. 集成测试通过 TEST_DATABASE_DSN 读取测试库 DSN
5. TEST_DATABASE_DSN 为空时自动 t.Skip
6. 集成测试启动时 DROP TABLE IF EXISTS <table>
7. 集成测试按生成字段创建测试表
8. 集成测试覆盖 Save
9. 集成测试覆盖 FindByID
10. 集成测试覆盖 DeleteByID
11. 集成测试验证生成 ID
12. 集成测试验证字段值
13. 集成测试支持 time 字段样例值
14. 不新增 testcontainers 或 sqlmock 依赖
15. 生成项目 README 增加 integration build tag 与 TEST_DATABASE_DSN 说明
16. README 当前状态将 Repository 集成测试模板标为已完成
17. DEVELOPMENT_PLAN 当前实现快照将 Repository 集成测试模板标为已完成
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 编译到 .tmp/gos-integration-check.exe 成功
3. 临时 CLI 可生成 api-clean 项目
4. make:repository 生成 invoice_repository_integration_test.go
5. integration test 文件包含 go:build integration
6. integration test 文件包含 TEST_DATABASE_DSN
7. 默认生成项目 go test ./... 通过
8. 未配置 TEST_DATABASE_DSN 时 go test -tags=integration ./internal/infrastructure/persistence/mysql 通过并跳过测试
9. 生成项目 go build ./cmd/api 通过
```

### 2026-06-04 测试数据库工作流进度

已完成：

```text
1. api-clean 模板新增 deployments/docker/docker-compose.test.yml
2. docker-compose.test.yml 提供 mysql_test 服务
3. mysql_test 默认创建 <app>_test 数据库
4. mysql_test 映射宿主机 3307 到容器 3306，避免和开发库端口冲突
5. mysql_test 增加 mysqladmin healthcheck
6. 生成项目 README 增加本地测试库启动命令
7. 生成项目 README 使用 3307 端口示例 TEST_DATABASE_DSN
8. GitHub Actions CI 增加 MySQL service
9. CI 增加 Integration Test 步骤
10. CI 中存在 internal/infrastructure/persistence/mysql 时运行 go test -tags=integration
11. CI 中不存在 mysql repository 包时跳过 integration test
12. 当前 README 后续路线图移除更完整测试数据库工作流
13. DEVELOPMENT_PLAN 当前实现快照将测试数据库工作流标为已完成
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 编译到 .tmp/gos-testdb-check.exe 成功
3. 临时 CLI 可生成 api-clean 项目
4. 生成项目包含 deployments/docker/docker-compose.test.yml
5. docker-compose.test.yml 包含 mysql_test、3307:3306 和 healthcheck
6. CI 文件包含 Integration Test、TEST_DATABASE_DSN 和 go test -tags=integration
7. make:repository 后默认生成项目 go test ./... 通过
8. 未配置 TEST_DATABASE_DSN 时 go test -tags=integration ./internal/infrastructure/persistence/mysql 通过并跳过测试
9. 生成项目 go build ./cmd/api 通过
```

### 2026-06-03 Assembly 数据库接入进度

已完成：

```text
1. api-clean 模板新增 internal/infrastructure/database 包
2. database.Open(ctx, cfg.Database) 使用标准库 database/sql 建立连接
3. DB_DSN 为空时数据库连接保持禁用，便于本地零依赖启动和测试
4. DB_DSN 非空但 DB_DRIVER 为空时返回配置错误
5. 建连后执行 PingContext 验证连接可用
6. Ping 失败时自动 Close 已打开的 DB
7. app.Dependencies 增加 DB *sql.DB
8. BuildDependencies 统一创建数据库连接
9. Dependencies 增加 Close 方法
10. App.Run 退出时关闭 Dependencies
11. make:repository --register 生成 NewXRepository(db)，不再使用 nil 占位
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 编译到 .tmp/gos-e2e.exe 成功
3. 临时 CLI 可生成 api-clean 项目
4. 生成项目包含 internal/infrastructure/database/database.go
5. 在生成项目内执行 make:repository invoice --register --with-migration 成功
6. assembly.go 包含 database.Open(ctx, cfg.Database)
7. assembly.go 包含 NewInvoiceRepository(db)
8. 生成项目 go test ./... 通过
9. 生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第十三次追加进度

已完成：

```text
1. 扩展 make:repository --fields DSL
2. 保持兼容 name:type 旧格式
3. 新增 nullable 字段属性
4. 新增 required 字段属性
5. 新增 unique 字段属性
6. 新增 index 字段属性
7. 新增 size=N 字段属性，仅支持 string 字段
8. migration up.sql 支持 NULL / NOT NULL
9. migration up.sql 支持 VARCHAR(N)
10. migration up.sql 支持 UNIQUE KEY
11. migration up.sql 支持普通 KEY
12. 字段属性解析支持 email:string:unique,size=320 形式
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 在生成项目内执行 make:repository customer --fields=email:string:unique,size=320,deleted_at:time:nullable,index,age:int:required --with-migration --force 成功
4. migration up.sql 包含 email VARCHAR(320)
5. migration up.sql 包含 deleted_at TIMESTAMP NULL
6. migration up.sql 包含 UNIQUE KEY uk_email
7. migration up.sql 包含 KEY idx_deleted_at
8. 生成项目 go test ./... 通过
9. 生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第十四次追加进度

已完成：

```text
1. api-clean 模板新增 api/openapi.yaml
2. OpenAPI 使用 3.0.3
3. 默认包含 /healthz 契约
4. 默认包含 SuccessResponse schema
5. 默认包含 ErrorResponse schema
6. README 增加 API contract 说明
7. 项目结构快照测试覆盖 api/openapi.yaml
8. 快照内容断言覆盖 openapi 版本和 /healthz
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 生成项目包含 api/openapi.yaml
4. api/openapi.yaml 包含 openapi: 3.0.3
5. api/openapi.yaml 包含 /healthz
6. 生成项目 go test ./... 通过
7. 生成项目 go build ./cmd/api 通过
8. docker compose -f deployments/docker/docker-compose.yml config 通过
```

### 2026-06-03 第十五次追加进度

已完成：

```text
1. gos make:repository 增加 --register 参数
2. api-clean 的 internal/app/assembly.go 增加稳定 marker
3. --register 会自动向 Dependencies 增加 Repository 字段
4. --register 会自动增加 mysql repository import
5. --register 会自动增加 Repository 构造语句
6. --register 会自动在返回 Dependencies 时填充字段
7. 非标准 assembly.go 会降级为 skipped，不中断 repository 文件生成
8. 项目结构快照测试覆盖 assembly marker
9. 增加 assembly 自动注册单元测试
10. 增加非标准 assembly 降级单元测试
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 在标准 assembly 项目内执行 make:repository order --register --force 成功
4. assembly.go 自动新增 OrderRepository 字段
5. assembly.go 自动新增 mysqlrepo import
6. assembly.go 自动新增 NewOrderRepository(nil)（历史记录；2026-06-03 后续已升级为 NewOrderRepository(db)）
7. 标准注册后生成项目 go test ./... 通过
8. 标准注册后生成项目 go build ./cmd/api 通过
9. 非标准 assembly 项目内执行 make:repository invoice --register --force 输出 skip internal/app/assembly.go
```

### 2026-06-03 第十六次追加进度

已完成：

```text
1. .gitignore 增加 .tmp/
2. 后续端到端临时项目改为放在当前项目 .tmp 下
3. Handler 测试模板增加 Content-Type 断言
4. Handler 测试模板增加 JSON 响应解码
5. Handler 测试模板增加 code == OK 断言
6. Handler 测试模板增加 message == success 断言
7. make:handler --register 遇到非标准 router 时会输出手动注册提示
8. 手动注册提示包含 NewXHandler 构造代码
9. 手动注册提示包含 RegisterRoutes(mux) 代码
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 构建到 .tmp/gos-e2e.exe
3. 临时项目生成到 .tmp/gola-e2e-handler-next
4. make:handler invoice --force 生成的测试包含响应结构断言
5. 生成项目 go test ./... 通过
6. 生成项目 go build ./cmd/api 通过
7. 非标准 router 下 make:handler payment --register --force 输出 skip router.go
8. 非标准 router 下输出手动注册提示
9. 降级后生成项目 go test ./... 通过
10. 降级后生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第十二次追加进度

已完成：

```text
1. make:handler --register 遇到非标准 router 时不再失败
2. 非标准 router 自动注册失败会降级为 skipped 结果
3. Handler 文件和 Handler 测试仍会正常生成
4. 标准 api-clean router 仍会自动 update
5. 新增非标准 router 注册降级单元测试
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 标准 router 下 make:handler payment --register --force 输出 update router.go
4. 非标准 router 下 make:handler invoice --register --force 输出 skip router.go
5. 降级后生成项目 go test ./... 通过
6. 降级后生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第六次追加进度

已完成：

```text
1. gos make:repository 增加 --with-migration 参数
2. gos make:repository 增加 --migration-dir 参数
3. Repository 生成时可同步生成 create table up/down SQL
4. 默认迁移目录为 migrations
5. 迁移文件名使用时间戳与表名，例如 create_sales_invoices_table
6. up.sql 生成基础 id BIGINT PRIMARY KEY AUTO_INCREMENT
7. down.sql 生成 DROP TABLE IF EXISTS
8. --dry-run 会同时展示 repository 文件和 migration 文件
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 在生成项目内执行 make:repository invoice --table=sales_invoices --with-migration --force 成功
4. 生成 migrations/*_create_sales_invoices_table.up.sql
5. 生成 migrations/*_create_sales_invoices_table.down.sql
6. up.sql 包含 CREATE TABLE sales_invoices
7. down.sql 包含 DROP TABLE IF EXISTS sales_invoices
8. 生成项目 go test ./... 通过
9. 生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第七次追加进度

已完成：

```text
1. gos make:repository 增加 --fields 参数
2. 字段格式为 name:type，多个字段用逗号分隔
3. 首批支持 string、int、int64、bool、time
4. 字段名会转换为 snake_case SQL column
5. 字段名会转换为 PascalCase Go struct field
6. time 字段会自动生成 import "time"
7. Repository SELECT 会包含字段列
8. Repository Scan 会包含字段目标
9. Repository INSERT 会包含字段列和值
10. Repository UPDATE 会包含字段 set 语句
11. --with-migration 生成的 up.sql 会包含字段列
12. 字段名 id 会被拒绝，避免和主键冲突
13. 不支持的字段类型会返回清晰错误
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 在生成项目内执行 make:repository invoice --table=sales_invoices --fields=number:string,total:int64,paid:bool,created_at:time --with-migration --force 成功
4. entity.go 包含 Number、Total、Paid、CreatedAt 字段
5. invoice_repository.go 包含字段化 SELECT、INSERT、UPDATE
6. migration up.sql 包含 number、total、paid、created_at 字段
7. 生成项目 go test ./... 通过
8. 生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第八次追加进度

已完成：

```text
1. 实现 gos make:test <usecase|handler|repository> <name>
2. make:test usecase <module>/<action> 生成 Usecase 测试骨架
3. make:test handler <module> 生成 Handler 测试骨架
4. make:test repository <module> 生成 Repository 测试骨架
5. make:test 支持 --force
6. make:test 支持 --dry-run
7. make:test 自动读取当前项目 go.mod 的 module path
8. 不支持的测试类型会返回清晰错误
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 在生成项目内执行 make:test usecase order/create --force 成功
4. 在生成项目内执行 make:test handler order --force 成功
5. 在生成项目内执行 make:test repository order --force 成功
6. 生成项目 go test ./... 通过
7. 生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第九次追加进度

已完成：

```text
1. 为 api-clean 增加项目结构快照测试
2. 快照覆盖所有默认生成文件
3. 快照测试使用稳定排序
4. 新增关键内容断言
5. 断言 router 默认接入 middleware.RequestID
6. 断言 router 默认接入 middleware.Timeout
7. 断言 docker-compose.yml 包含 MySQL
8. 断言 docker-compose.yml 包含 Redis
9. 断言 .env.example 包含 DB_DSN
10. 断言 .env.example 包含 REDIS_ADDR
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 生成项目 go test ./... 通过
4. 生成项目 go build ./cmd/api 通过
5. docker compose -f deployments/docker/docker-compose.yml config 通过
```

### 2026-06-03 第十次追加进度

已完成：

```text
1. api-clean 模板新增 internal/app/assembly.go
2. 新增 BuildDependencies(ctx, cfg) 作为显式依赖组装落点
3. App.New 默认调用 BuildDependencies
4. 默认不打开数据库、不引入第三方驱动，保持生成项目可直接运行和编译
5. README 增加 Dependency assembly 说明
6. 项目结构快照测试覆盖 assembly.go
7. 快照内容断言覆盖 BuildDependencies 调用
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 生成项目包含 internal/app/assembly.go
4. 生成项目 go test ./... 通过
5. 生成项目 go build ./cmd/api 通过
6. docker compose -f deployments/docker/docker-compose.yml config 通过
```

### 2026-06-03 第十一次追加进度

已完成：

```text
1. api-clean 模板新增 middleware 测试模板
2. middleware 测试覆盖 RequestID 使用传入 X-Request-ID
3. middleware 测试覆盖 Recover 输出 500
4. middleware 测试覆盖 Chain 中间件执行顺序
5. api-clean 模板新增 httperror 测试模板
6. httperror 测试覆盖 AppError 状态码映射
7. httperror 测试覆盖未知错误兜底为 INTERNAL_ERROR
8. 项目结构快照测试覆盖新增测试文件
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 生成项目包含 middleware_test.go
4. 生成项目包含 mapper_test.go
5. 生成项目 go test ./... 通过
6. 生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第三次追加进度

已完成：

```text
1. api-clean 配置结构增加 DatabaseConfig
2. api-clean 配置结构增加 RedisConfig
3. .env.example 增加 DB_DRIVER、DB_DSN、REDIS_ADDR、REDIS_PASSWORD、REDIS_DB
4. docker-compose.yml 增加 MySQL 服务
5. docker-compose.yml 增加 Redis 服务
6. API 服务增加 MySQL 与 Redis 环境变量
7. 新增标准库 HTTP middleware 包
8. 新增 RequestID 中间件
9. 新增 Recover 中间件
10. 新增 AccessLog 中间件
11. 新增 Timeout 中间件
12. Router 默认接入中间件链
13. 新增 apperror 包
14. 新增 httperror 错误映射包
15. README 增加配置与 Docker 说明
16. 生成项目测试断言覆盖新增模板文件
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成增强后的 api-clean 项目
3. 生成项目 go test ./... 通过
4. 生成项目 go build ./cmd/api 通过
5. docker compose -f deployments/docker/docker-compose.yml config 通过
6. 增强模板项目内执行 make:handler 成功
7. 增强模板项目内执行 make:repository 成功
8. 新增 handler/repository 后生成项目 go test ./... 通过
9. 新增 handler/repository 后生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第四次追加进度

已完成：

```text
1. gos make:handler 增加 --register 参数
2. --register 会自动更新 internal/interfaces/http/router.go
3. 自动注册会补充 handler import
4. 自动注册会在 NewRouter 中实例化 Handler
5. 自动注册会调用 Handler.RegisterRoutes(mux)
6. 默认 make:handler 行为保持只生成 handler 文件
7. 文件写入器增加按文件 Overwrite 能力
8. 自动注册只在识别 api-clean 标准 router 结构时执行
9. 增加 router 自动注册单元测试
10. 增加文件 Overwrite 单元测试
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 在生成项目内执行 make:handler payment --register --force 成功
4. router.go 自动新增 handler import
5. router.go 自动新增 paymentHandler.RegisterRoutes(mux)
6. 自动注册后生成项目 go test ./... 通过
7. 自动注册后生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第五次追加进度

已完成：

```text
1. gos make:repository 增加 --table 参数
2. repository 默认表名按模块名推断，例如 order -> orders
3. 支持基础复数规则，例如 category -> categories
4. 表名会校验为安全 SQL identifier
5. MySQL Repository 从未实现占位升级为基础 CRUD 骨架
6. 生成 FindByID(ctx, id)
7. 生成 Save(ctx, entity)
8. 生成 DeleteByID(ctx, id)
9. Domain Repository 接口同步增加 DeleteByID
10. Repository 测试继续保持无外部数据库依赖
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 在生成项目内执行 make:repository invoice --table=sales_invoices --force 成功
4. 生成 Repository SQL 使用 sales_invoices 表名
5. 生成项目 go test ./... 通过
6. 生成项目 go build ./cmd/api 通过
```

### 2026-06-03 第二次追加进度

已完成：

```text
1. 实现 gos make:handler <module>
2. 实现 gos make:repository <module>
3. make:handler 自动读取当前项目 go.mod 的 module path
4. make:repository 自动读取当前项目 go.mod 的 module path
5. make:handler 默认生成 HTTP Handler 和 Handler 测试骨架
6. make:repository 默认生成 MySQL Repository、Repository 测试骨架
7. make:repository 会补充最小 Domain Entity 与 Repository 接口
8. 已有 Domain 辅助文件默认跳过，不强制覆盖
9. 增加项目 module path 检测模块
10. 增加 handler、repository、project 检测、可选文件跳过相关测试
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 在生成项目内执行 make:handler 成功
4. 在生成项目内执行 make:repository 成功
5. 生成项目 go test ./... 通过
6. 生成项目 go build ./cmd/api 通过
```

### 2026-06-05 多模板支持进度

已完成：

```text
1. ProjectGenerator 增加内置模板发现能力
2. 未知模板错误会输出可用模板列表
3. gos new --template 支持 api-clean 与 api-minimal
4. 新增 api-minimal 极简 HTTP 项目模板
5. api-minimal 生成 go.mod、README.md、Makefile、cmd/api、config、router 和 router 测试
6. README 和 docs/CLI_GUIDE.md 补充 api-minimal 使用说明
7. DEVELOPMENT_PLAN 当前实现快照将多模板标为已完成
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 临时 CLI 可生成 api-minimal 项目
4. api-clean 生成项目 go test ./... 通过
5. api-clean 生成项目 go build ./cmd/api 通过
6. api-minimal 生成项目 go test ./... 通过
7. api-minimal 生成项目 go build ./cmd/api 通过
```

### 2026-06-05 cmd 子命令与命令脚本进度

已完成：

```text
1. api-clean cmd/api 模板改为 Cobra 子命令入口
2. api-minimal cmd/api 模板改为 Cobra 子命令入口
3. 默认子命令包含 serve、schedule、queue、help
4. go run ./cmd/api 默认等同 serve
5. schedule 提供定时任务循环占位
6. queue 提供队列消费循环占位
7. cmd/api/main.go 增加 gos:command-imports 与 gos:commands 注册 marker
8. 新增 gos make:command <name>
9. make:command 默认生成 internal/command/<name>.go Cobra 命令和测试
10. make:command --register 自动更新标准 cmd/api/main.go Cobra root command
11. README、docs/CLI_GUIDE.md、docs/GENERATED_PROJECT_GUIDE.md 和生成项目 README 模板补充命令说明
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. 临时 CLI 可生成 api-clean 项目
3. 临时 CLI 可生成 api-minimal 项目
4. 生成项目 go test ./... 通过
5. 生成项目 go build ./cmd/api 通过
6. 在 api-clean 生成项目内执行 gos make:command sync-orders --register 成功
7. 注册后 api-clean 生成项目 go test ./... 通过
8. 注册后 api-clean 生成项目 go build ./cmd/api 通过
9. 在 api-minimal 生成项目内执行 gos make:command sync-orders --register 成功
10. 注册后 api-minimal 生成项目 go test ./... 通过
11. 注册后 api-minimal 生成项目 go build ./cmd/api 通过
12. go run ./cmd/api help 和 go run ./cmd/api sync-orders 可实际执行
```

### 2026-06-05 OpenTelemetry 可选支持进度

已完成：

```text
1. gos new 增加 --with-otel 参数
2. 模板渲染器支持跳过渲染后为空的可选模板文件
3. api-clean --with-otel 生成 internal/observability/otel.go
4. api-minimal --with-otel 生成 internal/observability/otel.go
5. 生成项目配置增加 ObservabilityConfig
6. HTTP router 在启用 OTEL 时使用 otelhttp 包裹
7. 启动阶段按配置初始化 OTLP HTTP trace exporter
8. 默认不使用 --with-otel 时不生成 OpenTelemetry 依赖和 observability 文件
9. README、docs/CLI_GUIDE.md、docs/GENERATED_PROJECT_GUIDE.md 和生成项目 README 模板补充 OTEL 使用说明
10. docs/OPEN_TELEMETRY.md 落地 OpenTelemetry tracing 使用、配置、代码落点和验证说明
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. api-minimal --with-otel 生成项目 go test ./... 通过
3. api-minimal --with-otel 生成项目 go build ./cmd/api 通过
4. api-clean --with-otel 生成项目 go test ./... 通过
5. api-clean --with-otel 生成项目 go build ./cmd/api 通过
6. 默认 api-minimal 生成项目不包含 OpenTelemetry 依赖和 internal/observability
7. 默认 api-minimal 生成项目 go test ./... 与 go build ./cmd/api 通过
```

### 2026-06-05 高优先级工程增强进度

已完成：

```text
1. api-clean 模板新增 internal/logging/logging.go
2. api-minimal 模板新增 internal/logging/logging.go
3. LOG_LEVEL 在生成项目启动阶段真正生效，支持 debug/info/warn/warning/error
4. 生成项目启动时调用 slog.SetDefault，让中间件、命令脚本和业务日志共享默认 logger
5. 使用 --with-otel 时 logging handler 从 context 注入 trace_id/span_id
6. gos CLI 自身改为基于 Cobra 的 root/subcommand 结构
7. 保留现有 runNew/runMake* 逻辑，降低迁移风险
8. 新增生成项目矩阵编译测试，覆盖 api-clean、api-clean --with-otel、api-minimal、api-minimal --with-otel
9. api-clean 配置中的 DB_ENABLE_NESTED_TRANSACTION、REDIS_DB、OTEL_* 使用严格解析
10. api-minimal 的 config.Load 改为返回 (Config, error)，OTEL_* 使用严格解析
11. cmd/api root 增加 configureLogging，注册后的自定义命令可继承默认 logger 初始化
```

验证结果：

```text
1. 脚手架自身 go test ./... 通过
2. internal/scaffold 矩阵测试中四种生成项目均 go test ./... 通过
3. internal/scaffold 矩阵测试中四种生成项目均 go build ./cmd/api 通过
```

### 2026-06-05 后台 Worker 骨架进度

已完成：

```text
1. api-clean 模板新增 internal/worker/worker.go
2. api-clean 模板新增 internal/worker/worker_test.go
3. api-minimal 模板新增 internal/worker/worker.go
4. api-minimal 模板新增 internal/worker/worker_test.go
5. schedule 子命令改为使用 worker.NewScheduler
6. queue 子命令改为使用 worker.NewQueueWorker
7. worker 骨架提供启动/停止日志、周期执行、错误日志和 panic recover
8. docs/OPTIMIZATION_BACKLOG.md 记录剩余优化项，并将 worker 骨架移入已完成优化
```

### 2026-06-05 HTTP 生产默认值进度

已完成：

```text
1. api-clean HTTPConfig 增加 ReadHeaderTimeout、ReadTimeout、WriteTimeout、IdleTimeout、MaxHeaderBytes
2. api-minimal 增加 HTTPConfig，并接入相同 HTTP server 配置
3. HTTP_READ_HEADER_TIMEOUT 默认 5s
4. HTTP_READ_TIMEOUT 默认 15s
5. HTTP_WRITE_TIMEOUT 默认 30s
6. HTTP_IDLE_TIMEOUT 默认 60s
7. HTTP_MAX_HEADER_BYTES 默认 1048576
8. duration/int 配置严格解析，非法值启动即失败
9. docs/OPTIMIZATION_BACKLOG.md 将 HTTP 生产默认值移入已完成优化
```

### 2026-06-09 文档与发布体验进度

已完成：

```text
1. 新增 docs/CONFIG_REFERENCE.md，整理环境变量类型、默认值和非法值行为
2. 新增 docs/TEMPLATE_DEPENDENCIES.md，整理 go.mod.tmpl/go.sum.tmpl 刷新流程
3. gos CLI 新增 version 子命令
4. version 输出 Version、Commit、BuildDate
5. Makefile build 支持通过 ldflags 注入版本、commit 和构建时间
6. docs/OPTIMIZATION_BACKLOG.md 将配置表和模板依赖刷新流程移入已完成优化
```

### 2026-06-09 请求体大小限制进度

已完成：

```text
1. api-clean HTTPConfig 增加 MaxBodyBytes
2. api-minimal HTTPConfig 增加 MaxBodyBytes
3. HTTP_MAX_BODY_BYTES 默认 10485760
4. HTTP_MAX_BODY_BYTES 使用 int64 严格解析，非法值启动即失败
5. api-clean router 使用 http.MaxBytesHandler 限制请求体大小
6. api-minimal router 使用 http.MaxBytesHandler 限制请求体大小
7. HTTP_MAX_BODY_BYTES=0 可关闭请求体大小限制
8. docs/OPTIMIZATION_BACKLOG.md 将请求体大小限制移入已完成优化
```

### 2026-06-09 CLI 发布体验进度

已完成：

```text
1. gos completion <bash|zsh|fish|powershell>
2. completion 使用 Cobra 原生 completion 生成器
3. docs/RELEASE.md 记录构建、版本注入、completion 和发布前检查
4. docs/CLI_GUIDE.md 补充 completion 使用方式
5. docs/OPTIMIZATION_BACKLOG.md 将 CLI 发布体验基础能力移入已完成优化
```

### 2026-06-09 OpenTelemetry 外部调用与本地环境进度

已完成：

```text
1. api-clean --with-otel 生成 internal/observability/http_client.go
2. api-minimal --with-otel 生成 internal/observability/http_client.go
3. observability.NewHTTPClient 使用 otelhttp.NewTransport
4. observability.NewHTTPTransport 支持传入自定义 base RoundTripper
5. docs/LOCAL_OBSERVABILITY.md 提供 otelcol debug exporter 示例
6. docs/LOCAL_OBSERVABILITY.md 提供 Jaeger all-in-one 示例
7. docs/OPEN_TELEMETRY.md 补充外部 HTTP client tracing 用法
```

### 2026-06-09 OpenTelemetry 数据库 tracing 进度

已完成：

```text
1. api-clean --with-otel 增加 github.com/XSAM/otelsql 依赖
2. api-clean database.Open 在启用 OTEL 模板时使用 otelsql.Open
3. 默认 api-clean 模板仍使用标准库 sql.Open，不引入 OTEL 依赖
4. Repository、TxManager 和业务代码继续使用标准库 *sql.DB/*sql.Tx 接口
5. database/sql span 增加 db.system.name 属性
6. Ping span 默认开启，driver.ErrSkip 默认不记录为 span error
7. docs/OPEN_TELEMETRY.md 补充数据库 tracing 说明
8. docs/LOCAL_OBSERVABILITY.md 补充本地 DB span 验证说明
9. docs/OPTIMIZATION_BACKLOG.md 将数据库 tracing 移入已完成优化
```

### 2026-06-10 安全默认值增强进度

已完成：

```text
1. api-clean Config 增加 CORSConfig
2. api-clean 新增 CORS_ALLOWED_ORIGINS、CORS_ALLOWED_METHODS、CORS_ALLOWED_HEADERS、CORS_ALLOW_CREDENTIALS、CORS_MAX_AGE
3. api-clean RouterOptions 增加 middleware.CORSOptions
4. app.New 将 cfg.CORS 映射到 HTTP router
5. Docker Compose 和 .env.example 补充 CORS_* 配置
6. api-clean/api-minimal logging handler 增加 ReplaceAttr 脱敏
7. 默认脱敏 password、passwd、secret、token、authorization、api_key、access_key、private_key、credential、dsn 等字段键
8. api-clean Recover 中间件不再记录 panic 原始值，改为记录 panic_type
9. api-clean/api-minimal 增加 logging 测试模板，验证敏感字段脱敏
10. docs/CONFIG_REFERENCE.md、docs/GENERATED_PROJECT_GUIDE.md 和生成项目 README 模板补充安全默认值说明
11. docs/OPTIMIZATION_BACKLOG.md 将安全默认值增强移入已完成优化
```

### 2026-06-10 OpenAPI 基础深化进度

已完成：

```text
1. api-clean 默认 openapi.yaml 增加 components.responses
2. 新增 BadRequest、NotFound、InternalServerError 可复用响应组件
3. 新增 ListResponse schema
4. healthz 响应增加 examples
5. gos make:handler --openapi 生成 tags
6. gos make:handler --openapi 的 200 响应引用 ListResponse
7. gos make:handler --openapi 增加 BadRequest 和 InternalServerError 响应引用
8. docs/OPTIMIZATION_BACKLOG.md 将 OpenAPI 基础深化移入已完成优化
```
