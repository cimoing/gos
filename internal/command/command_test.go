package command

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Execute(context.Background(), []string{"help"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "gos new <project>") {
		t.Fatalf("help output = %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "gos make:model") {
		t.Fatalf("help output missing make:model = %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "--openapi") {
		t.Fatalf("help output missing --openapi = %q", stdout.String())
	}
}

func TestExecuteNewDryRunAllowsFlagsAfterProject(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Execute(context.Background(), []string{"new", target, "--module=example.com/demo", "--dry-run"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "dry run: demo") {
		t.Fatalf("new output = %q", stdout.String())
	}
}

func TestExecuteMakeUsecaseDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Execute(context.Background(), []string{"make:usecase", "order/create", "--dry-run"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "internal/usecase/order/create.go") {
		t.Fatalf("make:usecase output = %q", stdout.String())
	}
}

func TestExecuteMakeHandlerDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Execute(context.Background(), []string{"make:handler", "order", "--dry-run"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "internal/interfaces/http/handler/order_handler.go") {
		t.Fatalf("make:handler output = %q", stdout.String())
	}
}

func TestExecuteMakeRepositoryDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Execute(context.Background(), []string{"make:repository", "order", "--dry-run"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "internal/infrastructure/persistence/mysql/order_repository.go") {
		t.Fatalf("make:repository output = %q", stdout.String())
	}
}

func TestExecuteMakeModelDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Execute(context.Background(), []string{"make:model", "invoice", "--fields=number:string,total:int64", "--dry-run"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "internal/domain/invoice/entity.go") {
		t.Fatalf("make:model output = %q", stdout.String())
	}
}

func TestExecuteMakeRepositoryWithMigrationDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Execute(context.Background(), []string{"make:repository", "invoice", "--with-migration", "--table=sales_invoices", "--dry-run"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "create_sales_invoices_table.up.sql") {
		t.Fatalf("make:repository output = %q", stdout.String())
	}
}

func TestExecuteMakeRepositoryWithFieldsDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Execute(context.Background(), []string{"make:repository", "invoice", "--fields=number:string,total:int64", "--with-migration", "--dry-run"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "create_invoices_table.up.sql") {
		t.Fatalf("make:repository output = %q", stdout.String())
	}
}

func TestExecuteMakeTestDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Execute(context.Background(), []string{"make:test", "handler", "invoice", "--dry-run"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "internal/interfaces/http/handler/invoice_handler_test.go") {
		t.Fatalf("make:test output = %q", stdout.String())
	}
}

func TestExecuteMakeMigrationDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Execute(context.Background(), []string{"make:migration", "create_users_table", "--dry-run"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "migrations/") {
		t.Fatalf("make:migration output = %q", stdout.String())
	}
}
