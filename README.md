# Go 原生后端脚手架项目整体设计文档

## 文档导航

如果你是第一次使用本项目，建议按下面顺序阅读：

```text
1. docs/CLI_GUIDE.md
   了解 gos 的安装、项目生成、make:* 命令、字段 DSL、dry-run/force 和常见开发流程。

2. docs/GENERATED_PROJECT_GUIDE.md
   了解 gos 生成出来的业务项目如何运行、配置、分层开发、使用事务、编写测试和接入数据库。

3. docs/OPEN_TELEMETRY.md
   了解 gos new --with-otel 生成的 OpenTelemetry tracing 能力、配置、代码落点和验证方式。

4. docs/OPTIMIZATION_BACKLOG.md
   查看后续优化清单、优先级和建议推进顺序。

5. docs/CONFIG_REFERENCE.md
   查看生成项目环境变量、默认值、类型和非法值行为。

6. docs/TEMPLATE_DEPENDENCIES.md
   查看 go.mod.tmpl/go.sum.tmpl 依赖刷新流程。

7. docs/RELEASE.md
   查看 gos CLI 构建、版本注入、shell completion 和发布前检查。

8. docs/LOCAL_OBSERVABILITY.md
   查看 --with-otel 项目的本地 Collector、Jaeger、外部 HTTP client 和数据库 tracing 示例。

9. README.md
   阅读整体设计原则、架构取舍和后续路线图。

10. DEVELOPMENT_PLAN.md
   查看开发进度、版本路线和已完成能力。
```

生成项目自身的 `README.md` 也会包含一份可直接落地的使用说明，覆盖生成代码的结构、配置、事务、测试、Docker 和最佳实践。

## 1. 项目定位

本项目旨在设计并开发一个面向 Go 语言后端应用的工程脚手架，用于快速创建结构清晰、易于测试、易于维护、适合中大型业务演进的后端项目。

项目定位为：

```text
一个 Go 原生后端应用工程脚手架。
```

它不是一个重型运行时框架，而是一个用于快速创建、规范组织和辅助开发 Go 后端项目的工具体系。

核心价值包括：

```text
1. 统一工程结构
2. 规范开发流程
3. 提供代码生成能力
4. 提供基础项目模板
5. 提供测试模板
6. 提供本地开发工具
7. 提供工程最佳实践
```

---

## 当前实现状态（2026-06-10）

当前仓库已实现一个可运行的 `gos` CLI，并内置 `api-clean` 与 `api-minimal` 项目模板。

已实现命令：

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

`api-clean` 当前包含：

```text
1. 标准库 HTTP API 项目骨架
2. 配置、统一响应、应用错误、HTTP 错误映射
3. RequestID、Recover、CORS、AccessLog、Timeout 中间件
4. OpenAPI 起始文件和 handler 自动追加能力
5. MySQL driver 注册、database/sql 连接入口、事务管理和依赖组装
6. MySQL Repository 代码生成、集成测试模板和迁移文件生成
7. cmd/api 子命令入口：serve、schedule、queue
8. 基于 Cobra 的命令脚本生成与自动注册
9. 可选 OpenTelemetry tracing 支持，包括 HTTP server、外部 HTTP client 和 database/sql
10. 生成项目日志配置支持 LOG_LEVEL，并在 OTEL 启用时注入 trace_id/span_id
11. 脚手架自身 CLI 基于 Cobra 组织命令
12. 生成项目模板矩阵编译验证
13. 生成项目配置读取对 bool/int 使用严格解析，非法环境变量启动即失败
14. 生成项目 HTTP server 默认配置 Read/Write/Idle timeout 和 MaxHeaderBytes
15. 生成项目 HTTP_MAX_BODY_BYTES 请求体大小限制
16. gos version 支持版本、commit、构建时间输出
17. gos completion 支持 bash、zsh、fish、powershell
18. 安全默认值增强：CORS 配置化、日志敏感字段脱敏、panic 输出边界
19. OpenAPI 基础深化：复用响应组件、列表响应 schema、错误响应引用和示例
20. 缓存接口：memory、file、memcache、redis 后端
21. Redis 分布式锁：SetNX 获取锁，Lua token 校验释放和续期
22. Dockerfile、Docker Compose
23. GitHub Actions CI
```

`api-minimal` 当前包含：

```text
1. 标准库 HTTP API 入口
2. 环境变量配置
3. /healthz 健康检查
4. 基础路由测试
5. Makefile 便利入口
```

仍需要继续整理的是 README/DEVELOPMENT_PLAN 中较早设计段落与当前实现状态之间的归档关系。

---

## 2. 项目目标

本项目的核心目标包括：

```text
1. 提供标准化的 Go 后端项目结构
2. 支持接口优先、测试优先、实现最后的开发流程
3. 降低新项目初始化成本
4. 降低团队成员之间的工程风格差异
5. 提供基础代码生成能力
6. 提供常见后端工程能力模板
7. 保持 Go 项目的显式、简单、可测试、可维护
```

项目应重点解决以下问题：

```text
1. 新项目启动时重复搭建目录结构
2. 不同项目分层方式不统一
3. 接口、测试、实现顺序混乱
4. HTTP、Usecase、Repository 等代码职责混杂
5. 配置、日志、错误处理、数据库迁移缺乏统一规范
6. 新人难以快速理解项目结构
7. 测试代码难以编写或难以隔离外部依赖
```

---

## 3. 设计原则

### 3.1 Go 原生优先

脚手架应优先遵循 Go 语言本身的工程习惯。

应坚持：

```text
1. 显式优于隐式
2. 组合优于继承
3. 小接口优于大接口
4. 构造函数注入优于运行时容器
5. 编译期检查优于运行时解析
6. 标准库优先，必要时再引入第三方库
7. 业务逻辑不依赖具体 Web 框架
8. 业务逻辑不依赖具体 ORM
9. 业务逻辑不依赖配置读取方式
10. 业务逻辑不依赖外部基础设施实现
```

### 3.2 轻框架、重规范

本项目不应优先开发大而全的框架核心。

更推荐的方向是：

```text
1. 提供项目模板
2. 提供代码生成器
3. 提供工程规范
4. 提供少量通用工具包
5. 提供可替换的技术选型
```

不建议默认实现：

```text
1. 重型运行时框架
2. 复杂生命周期管理
3. 复杂服务容器
4. 大量自定义抽象
5. 大量运行时反射
6. 隐式全局状态
```

### 3.3 接口优先

本项目要求遵循接口优先的开发流程。

Go 中接口不应滥用。接口应定义在“使用方”，而不是“实现方”。

不是因为存在一个 `MysqlUserRepository`，所以定义一个 `UserRepository` 接口。

而是因为某个 Usecase 需要“用户存储能力”，所以由 Usecase 或 Domain 定义该接口。

推荐原则：

```text
1. 在业务边界定义接口
2. 在需要替换实现的地方定义接口
3. 在需要 Mock 的地方定义接口
4. 在依赖外部系统的地方定义接口
5. 不为每个结构体机械创建接口
```

### 3.4 测试优先

项目应遵循：

```text
先定义接口
再定义测试
最后开发实现
```

推荐开发顺序：

```text
1. 定义接口契约
2. 定义 Usecase 输入输出
3. 定义 Usecase 依赖的外部能力接口
4. 编写 Usecase 单元测试
5. 实现 Usecase
6. 编写 Repository 集成测试
7. 实现 Repository
8. 编写 HTTP Handler 测试
9. 实现 HTTP Handler
10. 执行完整测试
```

---

## 4. 脚手架自身架构

脚手架项目本身建议采用如下结构：

```text
go-scaffold/
├── cmd/
│   └── gos/
│       └── main.go
│
├── internal/
│   ├── command/
│   │   ├── new.go
│   │   ├── make_usecase.go
│   │   ├── make_handler.go
│   │   ├── make_repository.go
│   │   ├── make_model.go
│   │   ├── make_migration.go
│   │   └── make_test.go
│   │
│   ├── generator/
│   │   ├── generator.go
│   │   ├── context.go
│   │   └── renderer.go
│   │
│   ├── template/
│   │   ├── loader.go
│   │   ├── embedded.go
│   │   └── resolver.go
│   │
│   ├── naming/
│   │   ├── case.go
│   │   ├── plural.go
│   │   └── path.go
│   │
│   ├── filesystem/
│   │   ├── writer.go
│   │   ├── conflict.go
│   │   └── formatter.go
│   │
│   └── project/
│       ├── detector.go
│       └── config.go
│
├── templates/
│   ├── api-clean/
│   ├── api-basic/
│   ├── grpc-service/
│   ├── worker/
│   └── monolith/
│
├── go.mod
├── go.sum
└── README.md
```

### 4.1 cmd

保存命令行程序入口。

```text
cmd/gos/main.go
```

`gos` 是脚手架命令名称，也可以替换为项目自定义名称。

### 4.2 internal/command

保存所有命令的实现逻辑。

例如：

```bash
gos new myapp
gos make:usecase user/register
gos make:handler user
gos make:repository user
gos make:migration create_users_table
gos make:command sync-orders
```

### 4.3 internal/generator

负责代码生成流程。

核心流程包括：

```text
1. 解析生成参数
2. 构建生成上下文
3. 加载模板
4. 渲染模板
5. 检查文件冲突
6. 写入文件
7. 执行 gofmt
8. 输出生成结果
```

### 4.4 internal/template

负责模板读取、嵌入、查找和解析。

模板建议使用 Go 标准库 `text/template`。

### 4.5 internal/naming

负责命名转换。

例如：

```text
UserProfile
user_profile
user-profile
userProfiles
UserProfiles
```

### 4.6 internal/filesystem

负责文件写入、目录创建、冲突检测、格式化处理等。

---

## 5. 生成后的业务项目结构

默认推荐生成 `api-clean` 模板。

生成后的业务项目结构如下：

```text
myapp/
├── cmd/
│   ├── api/
│   │   └── main.go
│   │
│   ├── worker/
│   │   └── main.go
│   │
│   └── migrate/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   └── wire.go
│   │
│   ├── config/
│   │   └── config.go
│   │
│   ├── domain/
│   │   └── user/
│   │       ├── entity.go
│   │       ├── repository.go
│   │       ├── service.go
│   │       └── errors.go
│   │
│   ├── usecase/
│   │   └── user/
│   │       ├── register.go
│   │       ├── login.go
│   │       └── dto.go
│   │
│   ├── interfaces/
│   │   ├── http/
│   │   │   ├── router.go
│   │   │   ├── middleware/
│   │   │   │   ├── request_id.go
│   │   │   │   ├── recover.go
│   │   │   │   ├── access_log.go
│   │   │   │   └── timeout.go
│   │   │   └── handler/
│   │   │       └── user_handler.go
│   │   │
│   │   ├── grpc/
│   │   └── cli/
│   │
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── mysql/
│   │   │   │   └── user_repository.go
│   │   │   └── redis/
│   │   │
│   │   ├── queue/
│   │   ├── mail/
│   │   ├── storage/
│   │   └── external/
│   │
│   └── pkg/
│       ├── response/
│       ├── validator/
│       ├── logger/
│       ├── errors/
│       ├── pagination/
│       └── transaction/
│
├── migrations/
│   └── 20260603120000_create_users_table.sql
│
├── test/
│   ├── unit/
│   ├── integration/
│   └── e2e/
│
├── api/
│   ├── openapi.yaml
│   └── proto/
│
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   │
│   └── k8s/
│
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

当前 `api-clean` 模板不生成 `scripts` 目录。常用操作直接使用 `go`、`gos`、`docker compose` 命令；`Makefile` 只是可选便利入口。

---

## 6. 业务项目目录说明

### 6.1 cmd

`cmd` 目录保存程序入口。

一个 Go 项目可以包含多个可执行程序，例如：

```text
cmd/api       HTTP API 服务入口
cmd/worker    后台任务服务入口
cmd/migrate   数据库迁移入口
```

`cmd/api/main.go` 负责：

```text
1. 加载配置
2. 初始化应用
3. 启动 HTTP 服务
4. 监听退出信号
5. 执行优雅关闭
```

### 6.2 internal

`internal` 保存项目私有业务代码。

Go 语言中 `internal` 目录具有包级访问限制，外部模块无法直接导入其中内容，因此适合保存项目主体代码。

### 6.3 internal/app

负责应用组装。

主要职责：

```text
1. 初始化数据库连接
2. 初始化日志
3. 初始化 Repository
4. 初始化 Usecase
5. 初始化 Handler
6. 初始化 Router
7. 组装 HTTP Server
8. 管理应用生命周期
```

### 6.4 internal/config

负责配置加载。

配置应在启动阶段读取，随后通过结构体传递给需要的模块。

业务代码不应直接读取环境变量。

### 6.5 internal/domain

保存领域模型、领域规则、领域错误、领域接口。

Domain 层不应依赖：

```text
1. HTTP 框架
2. ORM
3. 数据库
4. Redis
5. 配置文件
6. 外部 API SDK
```

Domain 层应尽量保持纯粹。

### 6.6 internal/usecase

保存应用用例。

Usecase 是业务动作的组织者。

例如：

```text
用户注册
用户登录
创建订单
支付订单
取消订单
生成报表
```

Usecase 层负责：

```text
1. 编排领域对象
2. 调用 Repository 接口
3. 调用外部能力接口
4. 控制事务边界
5. 返回应用层输入输出 DTO
```

### 6.7 internal/interfaces

保存输入适配层。

包括：

```text
1. HTTP Handler
2. gRPC Handler
3. CLI Handler
4. Webhook Handler
```

该层负责协议适配，不负责核心业务逻辑。

### 6.8 internal/infrastructure

保存基础设施实现。

包括：

```text
1. MySQL Repository
2. Redis Cache
3. Kafka / RabbitMQ / Redis Queue
4. 邮件发送
5. 文件存储
6. 第三方 API Client
```

Infrastructure 层实现 Domain 或 Usecase 定义的接口。

### 6.9 internal/pkg

保存项目内部通用工具。

例如：

```text
1. 统一响应
2. 错误处理
3. 日志封装
4. 参数校验
5. 分页
6. 事务管理
```

如果某些工具未来需要跨项目复用，可以再移动到项目根目录的 `pkg`。

### 6.10 migrations

保存数据库迁移文件。

推荐使用 SQL 文件，而不是一开始就做复杂迁移 DSL。

### 6.11 api

保存接口契约。

包括：

```text
1. OpenAPI 文档
2. Protobuf 文件
```

接口优先项目中，API 契约非常重要。

### 6.12 test

保存测试辅助代码和跨包测试。

普通单元测试也可以放在对应包目录下。

推荐：

```text
1. 单元测试靠近被测代码
2. 集成测试放在 test/integration
3. E2E 测试放在 test/e2e
```

---

## 7. 推荐架构风格

本项目推荐使用轻量级的：

```text
Clean Architecture + Hexagonal Architecture + Go Idioms
```

但不应教条化。

整体依赖方向如下：

```text
interfaces → usecase → domain
infrastructure → domain interface
app → 组装所有依赖
cmd → 启动具体程序
```

更具体地说：

```text
HTTP / gRPC / CLI
        ↓
Usecase
        ↓
Domain

Infrastructure 实现 Usecase 或 Domain 所需接口
App 负责把实现注入给 Usecase
```

依赖规则：

```text
1. Domain 不依赖 Usecase
2. Domain 不依赖 Infrastructure
3. Domain 不依赖 Interfaces
4. Usecase 可以依赖 Domain
5. Usecase 依赖抽象接口，而不是具体数据库实现
6. Interfaces 依赖 Usecase
7. Infrastructure 依赖 Domain 中定义的实体或接口
8. App 负责最终组装
```

---

## 8. 开发流程规范

本项目要求遵循如下开发流程：

```text
接口优先 → 测试优先 → 实现开发 → 重构优化
```

完整流程如下：

```text
1. 定义 API 契约
   OpenAPI / Proto / CLI Contract

2. 定义 Usecase 输入输出
   Input DTO / Output DTO

3. 定义 Usecase 需要的外部能力接口
   Repository / Hasher / Publisher / Mailer / Storage

4. 编写 Usecase 单元测试
   使用 Fake / Mock 隔离外部依赖

5. 实现 Domain 和 Usecase
   完成核心业务逻辑

6. 编写 Repository 集成测试
   验证数据库读写逻辑

7. 实现 Infrastructure
   MySQL / Redis / Queue / External API

8. 编写 Handler 测试
   验证 HTTP 状态码、响应结构、错误映射

9. 实现 Handler
   完成接口适配

10. 执行完整验证
    go test ./...
    go vet ./...
    golangci-lint run
```

---

## 9. 业务模块设计示例

以用户注册为例，推荐模块结构如下：

```text
internal/domain/user/
├── entity.go
├── repository.go
├── errors.go

internal/usecase/user/
├── register.go
├── dto.go
├── register_test.go

internal/infrastructure/persistence/mysql/
└── user_repository.go

internal/interfaces/http/handler/
└── user_handler.go
```

---

## 10. Domain 层设计

### 10.1 entity.go

```go
package user

import "time"

type User struct {
	ID        int64
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

func NewUser(name, email, hashedPassword string) (*User, error) {
	if name == "" {
		return nil, ErrInvalidName
	}

	if email == "" {
		return nil, ErrInvalidEmail
	}

	if hashedPassword == "" {
		return nil, ErrInvalidPassword
	}

	return &User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}, nil
}
```

### 10.2 repository.go

```go
package user

import "context"

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	Save(ctx context.Context, u *User) error
}
```

### 10.3 errors.go

```go
package user

import "errors"

var (
	ErrInvalidName        = errors.New("invalid user name")
	ErrInvalidEmail       = errors.New("invalid user email")
	ErrInvalidPassword    = errors.New("invalid user password")
	ErrEmailAlreadyExists = errors.New("email already exists")
)
```

---

## 11. Usecase 层设计

### 11.1 dto.go

```go
package user

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type RegisterOutput struct {
	ID    int64
	Name  string
	Email string
}
```

### 11.2 register.go

```go
package user

import (
	"context"

	domain "myapp/internal/domain/user"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
}

type RegisterUsecase struct {
	users  domain.Repository
	hasher PasswordHasher
}

func NewRegisterUsecase(
	users domain.Repository,
	hasher PasswordHasher,
) *RegisterUsecase {
	return &RegisterUsecase{
		users:  users,
		hasher: hasher,
	}
}

func (uc *RegisterUsecase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	exists, err := uc.users.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if exists != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	hashedPassword, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	u, err := domain.NewUser(input.Name, input.Email, hashedPassword)
	if err != nil {
		return nil, err
	}

	if err := uc.users.Save(ctx, u); err != nil {
		return nil, err
	}

	return &RegisterOutput{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}, nil
}
```

---

## 12. Usecase 单元测试设计

```go
package user_test

import (
	"context"
	"errors"
	"testing"

	domain "myapp/internal/domain/user"
	usecase "myapp/internal/usecase/user"
)

type fakeUserRepo struct {
	existing *domain.User
	saved    *domain.User
	err      error
}

func (r *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.existing, nil
}

func (r *fakeUserRepo) Save(ctx context.Context, u *domain.User) error {
	if r.err != nil {
		return r.err
	}

	u.ID = 1
	r.saved = u
	return nil
}

type fakeHasher struct {
	err error
}

func (h fakeHasher) Hash(password string) (string, error) {
	if h.err != nil {
		return "", h.err
	}

	return "hashed:" + password, nil
}

func TestRegisterUserSuccess(t *testing.T) {
	repo := &fakeUserRepo{}
	hasher := fakeHasher{}

	uc := usecase.NewRegisterUsecase(repo, hasher)

	out, err := uc.Execute(context.Background(), usecase.RegisterInput{
		Name:     "Jake",
		Email:    "jake@example.com",
		Password: "secret",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.ID != 1 {
		t.Fatalf("expected id 1, got %d", out.ID)
	}

	if out.Email != "jake@example.com" {
		t.Fatalf("expected email jake@example.com, got %s", out.Email)
	}

	if repo.saved == nil {
		t.Fatalf("expected user to be saved")
	}
}

func TestRegisterUserEmailAlreadyExists(t *testing.T) {
	repo := &fakeUserRepo{
		existing: &domain.User{
			ID:    1,
			Email: "jake@example.com",
		},
	}

	uc := usecase.NewRegisterUsecase(repo, fakeHasher{})

	_, err := uc.Execute(context.Background(), usecase.RegisterInput{
		Name:     "Jake",
		Email:    "jake@example.com",
		Password: "secret",
	})

	if !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}
```

---

## 13. Infrastructure 层设计

以 MySQL Repository 为例：

```go
package mysql

import (
	"context"
	"database/sql"
	"errors"

	domain "myapp/internal/domain/user"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password, created_at
		FROM users
		WHERE email = ?
	`, email)

	var u domain.User

	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) Save(ctx context.Context, u *domain.User) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users (name, email, password, created_at)
		VALUES (?, ?, ?, NOW())
	`, u.Name, u.Email, u.Password)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	u.ID = id
	return nil
}
```

---

## 14. Interfaces HTTP 层设计

HTTP 层只负责协议适配。

它不应该包含核心业务逻辑。

HTTP Handler 的职责包括：

```text
1. 读取请求参数
2. 执行参数校验
3. 调用 Usecase
4. 将业务错误映射为 HTTP 状态码
5. 返回统一响应
```

示例：

```go
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domain "myapp/internal/domain/user"
	usecase "myapp/internal/usecase/user"
	"myapp/internal/pkg/response"
)

type UserHandler struct {
	register *usecase.RegisterUsecase
}

func NewUserHandler(register *usecase.RegisterUsecase) *UserHandler {
	return &UserHandler{
		register: register,
	}
}

type RegisterUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.Error("VALIDATION_ERROR", err.Error()))
		return
	}

	out, err := h.register.Execute(c.Request.Context(), usecase.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		status, code := mapUserError(err)
		c.JSON(status, response.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(out))
}

func mapUserError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return http.StatusConflict, "EMAIL_ALREADY_EXISTS"
	case errors.Is(err, domain.ErrInvalidEmail):
		return http.StatusUnprocessableEntity, "INVALID_EMAIL"
	case errors.Is(err, domain.ErrInvalidName):
		return http.StatusUnprocessableEntity, "INVALID_NAME"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
```

---

## 15. 路由设计

```go
package http

import (
	"github.com/gin-gonic/gin"

	"myapp/internal/interfaces/http/handler"
	"myapp/internal/interfaces/http/middleware"
)

func NewRouter(userHandler *handler.UserHandler) *gin.Engine {
	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.Recover())
	r.Use(middleware.AccessLog())
	r.Use(middleware.Timeout())

	api := r.Group("/api")
	{
		api.POST("/users/register", userHandler.Register)
	}

	return r
}
```

---

## 16. 统一响应设计

```go
package response

type Body struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(data any) Body {
	return Body{
		Code:    "OK",
		Message: "success",
		Data:    data,
	}
}

func Error(code string, message string) Body {
	return Body{
		Code:    code,
		Message: message,
	}
}
```

成功响应示例：

```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "id": 1,
    "name": "Jake",
    "email": "jake@example.com"
  }
}
```

错误响应示例：

```json
{
  "code": "EMAIL_ALREADY_EXISTS",
  "message": "email already exists"
}
```

---

## 17. 配置设计

配置应使用结构体承载。

不建议业务代码直接读取环境变量。

```go
package config

import "os"

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Log      LogConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Addr string
}

type DatabaseConfig struct {
	Driver string
	DSN    string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type LogConfig struct {
	Level string
}

func Load() (Config, error) {
	return Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "myapp"),
			Env:  getEnv("APP_ENV", "local"),
		},
		HTTP: HTTPConfig{
			Addr: getEnv("HTTP_ADDR", ":8080"),
		},
		Database: DatabaseConfig{
			Driver: getEnv("DB_DRIVER", "mysql"),
			DSN:    getEnv("DB_DSN", ""),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
```

`.env.example` 示例：

```env
APP_NAME=myapp
APP_ENV=local

HTTP_ADDR=:8080

DB_DRIVER=mysql
DB_DSN=root:password@tcp(127.0.0.1:3306)/myapp?parseTime=true

REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=

LOG_LEVEL=info
```

---

## 18. 依赖注入设计

不建议默认使用运行时容器。

推荐使用：

```text
构造函数注入 + 手写依赖组装
```

中大型项目可以可选支持 Google Wire。

默认方式：

```go
package app

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"myapp/internal/config"
	mysqlrepo "myapp/internal/infrastructure/persistence/mysql"
	httpinterface "myapp/internal/interfaces/http"
	"myapp/internal/interfaces/http/handler"
	userusecase "myapp/internal/usecase/user"
	"myapp/internal/pkg/security"
)

type App struct {
	server *http.Server
	db     *sql.DB
	logger *slog.Logger
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := sql.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	userRepo := mysqlrepo.NewUserRepository(db)
	hasher := security.NewBcryptHasher()

	registerUser := userusecase.NewRegisterUsecase(userRepo, hasher)

	userHandler := handler.NewUserHandler(registerUser)

	router := httpinterface.NewRouter(userHandler)

	server := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: router,
	}

	return &App{
		server: server,
		db:     db,
		logger: logger,
	}, nil
}
```

优点：

```text
1. 依赖关系清晰
2. 不需要运行时反射
3. 编译期即可发现错误
4. 调试简单
5. 测试方便
```

---

## 19. 应用启动与优雅关闭

`cmd/api/main.go`：

```go
package main

import (
	"context"
	"log"

	"myapp/internal/app"
	"myapp/internal/config"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
```

`internal/app/app.go`：

```go
package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		a.logger.Info("http server started", "addr", a.server.Addr)

		err := a.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}

		errCh <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err

	case sig := <-quit:
		a.logger.Info("shutdown signal received", "signal", sig.String())

		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return err
		}

		if a.db != nil {
			if err := a.db.Close(); err != nil {
				return err
			}
		}

		a.logger.Info("application stopped")
		return nil
	}
}
```

---

## 20. 数据库迁移设计

推荐使用 SQL 迁移文件。

目录：

```text
migrations/
├── 20260603120000_create_users_table.up.sql
└── 20260603120000_create_users_table.down.sql
```

`up.sql`：

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

`down.sql`：

```sql
DROP TABLE IF EXISTS users;
```

命令：

```bash
make migrate-up
make migrate-down
```

或：

```bash
gos migrate up
gos migrate down
```

---

## 21. 事务管理设计

事务边界应由 Usecase 控制，而不是隐藏在 Repository 内部。

定义事务接口：

```go
package transaction

import "context"

type Manager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

使用示例：

```go
func (uc *CreateOrderUsecase) Execute(ctx context.Context, input CreateOrderInput) error {
	return uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.orders.Save(ctx, order); err != nil {
			return err
		}

		if err := uc.inventory.Decrease(ctx, input.ProductID, input.Quantity); err != nil {
			return err
		}

		return nil
	})
}
```

设计原则：

```text
1. Usecase 决定事务边界
2. Repository 只负责数据操作
3. 不在 HTTP Handler 中开启事务
4. 不在 Domain 中感知事务
```

默认的嵌套事务语义是复用外层事务：如果 `ctx` 中已经存在事务，内层 `WithinTx` 不会再次开启数据库事务，而是直接执行回调，并由最外层统一 commit 或 rollback。

生成项目支持通过 `enableNestedTransaction` 配置开启基于 savepoint 的嵌套事务。对应环境变量为：

```text
DB_ENABLE_NESTED_TRANSACTION=false
```

行为：

```text
1. enableNestedTransaction=false 时保持默认行为，内层事务复用外层事务
2. enableNestedTransaction=true 时，事务内再次调用 WithinTx 会创建 SAVEPOINT
3. 内层成功时 RELEASE SAVEPOINT
4. 内层失败时 ROLLBACK TO SAVEPOINT
5. 最终持久化仍由最外层事务 commit 决定
```

该能力适用于需要局部回滚但不希望中断整个业务事务的场景。默认关闭，以保持事务语义简单、可预测。

---

## 22. 日志设计

默认推荐使用 Go 标准库 `slog`。

日志规范：

```text
1. HTTP 中间件记录请求日志
2. Recover 中间件记录 panic
3. Infrastructure 层记录外部系统异常
4. Usecase 层只记录关键业务行为
5. 日志中应包含 request_id；启用 OpenTelemetry 时自动补充 trace_id 和 span_id
6. 不在日志中记录明文密码、Token、密钥等敏感信息
```

当前生成项目会生成 `internal/logging` 包，启动阶段读取 `LOG_LEVEL` 创建 `slog` JSON logger，并通过 `slog.SetDefault` 让 HTTP 中间件、生成命令和业务代码共享同一套日志配置。`LOG_LEVEL` 支持 `debug`、`info`、`warn`、`warning`、`error`。

推荐中间件：

```text
1. RequestID
2. AccessLog
3. Recover
4. Timeout
5. CORS
6. RateLimit
7. Auth
```

---

## 23. 错误处理设计

错误分为三类：

```text
1. Domain Error
   领域错误，例如邮箱已存在、余额不足

2. Application Error
   应用错误，例如权限不足、参数无效

3. Infrastructure Error
   基础设施错误，例如数据库连接失败、Redis 超时
```

错误处理原则：

```text
1. Domain 和 Usecase 返回业务错误
2. HTTP 层负责将错误映射为 HTTP 状态码
3. gRPC 层负责将错误映射为 gRPC Status
4. CLI 层负责将错误映射为终端输出
5. 不在 Usecase 中直接返回 HTTP 状态码
6. 不在 Domain 中依赖 HTTP 或 gRPC
```

HTTP 错误映射示例：

```go
func MapErrorStatus(err error) int {
	switch {
	case errors.Is(err, user.ErrEmailAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, user.ErrInvalidEmail):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
```

---

## 24. 并发与后台任务设计

Go 的并发能力是重要优势。

脚手架应内置后台任务模板，但不要把本地 goroutine 伪装成可靠队列。

建议提供三类能力：

```text
1. Worker
2. Scheduler
3. Async Task
```

### 24.1 Worker

用于处理：

```text
1. 消息队列
2. 邮件发送
3. 异步通知
4. 文件处理
5. 日志消费
```

入口：

```text
cmd/worker/main.go
```

### 24.2 Scheduler

用于处理：

```text
1. 定时任务
2. 数据同步
3. 报表生成
4. 过期订单处理
```

### 24.3 Async Task

本地轻量异步任务示例：

```go
package task

import (
	"context"
	"sync"
)

type Runner struct {
	wg sync.WaitGroup
}

func (r *Runner) Go(ctx context.Context, fn func(ctx context.Context) error) {
	r.wg.Add(1)

	go func() {
		defer r.wg.Done()
		_ = fn(ctx)
	}()
}

func (r *Runner) Wait() {
	r.wg.Wait()
}
```

注意：

```text
本地 goroutine 不等于可靠队列。
重要任务应使用 Redis、Kafka、RabbitMQ 等可靠消息系统。
```

---

## 25. API 契约设计

接口优先项目中，应优先定义接口契约。

推荐使用：

```text
1. OpenAPI
2. Protobuf
```

HTTP API 项目默认生成：

```text
api/openapi.yaml
```

流程：

```text
1. 先定义 OpenAPI
2. 再定义 HTTP Request / Response DTO
3. 再编写 Handler 测试
4. 最后实现 Handler
```

内部 RPC 服务可以使用：

```text
api/proto/*.proto
```

---

## 26. CLI 命令设计

脚手架命令不应过度复杂。

当前已支持：

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

示例：

```bash
gos new blog --module=example.com/blog --template=api-clean
```

```bash
gos new tiny-api --module=example.com/tiny-api --template=api-minimal
```

```bash
gos new traced-api --module=example.com/traced-api --with-otel
```

```bash
gos make:usecase user/register
```

```bash
gos make:handler user --register --openapi
```

```bash
gos make:model invoice --fields=number:string:json=invoice_number,total:int64,created_at:time --openapi
```

```bash
gos make:repository user --db=mysql --fields=email:string:unique,size=320,created_at:time --with-migration --register --openapi
```

```bash
gos make:command sync-orders --register
go run ./cmd/api sync-orders
```

### 26.1 new 命令

创建新项目。

```bash
gos new myapp
```

可选参数：

```bash
--template=api-clean
--template=api-minimal
--with-otel
--module=example.com/myapp
--force
--dry-run
```

### 26.2 make:usecase 命令

生成 Usecase 和测试骨架。

```bash
gos make:usecase user/register
```

生成：

```text
internal/usecase/user/register.go
internal/usecase/user/register_test.go
```

### 26.3 make:handler 命令

生成 HTTP Handler。

```bash
gos make:handler user --register --openapi
```

生成：

```text
internal/interfaces/http/handler/user_handler.go
internal/interfaces/http/handler/user_handler_test.go
internal/interfaces/http/router.go（使用 --register 时更新）
api/openapi.yaml（使用 --openapi 时更新）
```

### 26.4 make:repository 命令

生成 Repository 实现。

```bash
gos make:repository user --db=mysql --fields=email:string:unique,size=320,created_at:time --with-migration --register --openapi
```

生成：

```text
internal/domain/user/entity.go
internal/domain/user/repository.go
internal/infrastructure/persistence/mysql/user_repository.go
internal/infrastructure/persistence/mysql/user_repository_test.go
internal/infrastructure/persistence/mysql/user_repository_integration_test.go
migrations/<timestamp>_create_users_table.up.sql
migrations/<timestamp>_create_users_table.down.sql
internal/app/assembly.go（使用 --register 时更新）
```

Repository integration tests use the `integration` build tag and `TEST_DATABASE_DSN`.

```bash
TEST_DATABASE_DSN='root:password@tcp(127.0.0.1:3307)/myapp_test?parseTime=true' go test -tags=integration ./internal/infrastructure/persistence/mysql
```

Without `TEST_DATABASE_DSN`, the integration test is skipped.

Generated projects include `deployments/docker/docker-compose.test.yml` for a local MySQL test database:

```bash
docker compose -f deployments/docker/docker-compose.test.yml up -d
TEST_DATABASE_DSN='root:password@tcp(127.0.0.1:3307)/myapp_test?parseTime=true' go test -tags=integration ./internal/infrastructure/persistence/mysql
```

### 26.5 make:model 命令

生成 Domain Entity。

```bash
gos make:model invoice --fields=number:string:json=invoice_number,total:int64,created_at:time --openapi
```

生成：

```text
internal/domain/invoice/entity.go
```

字段 DSL 与 `make:repository --fields` 保持一致。默认不覆盖已存在的 entity 文件，需要覆盖时显式使用 `--force`。

### 26.6 make:migration 命令

生成数据库迁移文件。

```bash
gos make:migration create_users_table
```

生成：

```text
migrations/20260603120000_create_users_table.up.sql
migrations/20260603120000_create_users_table.down.sql
```

### 26.7 make:command 命令

生成可从 `cmd/api` 执行的 Cobra 命令脚本。

```bash
gos make:command sync-orders --register
go run ./cmd/api sync-orders
```

生成：

```text
internal/command/sync_orders.go
internal/command/sync_orders_test.go
```

`--register` 会更新标准 `cmd/api/main.go` 中的 Cobra root command；默认生成项目已内置 `serve`、`schedule`、`queue` 三个子命令。

---

## 27. 代码生成原则

代码生成应遵循以下原则：

```text
1. 只生成必要代码
2. 生成代码必须清晰可读
3. 生成代码必须可以自由修改
4. 不生成大量无意义空文件
5. 不强制业务逻辑写法
6. 不隐藏关键依赖
7. 生成后自动执行 gofmt
8. 文件存在时应提示冲突或允许覆盖
```

脚手架应辅助开发，而不是绑架开发。

---

## 28. 模板设计

当前内置模板：

```text
1. api-clean
   默认完整 API 模板

2. api-minimal
   极简标准库 HTTP 模板
```

模板目录建议：

```text
templates/
├── api-clean/
│   ├── project/
│   ├── cmd/
│   ├── internal/
│   ├── migrations/
│   ├── api/
│   └── deployments/
│
├── api-basic/
├── grpc-service/
├── worker/
└── monolith/
```

推荐内置模板：

```text
1. api-clean
   适合中大型 REST API 项目

2. api-basic
   适合小型 API 项目

3. grpc-service
   适合内部 RPC 服务

4. worker
   适合后台任务服务

5. monolith
   适合单体业务系统
```

---

## 29. 默认技术选型

默认推荐组合：

```text
HTTP: Gin 或 Chi
DB: sqlc / sqlx / GORM 可选
Migration: golang-migrate
Config: env + struct
Logger: slog
Validation: go-playground/validator
DI: 手写构造函数，可选 Wire
Testing: testing
Docs: OpenAPI
Container: Docker Compose
```

推荐模式：

```text
快速开发项目：
Gin + GORM + slog

企业级项目：
Chi + sqlc + slog + Wire

通用折中项目：
Gin + sqlx + slog
```

---

## 30. Makefile 设计

生成项目应内置 Makefile。

```makefile
run:
	go run ./cmd/api

worker:
	go run ./cmd/worker

test:
	go test ./...

test-unit:
	go test ./internal/... ./test/unit/...

test-integration:
	go test ./test/integration/...

lint:
	golangci-lint run

vet:
	go vet ./...

build:
	go build -o bin/api ./cmd/api

migrate-up:
	migrate -path migrations -database "$$DB_DSN" up

migrate-down:
	migrate -path migrations -database "$$DB_DSN" down

docker-up:
	docker compose -f deployments/docker/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker/docker-compose.yml down
```

---

## 31. Docker Compose 设计

开发环境应提供基础 Docker Compose。

```yaml
services:
  mysql:
    image: mysql:8
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: myapp
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7
    ports:
      - "6379:6379"

volumes:
  mysql_data:
```

---

## 32. 测试体系设计

测试分为：

```text
1. Unit Test
   单元测试，不依赖数据库、Redis、HTTP Server

2. Integration Test
   集成测试，依赖数据库、Redis、消息队列等

3. E2E Test
   端到端测试，测试完整接口流程
```

推荐规则：

```text
1. Usecase 必须有单元测试
2. Repository 必须有集成测试
3. Handler 应有功能测试
4. 复杂 Domain 规则必须有单元测试
5. 外部 API Client 应通过接口隔离并 Mock
```

测试执行：

```bash
make test
make test-unit
make test-integration
```

---

## 33. CI 设计

推荐内置 GitHub Actions 模板。

```yaml
name: CI

on:
  push:
    branches:
      - main
  pull_request:

permissions:
  contents: read

jobs:
  verify:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v5

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - name: Check formatting
        run: |
          unformatted="$(gofmt -l .)"
          if [ -n "$unformatted" ]; then
            echo "$unformatted"
            exit 1
          fi

      - name: Vet
        run: go vet ./...

      - name: Test
        run: go test ./...

      - name: Build
        run: go build ./cmd/api
```

---

## 34. 分阶段实施计划

### 34.1 第一阶段：MVP

目标：完成可用的脚手架最小版本。

功能：

```text
1. gos new
2. api-clean 项目模板
3. 基础目录结构
4. 配置加载模板
5. HTTP Server 模板
6. Router 模板
7. Usecase 模板
8. Handler 模板
9. Repository 模板
10. 单元测试模板
11. Makefile
12. Docker Compose
```

### 34.2 第二阶段：工程能力增强

功能：

```text
1. make:usecase
2. make:handler
3. make:repository
4. make:migration
5. make:test
6. 统一响应
7. 统一错误映射
8. 日志中间件
9. Recover 中间件
10. RequestID 中间件
11. 优雅关闭
```

### 34.3 第三阶段：测试与契约增强

功能：

```text
1. OpenAPI 模板
2. Handler 测试模板
3. Repository 集成测试模板
4. 测试数据库初始化
5. CI 模板
6. Lint 配置
```

### 34.4 第四阶段：多模板支持

当前已完成内置模板发现、未知模板校验，以及 `api-clean` / `api-minimal` 两个模板。下面列表保留为后续模板扩展方向。

功能：

```text
1. api-basic 模板
2. grpc-service 模板
3. worker 模板
4. monolith 模板
5. 不同数据库方案支持
6. 不同 HTTP 框架支持
```

### 34.5 第五阶段：高级能力

功能：

```text
1. Wire 支持
2. Queue 模板
3. Scheduler 模板
4. Auth 模板
5. 权限模板
6. 多租户模板
7. 插件机制
```

---

## 35. 不推荐设计

本项目不建议实现以下内容：

```text
1. 重型运行时 Service Container
2. 字符串式依赖解析
3. 大量反射注入
4. 自研大型 ORM
5. 自研大型 Web 框架
6. 过度封装标准库
7. 强制所有项目使用同一种目录
8. 生成大量无法维护的样板代码
9. 在业务代码中直接读取环境变量
10. 在 Usecase 或 Domain 中耦合 HTTP 状态码
```

---

## 36. 最终架构总结

最终推荐架构如下：

```text
cmd
 ├── api
 ├── worker
 └── migrate

internal
 ├── app
 ├── config
 ├── domain
 ├── usecase
 ├── interfaces
 ├── infrastructure
 └── pkg

api
 ├── openapi.yaml
 └── proto

migrations
test
deployments
```

核心依赖关系：

```text
interfaces → usecase → domain

infrastructure → domain interface

app → 负责组装所有依赖

cmd → 负责启动具体程序
```

---

## 37. 最终结论

本项目应设计为：

```text
一个 Go 原生优先、接口优先、测试优先、轻量分层、显式依赖的后端工程脚手架。
```

项目应重点提供：

```text
1. 标准项目结构
2. 代码生成工具
3. 测试优先流程
4. 接口契约管理
5. 配置加载模板
6. 日志与错误处理模板
7. 数据库迁移模板
8. 本地开发环境模板
9. CI 模板
10. 可替换的技术选型
```

一句话概括：

```text
这是一个面向 Go 后端项目的工程规范与代码生成工具，而不是一个重型运行时框架。
```
