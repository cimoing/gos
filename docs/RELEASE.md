# 发布说明

本文记录 `gos` CLI 的本地构建、版本注入、补全脚本和发布前检查流程。

## 本地构建

默认构建：

```bash
go build -o bin/gos ./cmd/gos
```

使用 Makefile 构建：

```bash
make build
```

`make build` 会通过 ldflags 注入：

```text
Version
Commit
BuildDate
```

手动注入示例：

```bash
go build -ldflags "-X github.com/cimoing/gos/internal/command.Version=v0.1.0 -X github.com/cimoing/gos/internal/command.Commit=abc1234 -X github.com/cimoing/gos/internal/command.BuildDate=2026-06-09T00:00:00Z" -o bin/gos ./cmd/gos
```

## 版本检查

```bash
gos version
```

开发构建默认输出：

```text
gos dev
commit none
built unknown
```

## Shell Completion

生成 Bash completion：

```bash
gos completion bash > gos.bash
```

生成 Zsh completion：

```bash
gos completion zsh > _gos
```

生成 Fish completion：

```bash
gos completion fish > gos.fish
```

生成 PowerShell completion：

```powershell
gos completion powershell > gos.ps1
```

团队可以将这些文件随二进制发布，或在安装脚本中按目标 shell 生成。

## 发布前检查

```bash
go test ./...
go run ./cmd/gos version
go run ./cmd/gos completion bash
```

`go test ./...` 会覆盖：

```text
1. gos CLI 命令测试。
2. api-clean 生成项目 go test ./... 和 go build ./cmd/api。
3. api-clean --with-otel 生成项目 go test ./... 和 go build ./cmd/api。
4. api-minimal 生成项目 go test ./... 和 go build ./cmd/api。
5. api-minimal --with-otel 生成项目 go test ./... 和 go build ./cmd/api。
```

## 后续可选增强

```text
1. 增加 GoReleaser 配置。
2. 为不同 OS/ARCH 生成压缩包和校验和。
3. 将 completion 文件打进 release artifacts。
4. 自动生成 changelog。
```
