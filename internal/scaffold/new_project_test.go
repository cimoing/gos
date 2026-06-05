package scaffold

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/jake/gola/internal/generator"
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
		"internal/pkg/apperror/error.go",
		"internal/pkg/response/response.go",
		"internal/usecase/user/register.go",
		"internal/usecase/user/register_test.go",
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
	assertFileContains(t, root, "internal/interfaces/http/router.go", "middleware.CORS(middleware.CORSOptions{})")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "middleware.Timeout(10*time.Second)")
	assertFileContains(t, root, "cmd/api/main.go", "github.com/spf13/cobra")
	assertFileContains(t, root, "cmd/api/main.go", "func newRootCommand(ctx context.Context) *cobra.Command")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newServeCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newScheduleCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newQueueCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "gos:command-imports")
	assertFileContains(t, root, "cmd/api/main.go", "gos:commands")
	assertFileContains(t, root, "deployments/docker/docker-compose.yml", "mysql:")
	assertFileContains(t, root, "deployments/docker/docker-compose.yml", "redis:")
	assertFileContains(t, root, "deployments/docker/docker-compose.test.yml", "mysql_test:")
	assertFileContains(t, root, "deployments/docker/docker-compose.test.yml", "3307:3306")
	assertFileContains(t, root, ".env.example", "DB_DSN=")
	assertFileContains(t, root, ".env.example", "DB_ENABLE_NESTED_TRANSACTION=false")
	assertFileContains(t, root, ".env.example", "REDIS_ADDR=")
	assertFileContains(t, root, "internal/config/config.go", "EnableNestedTransaction bool")
	assertFileContains(t, root, "internal/config/config.go", `getEnvBool("DB_ENABLE_NESTED_TRANSACTION", false)`)
	assertFileContains(t, root, "go.mod", "github.com/go-sql-driver/mysql v1.10.0")
	assertFileContains(t, root, "go.mod", "github.com/spf13/cobra v1.10.2")
	assertFileContains(t, root, "go.mod", "filippo.io/edwards25519 v1.2.0 // indirect")
	assertFileContains(t, root, "go.sum", "github.com/go-sql-driver/mysql v1.10.0 h1:")
	assertFileContains(t, root, "go.sum", "github.com/spf13/cobra v1.10.2 h1:")
	assertFileContains(t, root, "go.sum", "github.com/spf13/pflag v1.0.9 h1:")
	assertFileContains(t, root, "go.sum", "filippo.io/edwards25519 v1.2.0 h1:")
	assertFileNotContains(t, root, "go.mod", "go.opentelemetry.io/otel")
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
	assertFileContains(t, root, "internal/interfaces/http/middleware/middleware_test.go", "TestCORSHandlesPreflight")
	assertFileContains(t, root, "api/openapi.yaml", "openapi: 3.0.3")
	assertFileContains(t, root, "api/openapi.yaml", "/healthz:")
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
	assertFileContains(t, root, ".env.example", "OTEL_ENABLED=false")
	assertFileContains(t, root, "internal/config/config.go", "type ObservabilityConfig struct")
	assertFileContains(t, root, "internal/app/app.go", "observability.SetupOpenTelemetry(ctx, cfg.Observability)")
	assertFileContains(t, root, "internal/interfaces/http/router.go", "otelhttp.NewHandler(handler, opts.ServiceName)")
	assertFileContains(t, root, "internal/observability/otel.go", "func SetupOpenTelemetry")
	assertFileContains(t, root, "internal/observability/otel.go", "otlptracehttp.New")
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
	}

	gotFiles := listGeneratedFiles(t, root)
	if !reflect.DeepEqual(gotFiles, wantFiles) {
		t.Fatalf("generated files mismatch\nwant:\n%s\n\ngot:\n%s", strings.Join(wantFiles, "\n"), strings.Join(gotFiles, "\n"))
	}

	assertFileContains(t, root, "README.md", "api-minimal")
	assertFileContains(t, root, "cmd/api/main.go", "httpinterface.NewRouter()")
	assertFileContains(t, root, "cmd/api/main.go", "github.com/spf13/cobra")
	assertFileContains(t, root, "cmd/api/main.go", "func newRootCommand(ctx context.Context) *cobra.Command")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newServeCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newScheduleCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "rootCmd.AddCommand(newQueueCommand())")
	assertFileContains(t, root, "cmd/api/main.go", "gos:command-imports")
	assertFileContains(t, root, "cmd/api/main.go", "gos:commands")
	assertFileContains(t, root, "go.mod", "github.com/spf13/cobra v1.10.2")
	assertFileContains(t, root, "go.sum", "github.com/spf13/cobra v1.10.2 h1:")
	assertFileNotContains(t, root, "go.mod", "go.opentelemetry.io/otel")
	assertPathNotExists(t, root, "internal/observability")
	assertFileContains(t, root, "internal/config/config.go", `getEnv("APP_NAME", "demo")`)
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
	assertFileContains(t, root, "internal/interfaces/http/router.go", "otelhttp.NewHandler(mux, opts.ServiceName)")
	assertFileContains(t, root, "internal/observability/otel.go", "func SetupOpenTelemetry")
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
