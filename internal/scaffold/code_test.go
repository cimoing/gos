package scaffold

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cimoing/gos/internal/generator"
)

func TestCodeGeneratorGenerateUsecase(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())

	result, err := gen.GenerateUsecase(context.Background(), MakeUsecaseOptions{
		TargetDir: root,
		Name:      "order/create",
	})
	if err != nil {
		t.Fatalf("GenerateUsecase() error = %v", err)
	}
	if len(result.Created) != 2 {
		t.Fatalf("created files = %v, want 2 files", result.Created)
	}

	for _, path := range []string{
		"internal/usecase/order/create.go",
		"internal/usecase/order/create_test.go",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}
}

func TestCodeGeneratorRejectsInvalidUsecaseName(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateUsecase(context.Background(), MakeUsecaseOptions{
		TargetDir: t.TempDir(),
		Name:      "register",
	})
	if err == nil {
		t.Fatalf("GenerateUsecase() error = nil, want invalid name error")
	}
}

func TestCodeGeneratorGenerateHandler(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())

	result, err := gen.GenerateHandler(context.Background(), MakeHandlerOptions{
		TargetDir:  root,
		Name:       "order",
		ModulePath: "example.com/demo",
	})
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}
	if len(result.Created) != 2 {
		t.Fatalf("created files = %v, want 2 files", result.Created)
	}

	for _, path := range []string{
		"internal/interfaces/http/handler/order_handler.go",
		"internal/interfaces/http/handler/order_handler_test.go",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}
}

func TestCodeGeneratorGenerateHandlerRegistersRouter(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo")
	projectGen := NewProjectGenerator(generator.Default())
	_, err := projectGen.Generate(context.Background(), NewProjectOptions{
		ProjectName: "demo",
		ModulePath:  "example.com/demo",
		Template:    "api-clean",
		TargetDir:   root,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	codeGen := NewCodeGenerator(generator.Default())
	result, err := codeGen.GenerateHandler(context.Background(), MakeHandlerOptions{
		TargetDir:  root,
		Name:       "order",
		ModulePath: "example.com/demo",
		Register:   true,
	})
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}
	if len(result.Updated) != 1 {
		t.Fatalf("updated files = %v, want router update", result.Updated)
	}

	router, err := os.ReadFile(filepath.Join(root, "internal", "interfaces", "http", "router.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	routerText := string(router)
	if !strings.Contains(routerText, `"example.com/demo/internal/interfaces/http/handler"`) {
		t.Fatalf("router missing handler import:\n%s", routerText)
	}
	if !strings.Contains(routerText, "orderHandler.RegisterRoutes(mux)") {
		t.Fatalf("router missing handler registration:\n%s", routerText)
	}
}

func TestCodeGeneratorGenerateHandlerSkipsNonStandardRouterRegistration(t *testing.T) {
	root := t.TempDir()
	routerDir := filepath.Join(root, "internal", "interfaces", "http")
	if err := os.MkdirAll(routerDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	routerPath := filepath.Join(routerDir, "router.go")
	if err := os.WriteFile(routerPath, []byte("package http\n\nfunc NewRouter() {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	gen := NewCodeGenerator(generator.Default())
	result, err := gen.GenerateHandler(context.Background(), MakeHandlerOptions{
		TargetDir:  root,
		Name:       "order",
		ModulePath: "example.com/demo",
		Register:   true,
	})
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}
	if len(result.Created) != 2 {
		t.Fatalf("created files = %v, want handler and test", result.Created)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "internal/interfaces/http/router.go" {
		t.Fatalf("skipped files = %v, want router skipped", result.Skipped)
	}

	router, err := os.ReadFile(routerPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(router), "orderHandler.RegisterRoutes") {
		t.Fatalf("router was unexpectedly modified:\n%s", string(router))
	}
}

func TestCodeGeneratorGenerateHandlerRegistersOpenAPI(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo")
	projectGen := NewProjectGenerator(generator.Default())
	_, err := projectGen.Generate(context.Background(), NewProjectOptions{
		ProjectName: "demo",
		ModulePath:  "example.com/demo",
		Template:    "api-clean",
		TargetDir:   root,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	codeGen := NewCodeGenerator(generator.Default())
	result, err := codeGen.GenerateHandler(context.Background(), MakeHandlerOptions{
		TargetDir:  root,
		Name:       "order",
		ModulePath: "example.com/demo",
		OpenAPI:    true,
	})
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}
	if len(result.Updated) != 1 {
		t.Fatalf("updated files = %v, want OpenAPI update", result.Updated)
	}

	openAPI, err := os.ReadFile(filepath.Join(root, "api", "openapi.yaml"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	openAPIText := string(openAPI)
	for _, want := range []string{
		"  /orders:",
		"tags:",
		"- Orders",
		"operationId: listOrders",
		"$ref: \"#/components/schemas/ListResponse\"",
		"$ref: \"#/components/responses/BadRequest\"",
		"$ref: \"#/components/responses/InternalServerError\"",
		"data: []",
	} {
		if !strings.Contains(openAPIText, want) {
			t.Fatalf("openapi missing %q:\n%s", want, openAPIText)
		}
	}
	if strings.Index(openAPIText, "  /orders:") > strings.Index(openAPIText, "\ncomponents:") {
		t.Fatalf("openapi path was not inserted before components:\n%s", openAPIText)
	}
}

func TestCodeGeneratorGenerateHandlerSkipsNonStandardOpenAPIRegistration(t *testing.T) {
	root := t.TempDir()
	apiDir := filepath.Join(root, "api")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	openAPIPath := filepath.Join(apiDir, "openapi.yaml")
	if err := os.WriteFile(openAPIPath, []byte("openapi: 3.0.3\npaths: {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	gen := NewCodeGenerator(generator.Default())
	result, err := gen.GenerateHandler(context.Background(), MakeHandlerOptions{
		TargetDir:  root,
		Name:       "order",
		ModulePath: "example.com/demo",
		OpenAPI:    true,
	})
	if err != nil {
		t.Fatalf("GenerateHandler() error = %v", err)
	}
	if len(result.Created) != 2 {
		t.Fatalf("created files = %v, want handler and test", result.Created)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "api/openapi.yaml" {
		t.Fatalf("skipped files = %v, want OpenAPI skipped", result.Skipped)
	}

	openAPI, err := os.ReadFile(openAPIPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(openAPI), "/orders") {
		t.Fatalf("openapi was unexpectedly modified:\n%s", string(openAPI))
	}
}

func TestCodeGeneratorGenerateCommand(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())

	result, err := gen.GenerateCommand(context.Background(), MakeCommandOptions{
		TargetDir:  root,
		Name:       "sync-orders",
		ModulePath: "example.com/demo",
	})
	if err != nil {
		t.Fatalf("GenerateCommand() error = %v", err)
	}
	if len(result.Created) != 2 {
		t.Fatalf("created files = %v, want 2 files", result.Created)
	}

	for _, path := range []string{
		"internal/command/sync_orders.go",
		"internal/command/sync_orders_test.go",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}

	assertGeneratedFileContains(t, root, "internal/command/sync_orders.go", "func NewSyncOrdersCommand() *cobra.Command")
	assertGeneratedFileContains(t, root, "internal/command/sync_orders.go", `"command", "sync-orders"`)
	assertGeneratedFileContains(t, root, "internal/command/sync_orders.go", "github.com/spf13/cobra")
}

func TestCodeGeneratorGenerateCommandRegistersMain(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo")
	projectGen := NewProjectGenerator(generator.Default())
	_, err := projectGen.Generate(context.Background(), NewProjectOptions{
		ProjectName: "demo",
		ModulePath:  "example.com/demo",
		Template:    "api-clean",
		TargetDir:   root,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	codeGen := NewCodeGenerator(generator.Default())
	result, err := codeGen.GenerateCommand(context.Background(), MakeCommandOptions{
		TargetDir:  root,
		Name:       "sync-orders",
		ModulePath: "example.com/demo",
		Register:   true,
	})
	if err != nil {
		t.Fatalf("GenerateCommand() error = %v", err)
	}
	if len(result.Updated) != 1 {
		t.Fatalf("updated files = %v, want main update", result.Updated)
	}

	main, err := os.ReadFile(filepath.Join(root, "cmd", "api", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	mainText := string(main)
	if !strings.Contains(mainText, `appcommand "example.com/demo/internal/command"`) {
		t.Fatalf("main missing command import:\n%s", mainText)
	}
	if !strings.Contains(mainText, "rootCmd.AddCommand(appcommand.NewSyncOrdersCommand())") {
		t.Fatalf("main missing command registration:\n%s", mainText)
	}
}

func TestCodeGeneratorGenerateCommandSkipsNonStandardMainRegistration(t *testing.T) {
	root := t.TempDir()
	mainDir := filepath.Join(root, "cmd", "api")
	if err := os.MkdirAll(mainDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	mainPath := filepath.Join(mainDir, "main.go")
	if err := os.WriteFile(mainPath, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	gen := NewCodeGenerator(generator.Default())
	result, err := gen.GenerateCommand(context.Background(), MakeCommandOptions{
		TargetDir:  root,
		Name:       "sync-orders",
		ModulePath: "example.com/demo",
		Register:   true,
	})
	if err != nil {
		t.Fatalf("GenerateCommand() error = %v", err)
	}
	if len(result.Created) != 2 {
		t.Fatalf("created files = %v, want command and test", result.Created)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "cmd/api/main.go" {
		t.Fatalf("skipped files = %v, want main skipped", result.Skipped)
	}
}

func TestCodeGeneratorRejectsReservedCommandName(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateCommand(context.Background(), MakeCommandOptions{
		TargetDir:  t.TempDir(),
		Name:       "queue",
		ModulePath: "example.com/demo",
	})
	if err == nil {
		t.Fatalf("GenerateCommand() error = nil, want reserved command error")
	}
}

func TestCodeGeneratorGenerateRepository(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())

	result, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  root,
		Name:       "order",
		ModulePath: "example.com/demo",
		DB:         "mysql",
	})
	if err != nil {
		t.Fatalf("GenerateRepository() error = %v", err)
	}
	if len(result.Created) != 5 {
		t.Fatalf("created files = %v, want 5 files", result.Created)
	}

	for _, path := range []string{
		"internal/domain/order/entity.go",
		"internal/domain/order/repository.go",
		"internal/infrastructure/persistence/mysql/order_repository.go",
		"internal/infrastructure/persistence/mysql/order_repository_integration_test.go",
		"internal/infrastructure/persistence/mysql/order_repository_test.go",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}

	content, err := os.ReadFile(filepath.Join(root, "internal", "infrastructure", "persistence", "mysql", "order_repository.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "SELECT id FROM orders WHERE id = ?") {
		t.Fatalf("repository does not use default table name orders:\n%s", string(content))
	}
	if !strings.Contains(string(content), "DeleteByID") {
		t.Fatalf("repository does not include DeleteByID:\n%s", string(content))
	}
}

func TestCodeGeneratorGenerateRepositoryRegistersAssembly(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo")
	projectGen := NewProjectGenerator(generator.Default())
	_, err := projectGen.Generate(context.Background(), NewProjectOptions{
		ProjectName: "demo",
		ModulePath:  "example.com/demo",
		Template:    "api-clean",
		TargetDir:   root,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	codeGen := NewCodeGenerator(generator.Default())
	result, err := codeGen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  root,
		Name:       "order",
		ModulePath: "example.com/demo",
		DB:         "mysql",
		Register:   true,
	})
	if err != nil {
		t.Fatalf("GenerateRepository() error = %v", err)
	}
	if len(result.Updated) != 1 {
		t.Fatalf("updated files = %v, want assembly update", result.Updated)
	}

	assembly, err := os.ReadFile(filepath.Join(root, "internal", "app", "assembly.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	assemblyText := string(assembly)
	for _, want := range []string{
		`mysqlrepo "example.com/demo/internal/infrastructure/persistence/mysql"`,
		"OrderRepository *mysqlrepo.OrderRepository",
		"orderRepository := mysqlrepo.NewOrderRepository(db)",
		"OrderRepository: orderRepository",
	} {
		if !strings.Contains(assemblyText, want) {
			t.Fatalf("assembly missing %q:\n%s", want, assemblyText)
		}
	}
}

func TestCodeGeneratorGenerateRepositorySkipsNonStandardAssemblyRegistration(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "internal", "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	assemblyPath := filepath.Join(appDir, "assembly.go")
	if err := os.WriteFile(assemblyPath, []byte("package app\n\ntype Dependencies struct{}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	gen := NewCodeGenerator(generator.Default())
	result, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  root,
		Name:       "order",
		ModulePath: "example.com/demo",
		DB:         "mysql",
		Register:   true,
	})
	if err != nil {
		t.Fatalf("GenerateRepository() error = %v", err)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "internal/app/assembly.go" {
		t.Fatalf("skipped files = %v, want assembly skipped", result.Skipped)
	}
}

func TestCodeGeneratorGenerateModel(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())

	result, err := gen.GenerateModel(context.Background(), MakeModelOptions{
		TargetDir: root,
		Name:      "invoice",
		Fields:    "number:string:json=invoice_number,size=64,total:int64:default=100,created_at:time",
	})
	if err != nil {
		t.Fatalf("GenerateModel() error = %v", err)
	}
	if len(result.Created) != 1 {
		t.Fatalf("created files = %v, want entity file", result.Created)
	}

	entity, err := os.ReadFile(filepath.Join(root, "internal", "domain", "invoice", "entity.go"))
	if err != nil {
		t.Fatalf("ReadFile(entity) error = %v", err)
	}
	entityText := string(entity)
	for _, want := range []string{
		"package invoice",
		`import "time"`,
		"ID        int64     `json:\"id\"`",
		"Number    string    `json:\"invoice_number\"`",
		"Total     int64     `json:\"total\"`",
		"CreatedAt time.Time `json:\"created_at\"`",
	} {
		if !strings.Contains(entityText, want) {
			t.Fatalf("entity missing %q:\n%s", want, entityText)
		}
	}
}

func TestCodeGeneratorGenerateModelRejectsInvalidField(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateModel(context.Background(), MakeModelOptions{
		TargetDir: t.TempDir(),
		Name:      "invoice",
		Fields:    "id:string",
	})
	if err == nil {
		t.Fatalf("GenerateModel() error = nil, want invalid field error")
	}
}

func TestCodeGeneratorGenerateModelRejectsExistingFile(t *testing.T) {
	root := t.TempDir()
	entityDir := filepath.Join(root, "internal", "domain", "invoice")
	if err := os.MkdirAll(entityDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(entityDir, "entity.go"), []byte("package invoice\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	gen := NewCodeGenerator(generator.Default())
	_, err := gen.GenerateModel(context.Background(), MakeModelOptions{
		TargetDir: root,
		Name:      "invoice",
	})
	if err == nil {
		t.Fatalf("GenerateModel() error = nil, want conflict error")
	}
}

func TestCodeGeneratorRejectsUnsupportedRepositoryDB(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  t.TempDir(),
		Name:       "order",
		ModulePath: "example.com/demo",
		DB:         "postgres",
	})
	if err == nil {
		t.Fatalf("GenerateRepository() error = nil, want unsupported db error")
	}
}

func TestCodeGeneratorGenerateRepositoryWithCustomTable(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  root,
		Name:       "order",
		ModulePath: "example.com/demo",
		DB:         "mysql",
		TableName:  "sales_orders",
	})
	if err != nil {
		t.Fatalf("GenerateRepository() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(root, "internal", "infrastructure", "persistence", "mysql", "order_repository.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "SELECT id FROM sales_orders WHERE id = ?") {
		t.Fatalf("repository does not use custom table name:\n%s", string(content))
	}
}

func TestCodeGeneratorGenerateRepositoryWithMigration(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())
	gen.now = func() time.Time {
		return time.Date(2026, 6, 3, 13, 45, 0, 0, time.UTC)
	}

	result, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:     root,
		Name:          "invoice",
		ModulePath:    "example.com/demo",
		DB:            "mysql",
		TableName:     "sales_invoices",
		WithMigration: true,
	})
	if err != nil {
		t.Fatalf("GenerateRepository() error = %v", err)
	}
	if len(result.Created) != 7 {
		t.Fatalf("created files = %v, want 7 files", result.Created)
	}

	upPath := filepath.Join(root, "migrations", "20260603134500_create_sales_invoices_table.up.sql")
	downPath := filepath.Join(root, "migrations", "20260603134500_create_sales_invoices_table.down.sql")
	up, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("ReadFile(up) error = %v", err)
	}
	down, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("ReadFile(down) error = %v", err)
	}
	if !strings.Contains(string(up), "CREATE TABLE sales_invoices") {
		t.Fatalf("up migration = %q", string(up))
	}
	if !strings.Contains(string(down), "DROP TABLE IF EXISTS sales_invoices") {
		t.Fatalf("down migration = %q", string(down))
	}
}

func TestCodeGeneratorGenerateRepositoryWithFields(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())
	gen.now = func() time.Time {
		return time.Date(2026, 6, 3, 14, 10, 0, 0, time.UTC)
	}

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:     root,
		Name:          "invoice",
		ModulePath:    "example.com/demo",
		DB:            "mysql",
		TableName:     "sales_invoices",
		Fields:        "number:string,total:int64,paid:bool,created_at:time",
		WithMigration: true,
	})
	if err != nil {
		t.Fatalf("GenerateRepository() error = %v", err)
	}

	entity, err := os.ReadFile(filepath.Join(root, "internal", "domain", "invoice", "entity.go"))
	if err != nil {
		t.Fatalf("ReadFile(entity) error = %v", err)
	}
	entityText := string(entity)
	for _, want := range []string{
		`import "time"`,
		"ID        int64     `json:\"id\"`",
		"Number    string    `json:\"number\"`",
		"Total     int64     `json:\"total\"`",
		"Paid      bool      `json:\"paid\"`",
		"CreatedAt time.Time `json:\"created_at\"`",
	} {
		if !strings.Contains(entityText, want) {
			t.Fatalf("entity missing %q:\n%s", want, entityText)
		}
	}

	repository, err := os.ReadFile(filepath.Join(root, "internal", "infrastructure", "persistence", "mysql", "invoice_repository.go"))
	if err != nil {
		t.Fatalf("ReadFile(repository) error = %v", err)
	}
	repositoryText := string(repository)
	for _, want := range []string{
		`"example.com/demo/internal/infrastructure/database"`,
		"SELECT id, number, total, paid, created_at FROM sales_invoices WHERE id = ?",
		"executor := database.ExecutorFromContext(ctx, r.db)",
		"INSERT INTO sales_invoices (number, total, paid, created_at) VALUES (?, ?, ?, ?)",
		"UPDATE sales_invoices SET number = ?, total = ?, paid = ?, created_at = ? WHERE id = ?",
		"&invoice.CreatedAt",
	} {
		if !strings.Contains(repositoryText, want) {
			t.Fatalf("repository missing %q:\n%s", want, repositoryText)
		}
	}

	migration, err := os.ReadFile(filepath.Join(root, "migrations", "20260603141000_create_sales_invoices_table.up.sql"))
	if err != nil {
		t.Fatalf("ReadFile(migration) error = %v", err)
	}
	migrationText := string(migration)
	for _, want := range []string{
		"number VARCHAR(255) NOT NULL DEFAULT ''",
		"total BIGINT NOT NULL DEFAULT 0",
		"paid BOOLEAN NOT NULL DEFAULT FALSE",
		"created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP",
	} {
		if !strings.Contains(migrationText, want) {
			t.Fatalf("migration missing %q:\n%s", want, migrationText)
		}
	}
}

func TestCodeGeneratorGenerateRepositoryWithFieldOptions(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())
	gen.now = func() time.Time {
		return time.Date(2026, 6, 3, 16, 20, 0, 0, time.UTC)
	}

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:     root,
		Name:          "customer",
		ModulePath:    "example.com/demo",
		DB:            "mysql",
		Fields:        "email:string:unique,size=320,deleted_at:time:nullable,index,age:int:required",
		WithMigration: true,
	})
	if err != nil {
		t.Fatalf("GenerateRepository() error = %v", err)
	}

	migration, err := os.ReadFile(filepath.Join(root, "migrations", "20260603162000_create_customers_table.up.sql"))
	if err != nil {
		t.Fatalf("ReadFile(migration) error = %v", err)
	}
	migrationText := string(migration)
	for _, want := range []string{
		"email VARCHAR(320) NOT NULL DEFAULT ''",
		"deleted_at TIMESTAMP NULL",
		"age INT NOT NULL DEFAULT 0",
		"UNIQUE KEY uk_email (email)",
		"KEY idx_deleted_at (deleted_at)",
	} {
		if !strings.Contains(migrationText, want) {
			t.Fatalf("migration missing %q:\n%s", want, migrationText)
		}
	}
}

func TestCodeGeneratorGenerateRepositoryWithAdvancedFieldOptions(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())
	gen.now = func() time.Time {
		return time.Date(2026, 6, 3, 17, 5, 0, 0, time.UTC)
	}

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:     root,
		Name:          "profile",
		ModulePath:    "example.com/demo",
		DB:            "mysql",
		Fields:        "email:string:json=email_address,size=320,score:int64:default=100,status:string:default=active,meta:string:sql=TEXT,ratio:int:sql=DECIMAL(10,2),enabled:bool:default=true,seen_at:time:default=now",
		WithMigration: true,
	})
	if err != nil {
		t.Fatalf("GenerateRepository() error = %v", err)
	}

	entity, err := os.ReadFile(filepath.Join(root, "internal", "domain", "profile", "entity.go"))
	if err != nil {
		t.Fatalf("ReadFile(entity) error = %v", err)
	}
	entityText := string(entity)
	if !strings.Contains(entityText, "Email   string    `json:\"email_address\"`") {
		t.Fatalf("entity missing custom json tag:\n%s", entityText)
	}

	migration, err := os.ReadFile(filepath.Join(root, "migrations", "20260603170500_create_profiles_table.up.sql"))
	if err != nil {
		t.Fatalf("ReadFile(migration) error = %v", err)
	}
	migrationText := string(migration)
	for _, want := range []string{
		"email VARCHAR(320) NOT NULL DEFAULT ''",
		"score BIGINT NOT NULL DEFAULT 100",
		"status VARCHAR(255) NOT NULL DEFAULT 'active'",
		"meta TEXT NOT NULL DEFAULT ''",
		"ratio DECIMAL(10,2) NOT NULL DEFAULT 0",
		"enabled BOOLEAN NOT NULL DEFAULT TRUE",
		"seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP",
	} {
		if !strings.Contains(migrationText, want) {
			t.Fatalf("migration missing %q:\n%s", want, migrationText)
		}
	}
}

func TestCodeGeneratorGenerateUsecaseTest(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())

	result, err := gen.GenerateTest(context.Background(), MakeTestOptions{
		TargetDir:  root,
		Kind:       "usecase",
		Name:       "order/create",
		ModulePath: "example.com/demo",
	})
	if err != nil {
		t.Fatalf("GenerateTest() error = %v", err)
	}
	if len(result.Created) != 1 {
		t.Fatalf("created files = %v, want one file", result.Created)
	}
	if _, err := os.Stat(filepath.Join(root, "internal", "usecase", "order", "create_test.go")); err != nil {
		t.Fatalf("expected usecase test file: %v", err)
	}
}

func TestCodeGeneratorGenerateHandlerTest(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())

	result, err := gen.GenerateTest(context.Background(), MakeTestOptions{
		TargetDir:  root,
		Kind:       "handler",
		Name:       "order",
		ModulePath: "example.com/demo",
	})
	if err != nil {
		t.Fatalf("GenerateTest() error = %v", err)
	}
	if len(result.Created) != 1 {
		t.Fatalf("created files = %v, want one file", result.Created)
	}
	if _, err := os.Stat(filepath.Join(root, "internal", "interfaces", "http", "handler", "order_handler_test.go")); err != nil {
		t.Fatalf("expected handler test file: %v", err)
	}
}

func TestCodeGeneratorGenerateRepositoryTest(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())

	result, err := gen.GenerateTest(context.Background(), MakeTestOptions{
		TargetDir:  root,
		Kind:       "repository",
		Name:       "order",
		ModulePath: "example.com/demo",
	})
	if err != nil {
		t.Fatalf("GenerateTest() error = %v", err)
	}
	if len(result.Created) != 1 {
		t.Fatalf("created files = %v, want one file", result.Created)
	}
	if _, err := os.Stat(filepath.Join(root, "internal", "infrastructure", "persistence", "mysql", "order_repository_test.go")); err != nil {
		t.Fatalf("expected repository test file: %v", err)
	}
}

func TestCodeGeneratorRejectsUnsupportedTestKind(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateTest(context.Background(), MakeTestOptions{
		TargetDir:  t.TempDir(),
		Kind:       "service",
		Name:       "order",
		ModulePath: "example.com/demo",
	})
	if err == nil {
		t.Fatalf("GenerateTest() error = nil, want unsupported kind error")
	}
}

func TestCodeGeneratorRejectsInvalidRepositoryTable(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  t.TempDir(),
		Name:       "order",
		ModulePath: "example.com/demo",
		DB:         "mysql",
		TableName:  "orders;DROP",
	})
	if err == nil {
		t.Fatalf("GenerateRepository() error = nil, want invalid table error")
	}
}

func TestCodeGeneratorRejectsInvalidRepositoryField(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  t.TempDir(),
		Name:       "invoice",
		ModulePath: "example.com/demo",
		DB:         "mysql",
		Fields:     "id:string",
	})
	if err == nil {
		t.Fatalf("GenerateRepository() error = nil, want invalid field error")
	}
}

func TestCodeGeneratorRejectsInvalidRepositoryFieldOption(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  t.TempDir(),
		Name:       "invoice",
		ModulePath: "example.com/demo",
		DB:         "mysql",
		Fields:     "amount:int:size=10",
	})
	if err == nil {
		t.Fatalf("GenerateRepository() error = nil, want invalid field option error")
	}
}

func TestCodeGeneratorRejectsInvalidRepositoryFieldDefault(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  t.TempDir(),
		Name:       "invoice",
		ModulePath: "example.com/demo",
		DB:         "mysql",
		Fields:     "amount:int:default=abc",
	})
	if err == nil {
		t.Fatalf("GenerateRepository() error = nil, want invalid default error")
	}
}

func TestCodeGeneratorRejectsNullDefaultForRequiredRepositoryField(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  t.TempDir(),
		Name:       "invoice",
		ModulePath: "example.com/demo",
		DB:         "mysql",
		Fields:     "status:string:default=null",
	})
	if err == nil {
		t.Fatalf("GenerateRepository() error = nil, want default null error")
	}
}

func TestCodeGeneratorRejectsInvalidRepositoryFieldSQLType(t *testing.T) {
	gen := NewCodeGenerator(generator.Default())

	_, err := gen.GenerateRepository(context.Background(), MakeRepositoryOptions{
		TargetDir:  t.TempDir(),
		Name:       "invoice",
		ModulePath: "example.com/demo",
		DB:         "mysql",
		Fields:     "meta:string:sql=TEXT;DROP",
	})
	if err == nil {
		t.Fatalf("GenerateRepository() error = nil, want invalid sql option error")
	}
}

func TestCodeGeneratorGenerateMigration(t *testing.T) {
	root := t.TempDir()
	gen := NewCodeGenerator(generator.Default())
	gen.now = func() time.Time {
		return time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	}

	result, err := gen.GenerateMigration(context.Background(), MakeMigrationOptions{
		TargetDir: root,
		Name:      "create_users_table",
	})
	if err != nil {
		t.Fatalf("GenerateMigration() error = %v", err)
	}
	if len(result.Created) != 2 {
		t.Fatalf("created files = %v, want 2 files", result.Created)
	}

	for _, path := range []string{
		"migrations/20260603120000_create_users_table.up.sql",
		"migrations/20260603120000_create_users_table.down.sql",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}
}

func assertGeneratedFileContains(t *testing.T, root string, path string, want string) {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Fatalf("%s missing %q:\n%s", path, want, string(content))
	}
}
