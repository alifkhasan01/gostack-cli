package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alifkhasan01/gostack-cli/internal/scaffold"
)

func TestGenerate_Gin_Clean(t *testing.T) {
	dir := t.TempDir()

	cfg := scaffold.Config{
		ProjectName:  "my-api",
		ModuleName:   "github.com/test/my-api",
		Framework:    "gin",
		Architecture: "clean",
		Database:     "postgres",
		ORM:          "gorm",
		Auth:         "jwt",
		Docker:       true,
		Swagger:      true,
	}

	if err := scaffold.Generate(dir, cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mustExist(t, dir,
		"go.mod",
		"cmd/api/main.go",
		"internal/config/config.go",
		"internal/database/database.go",
		"internal/routes/routes.go",
		"internal/middleware/cors.go",
		"internal/middleware/jwt.go",
		"internal/entity/user.go",
		".env",
		".env.example",
		".gitignore",
		"Makefile",
		"README.md",
		"Dockerfile",
		"docker-compose.yml",
		"docs/.gitkeep",
	)
}

func TestGenerate_Fiber_NoDocker_NoSwagger(t *testing.T) {
	dir := t.TempDir()

	cfg := scaffold.Config{
		ProjectName:  "fiber-app",
		ModuleName:   "github.com/test/fiber-app",
		Framework:    "fiber",
		Architecture: "standard",
		Database:     "mysql",
		ORM:          "bun",
		Auth:         "none",
		Docker:       false,
		Swagger:      false,
	}

	if err := scaffold.Generate(dir, cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mustExist(t, dir, "go.mod", "cmd/api/main.go", "internal/routes/routes.go")
	mustNotExist(t, dir, "Dockerfile", "docker-compose.yml", "docs/.gitkeep", "internal/middleware/jwt.go")
}

func TestGenerate_PlaceholderReplaced(t *testing.T) {
	dir := t.TempDir()

	cfg := scaffold.Config{
		ProjectName: "cool-project",
		ModuleName:  "github.com/me/cool-project",
		Framework:   "gin",
		Database:    "sqlite",
		ORM:         "sqlx",
		Auth:        "none",
	}

	if err := scaffold.Generate(dir, cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	checkContains(t, filepath.Join(dir, "go.mod"), "github.com/me/cool-project")
	checkContains(t, filepath.Join(dir, ".env"), "cool-project")
	checkContains(t, filepath.Join(dir, "README.md"), "cool-project")
}

func TestGenerate_AllFrameworks(t *testing.T) {
	frameworks := []string{"gin", "fiber", "echo", "chi"}
	for _, fw := range frameworks {
		t.Run(fw, func(t *testing.T) {
			dir := t.TempDir()
			cfg := scaffold.Config{
				ProjectName: "test-" + fw,
				ModuleName:  "github.com/test/" + fw,
				Framework:   fw,
				Database:    "postgres",
				ORM:         "gorm",
				Auth:        "none",
			}
			if err := scaffold.Generate(dir, cfg); err != nil {
				t.Fatalf("Generate(%s) failed: %v", fw, err)
			}
			mustExist(t, dir, "internal/routes/routes.go", "internal/middleware/cors.go")
		})
	}
}

// -------------------------------------------------------------------------
// helpers
// -------------------------------------------------------------------------

func mustExist(t *testing.T, base string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		full := filepath.Join(base, p)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected file to exist: %s", p)
		}
	}
}

func mustNotExist(t *testing.T, base string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		full := filepath.Join(base, p)
		if _, err := os.Stat(full); err == nil {
			t.Errorf("expected file to NOT exist: %s", p)
		}
	}
}

func checkContains(t *testing.T, path, substr string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !contains(string(data), substr) {
		t.Errorf("file %s does not contain %q", path, substr)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
