package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alifkhasan01/gostack-cli/internal/generator"
)

// setupFakeProject creates a temp dir with a go.mod so findProjectRoot works.
func setupFakeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/test/app\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// chdir into the fake project
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestGenerate_Handler(t *testing.T) {
	dir := setupFakeProject(t)
	if err := generator.Generate(generator.KindHandler, "User"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, filepath.Join(dir, "internal", "handler", "user_handler.go"), "UserHandler")
}

func TestGenerate_Service(t *testing.T) {
	dir := setupFakeProject(t)
	if err := generator.Generate(generator.KindService, "Order"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, filepath.Join(dir, "internal", "service", "order_service.go"), "OrderService")
}

func TestGenerate_Repository(t *testing.T) {
	dir := setupFakeProject(t)
	if err := generator.Generate(generator.KindRepository, "Product"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, filepath.Join(dir, "internal", "repository", "product_repository.go"), "ProductRepository")
}

func TestGenerate_Migration(t *testing.T) {
	dir := setupFakeProject(t)
	if err := generator.Generate(generator.KindMigration, "create_users"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "migrations"))
	if len(entries) == 0 {
		t.Fatal("expected migration file to be created")
	}
}

func TestGenerate_Module(t *testing.T) {
	dir := setupFakeProject(t)
	if err := generator.Generate(generator.KindModule, "Invoice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, filepath.Join(dir, "internal", "handler", "invoice_handler.go"), "InvoiceHandler")
	assertFile(t, filepath.Join(dir, "internal", "service", "invoice_service.go"), "InvoiceService")
	assertFile(t, filepath.Join(dir, "internal", "repository", "invoice_repository.go"), "InvoiceRepository")
}

func TestGenerate_DuplicateReturnsError(t *testing.T) {
	setupFakeProject(t)
	if err := generator.Generate(generator.KindHandler, "User"); err != nil {
		t.Fatalf("first generate failed: %v", err)
	}
	if err := generator.Generate(generator.KindHandler, "User"); err == nil {
		t.Fatal("expected error on duplicate generate, got nil")
	}
}

// -------------------------------------------------------------------------
// helpers
// -------------------------------------------------------------------------

func assertFile(t *testing.T, path, mustContain string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not found: %s", path)
	}
	content := string(data)
	for i := 0; i <= len(content)-len(mustContain); i++ {
		if content[i:i+len(mustContain)] == mustContain {
			return
		}
	}
	t.Errorf("file %s does not contain %q", path, mustContain)
}
