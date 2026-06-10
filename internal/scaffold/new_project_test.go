package scaffold

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/cimoing/gos/internal/generator"
)

func TestProjectGeneratorGenerateAPIClean(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo")
	gen := NewProjectGenerator(generator.Default())

	result, err := gen.Generate(context.Background(), NewProjectOptions{
		ProjectName: "demo",
		ModulePath:  "example.com/demo",
		Template:    "api-clean",
		TargetDir:   root,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Created) == 0 {
		t.Fatalf("created files is empty")
	}

	wantFiles := []string{
		".env.example",
		".github/workflows/ci.yml",
		".gitignore",
		"Makefile",
		"README.md",
		"api/openapi.yaml",
		"cmd/api/main.go",
		"deployments/docker/Dockerfile",
		"deployments/docker/docker-compose.test.yml",
		"deployments/docker/docker-compose.yml",
		"go.mod",
		"go.sum",
		"internal/app/app.go",
		"internal/app/assembly.go",
		"internal/config/config.go",
		"internal/infrastructure/database/database.go",
		"internal/infrastructure/database/database_test.go",
		"internal/infrastructure/database/mysql.go",
		"internal/infrastructure/database/transaction.go",
		"internal/interfaces/http/httperror/mapper.go",
		"internal/interfaces/http/httperror/mapper_test.go",
		"internal/interfaces/http/middleware/access_log.go",
		"internal/interfaces/http/middleware/chain.go",
		"internal/interfaces/http/middleware/cors.go",
		"internal/interfaces/http/middleware/middleware_test.go",
		"internal/interfaces/http/middleware/recover.go",
		"internal/interfaces/http/middleware/request_id.go",
		"internal/interfaces/http/middleware/timeout.go",
		"internal/interfaces/http/router.go",
		"internal/logging/logging.go",
		"internal/logging/logging_test.go",
		"internal/pkg/apperror/error.go",
		"internal/pkg/response/response.go",
		"internal/usecase/user/register.go",
		"internal/usecase/user/register_test.go",
		"internal/worker/worker.go",
		"internal/worker/worker_test.go",
	}

	gotFiles := listGeneratedFiles(t, root)
	if !reflect.DeepEqual(gotFiles, wantFiles) {
		t.Fatalf("generated files mismatch\nwant:\n%s\n\ngot:\n%s", strings.Join(wantFiles, "\n"), strings.Join(gotFiles, "\n"))
	}

	for _, path := range wantFiles {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}

	assertFileContains(t, root, "internal/interfaces/http/router.go", "middleware.RequestID()")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "middleware.CORS(opts.CORS)")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "middleware.Timeout(10*time.Second)")
	assertFileContains(t, root, "cmd/api/main.go", "github.com/spf13/cobra")
	assertFileContains(t, root, "cmd/api/main.go", "func newRootCommand(ctx context.Context) *cobra.Command")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newServeCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newScheduleCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newQueueCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "gos:command-imports")
	assertFileContains(t, root, "cmd/api/main.go", "gos:commands")
	assertFileContains(t, root, "cmd/api/main.go", "worker.NewScheduler")
	assertFileContains(t, root, "cmd/api/main.go", "worker.NewQueueWorker")
	assertFileContains(t, root, "internal/worker/worker.go", "type Scheduler struct")
	assertFileContains(t, root, "internal/worker/worker.go", "type QueueWorker struct")
	assertFileContains(t, root, "internal/worker/worker.go", "func runSafely")
	assertFileContains(t, root, "internal/worker/worker_test.go", "TestRunSafelyRecoversPanic")
	assertFileContains(t, root, "deployments/docker/docker-compose.yml", "mysql:")
	assertFileContains(t, root, "deployments/docker/docker-compose.yml", "redis:")
	assertFileContains(t, root, "deployments/docker/docker-compose.test.yml", "mysql_test:")
	assertFileContains(t, root, "deployments/docker/docker-compose.test.yml", "3307:3306")
	assertFileContains(t, root, ".env.example", "DB_DSN=")
	assertFileContains(t, root, ".env.example", "DB_ENABLE_NESTED_TRANSACTION=false")
	assertFileContains(t, root, ".env.example", "REDIS_ADDR=")
	assertFileContains(t, root, ".env.example", "LOG_LEVEL=info")
	assertFileContains(t, root, ".env.example", "HTTP_READ_HEADER_TIMEOUT=5s")
	assertFileContains(t, root, ".env.example", "HTTP_READ_TIMEOUT=15s")
	assertFileContains(t, root, ".env.example", "HTTP_WRITE_TIMEOUT=30s")
	assertFileContains(t, root, ".env.example", "HTTP_IDLE_TIMEOUT=60s")
	assertFileContains(t, root, ".env.example", "HTTP_MAX_HEADER_BYTES=1048576")
	assertFileContains(t, root, ".env.example", "HTTP_MAX_BODY_BYTES=10485760")
	assertFileContains(t, root, ".env.example", "CORS_ALLOWED_ORIGINS=*")
	assertFileContains(t, root, ".env.example", "CORS_ALLOW_CREDENTIALS=false")
	assertFileContains(t, root, ".env.example", "CORS_MAX_AGE=600")
	assertFileContains(t, root, "internal/config/config.go", "ReadHeaderTimeout time.Duration")
	assertFileContains(t, root, "internal/config/config.go", "MaxBodyBytes      int64")
	assertFileContains(t, root, "internal/config/config.go", "type CORSConfig struct")
	assertFileContains(t, root, "internal/config/config.go", `getEnvCSV("CORS_ALLOWED_ORIGINS", []string{"*"})`)
	assertFileContains(t, root, "internal/config/config.go", `corsAllowCredentials, err := getEnvBool("CORS_ALLOW_CREDENTIALS", false)`)
	assertFileContains(t, root, "internal/config/config.go", `corsMaxAge, err := getEnvInt("CORS_MAX_AGE", 600)`)
	assertFileContains(t, root, "internal/config/config.go", `httpReadTimeout, err := getEnvDuration("HTTP_READ_TIMEOUT", 15*time.Second)`)
	assertFileContains(t, root, "internal/config/config.go", `httpMaxHeaderBytes, err := getEnvInt("HTTP_MAX_HEADER_BYTES", 1<<20)`)
	assertFileContains(t, root, "internal/config/config.go", `httpMaxBodyBytes, err := getEnvInt64("HTTP_MAX_BODY_BYTES", 10<<20)`)
	assertFileContains(t, root, "internal/config/config.go", `return 0, fmt.Errorf("parse %s as duration: %w", key, err)`)
	assertFileContains(t, root, "internal/config/config.go", `return 0, fmt.Errorf("parse %s as int64: %w", key, err)`)
	assertFileContains(t, root, "internal/config/config.go", "EnableNestedTransaction bool")
	assertFileContains(t, root, "internal/config/config.go", `enableNestedTransaction, err := getEnvBool("DB_ENABLE_NESTED_TRANSACTION", false)`)
	assertFileContains(t, root, "internal/config/config.go", `return false, fmt.Errorf("parse %s as bool: %w", key, err)`)
	assertFileContains(t, root, "internal/config/config.go", `return 0, fmt.Errorf("parse %s as int: %w", key, err)`)
	assertFileContains(t, root, "internal/logging/logging.go", "func New(cfg config.LogConfig) (*slog.Logger, error)")
	assertFileContains(t, root, "internal/logging/logging.go", `case "debug":`)
	assertFileContains(t, root, "internal/logging/logging.go", "ReplaceAttr: redactAttr")
	assertFileContains(t, root, "internal/logging/logging.go", `return slog.String(attr.Key, "[REDACTED]")`)
	assertFileContains(t, root, "internal/logging/logging_test.go", "TestLoggerRedactsSensitiveFields")
	assertFileContains(t, root, "internal/app/app.go", "logging.New(cfg.Log)")
	assertFileContains(t, root, "internal/app/app.go", "slog.SetDefault(logger)")
	assertFileContains(t, root, "internal/app/app.go", "ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout")
	assertFileContains(t, root, "internal/app/app.go", "WriteTimeout:      cfg.HTTP.WriteTimeout")
	assertFileContains(t, root, "internal/app/app.go", "MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes")
	assertFileContains(t, root, "internal/app/app.go", "MaxBodyBytes: cfg.HTTP.MaxBodyBytes")
	assertFileContains(t, root, "internal/app/app.go", "AllowedOrigins:   cfg.CORS.AllowedOrigins")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "MaxBodyBytes int64")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "CORS         middleware.CORSOptions")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "http.MaxBytesHandler(handler, opts.MaxBodyBytes)")
	assertFileContains(t, root, "cmd/api/main.go", "func configureLogging() error")
	assertFileContains(t, root, "cmd/api/main.go", "PersistentPreRunE")
	assertFileNotContains(t, root, "internal/logging/logging.go", "trace_id")
	assertFileContains(t, root, "go.mod", "github.com/go-sql-driver/mysql v1.10.0")
	assertFileContains(t, root, "go.mod", "github.com/spf13/cobra v1.10.2")
	assertFileContains(t, root, "go.mod", "filippo.io/edwards25519 v1.2.0 // indirect")
	assertFileContains(t, root, "go.sum", "github.com/go-sql-driver/mysql v1.10.0 h1:")
	assertFileContains(t, root, "go.sum", "github.com/spf13/cobra v1.10.2 h1:")
	assertFileContains(t, root, "go.sum", "github.com/spf13/pflag v1.0.9 h1:")
	assertFileContains(t, root, "go.sum", "filippo.io/edwards25519 v1.2.0 h1:")
	assertFileNotContains(t, root, "go.mod", "go.opentelemetry.io/otel")
	assertFileNotContains(t, root, "go.mod", "github.com/XSAM/otelsql")
	assertPathNotExists(t, root, "internal/observability")
	assertFileContains(t, root, "README.md", "github.com/go-sql-driver/mysql")
	assertFileContains(t, root, "README.md", ".github/workflows/ci.yml")
	assertFileContains(t, root, ".github/workflows/ci.yml", "uses: actions/checkout@v5")
	assertFileContains(t, root, ".github/workflows/ci.yml", "uses: actions/setup-go@v6")
	assertFileContains(t, root, ".github/workflows/ci.yml", "go-version-file: go.mod")
	assertFileContains(t, root, ".github/workflows/ci.yml", "go vet ./...")
	assertFileContains(t, root, ".github/workflows/ci.yml", "TEST_DATABASE_DSN:")
	assertFileContains(t, root, ".github/workflows/ci.yml", "go test -tags=integration ./internal/infrastructure/persistence/mysql")
	assertPathNotExists(t, root, "scripts")
	assertFileContains(t, root, "internal/app/app.go", "BuildDependencies(ctx, cfg)")
	assertFileContains(t, root, "internal/app/app.go", "dependencies.Close()")
	assertFileContains(t, root, "internal/app/assembly.go", "type Dependencies struct")
	assertFileContains(t, root, "internal/app/assembly.go", "DB           *sql.DB")
	assertFileContains(t, root, "internal/app/assembly.go", "Transactions *database.TxManager")
	assertFileContains(t, root, "internal/app/assembly.go", "database.Open(ctx, cfg.Database)")
	assertFileContains(t, root, "internal/app/assembly.go", "database.NewTxManager(db, database.TxOptions{")
	assertFileContains(t, root, "internal/app/assembly.go", "EnableNestedTransaction: cfg.Database.EnableNestedTransaction")
	assertFileContains(t, root, "internal/app/assembly.go", "gos:dependencies")
	assertFileContains(t, root, "internal/app/assembly.go", "gos:assemble")
	assertFileContains(t, root, "internal/app/assembly.go", "gos:return-dependencies")
	assertFileContains(t, root, "internal/infrastructure/database/database.go", "func Open(ctx context.Context, cfg config.DatabaseConfig)")
	assertFileContains(t, root, "internal/infrastructure/database/database.go", "return sql.Open(cfg.Driver, cfg.DSN)")
	assertFileContains(t, root, "internal/infrastructure/database/database_test.go", "TestOpenSkipsEmptyDSN")
	assertFileContains(t, root, "internal/infrastructure/database/database_test.go", "TestTxManagerRejectsMissingDatabase")
	assertFileContains(t, root, "internal/infrastructure/database/database_test.go", "TestNewTxManagerStoresOptions")
	assertFileContains(t, root, "internal/infrastructure/database/mysql.go", `_ "github.com/go-sql-driver/mysql"`)
	assertFileContains(t, root, "internal/infrastructure/database/transaction.go", "func (m *TxManager) WithinTx")
	assertFileContains(t, root, "internal/infrastructure/database/transaction.go", "func ExecutorFromContext")
	assertFileContains(t, root, "internal/infrastructure/database/transaction.go", "SAVEPOINT ")
	assertFileContains(t, root, "internal/infrastructure/database/transaction.go", "ROLLBACK TO SAVEPOINT ")
	assertFileContains(t, root, "internal/infrastructure/database/transaction.go", "RELEASE SAVEPOINT ")
	assertFileContains(t, root, "internal/interfaces/http/middleware/cors.go", "func CORS(options CORSOptions) Middleware")
	assertFileContains(t, root, "internal/interfaces/http/middleware/recover.go", `"panic_type"`)
	assertFileContains(t, root, "internal/interfaces/http/middleware/middleware_test.go", "TestCORSHandlesPreflight")
	assertFileContains(t, root, "api/openapi.yaml", "openapi: 3.0.3")
	assertFileContains(t, root, "api/openapi.yaml", "/healthz:")
	assertFileContains(t, root, "api/openapi.yaml", "responses:")
	assertFileContains(t, root, "api/openapi.yaml", "BadRequest:")
	assertFileContains(t, root, "api/openapi.yaml", "InternalServerError:")
	assertFileContains(t, root, "api/openapi.yaml", "ListResponse:")
	assertFileContains(t, root, "api/openapi.yaml", "examples:")
}

func TestProjectGeneratorGenerateAPICleanWithOpenTelemetry(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo")
	gen := NewProjectGenerator(generator.Default())

	_, err := gen.Generate(context.Background(), NewProjectOptions{
		ProjectName:       "demo",
		ModulePath:        "example.com/demo",
		Template:          "api-clean",
		TargetDir:         root,
		WithOpenTelemetry: true,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	assertFileContains(t, root, "go.mod", "go.opentelemetry.io/otel v1.43.0")
	assertFileContains(t, root, "go.mod", "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.68.0")
	assertFileContains(t, root, "go.mod", "github.com/XSAM/otelsql v0.42.0")
	assertFileContains(t, root, "go.sum", "github.com/XSAM/otelsql v0.42.0 h1:")
	assertFileContains(t, root, ".env.example", "OTEL_ENABLED=false")
	assertFileContains(t, root, "internal/config/config.go", "type ObservabilityConfig struct")
	assertFileContains(t, root, "internal/app/app.go", "observability.SetupOpenTelemetry(ctx, cfg.Observability)")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "otelhttp.NewHandler(handler, opts.ServiceName)")
	assertFileContains(t, root, "internal/observability/otel.go", "func SetupOpenTelemetry")
	assertFileContains(t, root, "internal/observability/otel.go", "otlptracehttp.New")
	assertFileContains(t, root, "internal/observability/http_client.go", "func NewHTTPClient")
	assertFileContains(t, root, "internal/observability/http_client.go", "otelhttp.NewTransport(base)")
	assertFileContains(t, root, "internal/infrastructure/database/database.go", "otelsql.Open(cfg.Driver, cfg.DSN")
	assertFileContains(t, root, "internal/infrastructure/database/database.go", `attribute.String("db.system.name", cfg.Driver)`)
	assertFileContains(t, root, "internal/logging/logging.go", "trace.SpanContextFromContext(ctx)")
	assertFileContains(t, root, "internal/logging/logging.go", `"trace_id"`)
	assertFileContains(t, root, "internal/logging/logging.go", `"span_id"`)
}

func TestProjectGeneratorGenerateAPIMinimal(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo")
	gen := NewProjectGenerator(generator.Default())

	result, err := gen.Generate(context.Background(), NewProjectOptions{
		ProjectName: "demo",
		ModulePath:  "example.com/demo",
		Template:    "api-minimal",
		TargetDir:   root,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Created) == 0 {
		t.Fatalf("created files is empty")
	}

	wantFiles := []string{
		".gitignore",
		"Makefile",
		"README.md",
		"cmd/api/main.go",
		"go.mod",
		"go.sum",
		"internal/config/config.go",
		"internal/interfaces/http/router.go",
		"internal/interfaces/http/router_test.go",
		"internal/logging/logging.go",
		"internal/logging/logging_test.go",
		"internal/worker/worker.go",
		"internal/worker/worker_test.go",
	}

	gotFiles := listGeneratedFiles(t, root)
	if !reflect.DeepEqual(gotFiles, wantFiles) {
		t.Fatalf("generated files mismatch\nwant:\n%s\n\ngot:\n%s", strings.Join(wantFiles, "\n"), strings.Join(gotFiles, "\n"))
	}

	assertFileContains(t, root, "README.md", "api-minimal")
	assertFileContains(t, root, "cmd/api/main.go", "httpinterface.NewRouter(httpinterface.RouterOptions{")
	assertFileContains(t, root, "cmd/api/main.go", "github.com/spf13/cobra")
	assertFileContains(t, root, "cmd/api/main.go", "func newRootCommand(ctx context.Context) *cobra.Command")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newServeCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newScheduleCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newQueueCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "gos:command-imports")
	assertFileContains(t, root, "cmd/api/main.go", "gos:commands")
	assertFileContains(t, root, "cmd/api/main.go", "worker.NewScheduler")
	assertFileContains(t, root, "cmd/api/main.go", "worker.NewQueueWorker")
	assertFileContains(t, root, "internal/worker/worker.go", "type Scheduler struct")
	assertFileContains(t, root, "internal/worker/worker.go", "type QueueWorker struct")
	assertFileContains(t, root, "internal/worker/worker_test.go", "TestQueueWorkerRunsProcessor")
	assertFileContains(t, root, "go.mod", "github.com/spf13/cobra v1.10.2")
	assertFileContains(t, root, "go.sum", "github.com/spf13/cobra v1.10.2 h1:")
	assertFileNotContains(t, root, "go.mod", "go.opentelemetry.io/otel")
	assertPathNotExists(t, root, "internal/observability")
	assertFileContains(t, root, "internal/config/config.go", `getEnv("APP_NAME", "demo")`)
	assertFileContains(t, root, "internal/config/config.go", "func Load() (Config, error)")
	assertFileContains(t, root, "internal/config/config.go", "type HTTPConfig struct")
	assertFileContains(t, root, "internal/config/config.go", `httpReadHeaderTimeout, err := getEnvDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)`)
	assertFileContains(t, root, "internal/config/config.go", `httpMaxHeaderBytes, err := getEnvInt("HTTP_MAX_HEADER_BYTES", 1<<20)`)
	assertFileContains(t, root, "internal/config/config.go", `httpMaxBodyBytes, err := getEnvInt64("HTTP_MAX_BODY_BYTES", 10<<20)`)
	assertFileContains(t, root, "internal/config/config.go", `return 0, fmt.Errorf("parse %s as duration: %w", key, err)`)
	assertFileContains(t, root, "internal/config/config.go", "type LogConfig struct")
	assertFileContains(t, root, "cmd/api/main.go", "logging.New(cfg.Log)")
	assertFileContains(t, root, "cmd/api/main.go", "slog.SetDefault(logger)")
	assertFileContains(t, root, "cmd/api/main.go", "ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout")
	assertFileContains(t, root, "cmd/api/main.go", "MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes")
	assertFileContains(t, root, "cmd/api/main.go", "MaxBodyBytes: cfg.HTTP.MaxBodyBytes")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "http.MaxBytesHandler(handler, opts.MaxBodyBytes)")
	assertFileContains(t, root, "cmd/api/main.go", "func configureLogging() error")
	assertFileContains(t, root, "cmd/api/main.go", "PersistentPreRunE")
	assertFileContains(t, root, "internal/logging/logging.go", `case "debug":`)
	assertFileContains(t, root, "internal/logging/logging.go", "ReplaceAttr: redactAttr")
	assertFileContains(t, root, "internal/logging/logging_test.go", "TestLoggerRedactsSensitiveFields")
	assertFileNotContains(t, root, "internal/logging/logging.go", "trace_id")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "GET /healthz")
	assertPathNotExists(t, root, "api")
	assertPathNotExists(t, root, "deployments")
}

func TestProjectGeneratorGenerateAPIMinimalWithOpenTelemetry(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo")
	gen := NewProjectGenerator(generator.Default())

	_, err := gen.Generate(context.Background(), NewProjectOptions{
		ProjectName:       "demo",
		ModulePath:        "example.com/demo",
		Template:          "api-minimal",
		TargetDir:         root,
		WithOpenTelemetry: true,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	assertFileContains(t, root, "go.mod", "go.opentelemetry.io/otel v1.43.0")
	assertFileContains(t, root, "internal/config/config.go", "type ObservabilityConfig struct")
	assertFileContains(t, root, "cmd/api/main.go", "observability.SetupOpenTelemetry(ctx, cfg.Observability)")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "otelhttp.NewHandler(handler, opts.ServiceName)")
	assertFileContains(t, root, "internal/observability/otel.go", "func SetupOpenTelemetry")
	assertFileContains(t, root, "internal/observability/http_client.go", "func NewHTTPTransport")
	assertFileContains(t, root, "internal/observability/http_client.go", "otelhttp.NewTransport(base)")
	assertFileContains(t, root, "internal/logging/logging.go", "trace.SpanContextFromContext(ctx)")
	assertFileContains(t, root, "internal/logging/logging.go", `"trace_id"`)
}

func TestGeneratedProjectMatrixCompiles(t *testing.T) {
	cases := []struct {
		name              string
		template          string
		withOpenTelemetry bool
	}{
		{name: "api-clean", template: "api-clean"},
		{name: "api-clean-otel", template: "api-clean", withOpenTelemetry: true},
		{name: "api-minimal", template: "api-minimal"},
		{name: "api-minimal-otel", template: "api-minimal", withOpenTelemetry: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "demo")
			gen := NewProjectGenerator(generator.Default())

			_, err := gen.Generate(context.Background(), NewProjectOptions{
				ProjectName:       "demo",
				ModulePath:        "example.com/demo",
				Template:          tc.template,
				TargetDir:         root,
				WithOpenTelemetry: tc.withOpenTelemetry,
			})
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			runGoCommand(t, root, "test", "./...")
			runGoCommand(t, root, "build", "./cmd/api")
		})
	}
}

func TestSupportedProjectTemplates(t *testing.T) {
	got := SupportedProjectTemplates()
	want := []string{"api-clean", "api-minimal"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedProjectTemplates() = %v, want %v", got, want)
	}
}

func TestProjectGeneratorRejectsUnknownTemplate(t *testing.T) {
	gen := NewProjectGenerator(generator.Default())

	_, err := gen.Generate(context.Background(), NewProjectOptions{
		ProjectName: "demo",
		ModulePath:  "example.com/demo",
		Template:    "unknown",
		TargetDir:   t.TempDir(),
	})
	if err == nil {
		t.Fatalf("Generate() error = nil, want unknown template error")
	}
	if !strings.Contains(err.Error(), "available: api-clean, api-minimal") {
		t.Fatalf("Generate() error = %v, want available templates", err)
	}
}

func listGeneratedFiles(t *testing.T, root string) []string {
	t.Helper()

	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir() error = %v", err)
	}

	sort.Strings(files)
	return files
}

func assertFileContains(t *testing.T, root string, path string, want string) {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Fatalf("%s missing %q:\n%s", path, want, string(content))
	}
}

func assertPathNotExists(t *testing.T, root string, path string) {
	t.Helper()

	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err == nil {
		t.Fatalf("expected %s to not exist", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("Stat(%s) error = %v", path, err)
	}
}

func assertFileNotContains(t *testing.T, root string, path string, wantAbsent string) {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	if strings.Contains(string(content), wantAbsent) {
		t.Fatalf("%s unexpectedly contains %q:\n%s", path, wantAbsent, string(content))
	}
}

func runGoCommand(t *testing.T, dir string, args ...string) {
	t.Helper()

	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go command not available: %v", err)
	}

	cacheDir := filepath.Join(dir, ".gocache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", cacheDir, err)
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOCACHE="+cacheDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}
