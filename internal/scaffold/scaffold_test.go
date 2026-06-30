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
		"cmd/server/main.go",
		"internal/config/config.go",
		"internal/model/user.go",
		"internal/handler/user.go",
		"internal/middleware/logger.go",
		"internal/middleware/auth.go",
		"internal/service/user.go",
		"internal/repository/user.go",
		"pkg/jwt/jwt.go",
		"pkg/response/response.go",
		"pkg/validator/validator.go",
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

	mustExist(t, dir, "go.mod", "cmd/server/main.go")
	mustNotExist(t, dir, "Dockerfile", "docker-compose.yml", "pkg/jwt/jwt.go")
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
			mustExist(t, dir, "internal/middleware/logger.go", "internal/middleware/auth.go")
		})
	}
}

func TestGenerate_CLI_Cobra(t *testing.T) {
	dir := t.TempDir()

	cfg := scaffold.Config{
		ProjectName: "mycli",
		ModuleName:  "github.com/test/mycli",
		Type:        "cli",
		CLILib:      "cobra",
		Docker:      true,
	}

	if err := scaffold.Generate(dir, cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mustExist(t, dir,
		"go.mod",
		"main.go",
		"cmd/root.go",
		"cmd/init.go",
		"cmd/build.go",
		"cmd/run.go",
		"internal/config/config.go",
		"internal/runner/runner.go",
		"internal/output/printer.go",
		"pkg/util/file.go",
		"Makefile",
		"README.md",
		".goreleaser.yml",
		"Dockerfile",
	)
}

func TestGenerate_CLI_Plain(t *testing.T) {
	dir := t.TempDir()

	cfg := scaffold.Config{
		ProjectName: "mycli",
		ModuleName:  "github.com/test/mycli",
		Type:        "cli",
		CLILib:      "plain",
		Docker:      false,
	}

	if err := scaffold.Generate(dir, cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mustExist(t, dir,
		"go.mod",
		"main.go",
		"cmd/root.go",
		"internal/config/config.go",
	)
	mustNotExist(t, dir, "Dockerfile", "cmd/init.go")
}

func TestGenerate_Microservice(t *testing.T) {
	dir := t.TempDir()

	cfg := scaffold.Config{
		ProjectName: "mysvc",
		ModuleName:  "github.com/test/mysvc",
		Type:        "microservice",
		Services:    "api,worker",
	}

	if err := scaffold.Generate(dir, cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mustExist(t, dir,
		"go.mod",
		"Makefile",
		"shared/go.mod",
		"shared/logger/logger.go",
		"shared/middleware/auth.go",
		"services/api-service/cmd/main.go",
		"services/api-service/internal/handler/api.go",
		"services/api-service/internal/service/api.go",
		"services/api-service/internal/repository/api.go",
		"services/api-service/internal/model/api.go",
		"services/api-service/Dockerfile",
		"services/api-service/go.mod",
		"services/worker-service/cmd/main.go",
		"services/worker-service/internal/handler/worker.go",
		"gateway/cmd/main.go",
		"gateway/internal/proxy/proxy.go",
		"infra/docker-compose.yml",
		"infra/k8s/api-deployment.yaml",
		"infra/k8s/worker-deployment.yaml",
		"infra/nginx/nginx.conf",
	)
}

func TestGenerate_Fullstack(t *testing.T) {
	dir := t.TempDir()

	cfg := scaffold.Config{
		ProjectName:    "myapp",
		ModuleName:     "github.com/test/myapp",
		Type:           "fullstack",
		Database:       "postgres",
		CSSFramework:   "tailwind",
		TemplateEngine: "html",
		Docker:         true,
	}

	if err := scaffold.Generate(dir, cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mustExist(t, dir,
		"go.mod",
		"cmd/web/main.go",
		"internal/handler/home.go",
		"internal/handler/auth.go",
		"internal/template/renderer.go",
		"internal/model/user.go",
		"internal/repository/user.go",
		"internal/service/user.go",
		"web/static/css/style.css",
		"web/static/js/app.js",
		"web/templates/layout/base.html",
		"web/templates/pages/home.html",
		"web/templates/components/navbar.html",
		".env",
		".env.example",
		"Makefile",
		"README.md",
		"Dockerfile",
		".gitignore",
	)

	checkContains(t, filepath.Join(dir, "README.md"), "postgres")
}

func TestGenerate_Fullstack_NoDB(t *testing.T) {
	dir := t.TempDir()

	cfg := scaffold.Config{
		ProjectName: "simpleweb",
		ModuleName:  "github.com/test/simpleweb",
		Type:        "fullstack",
		Database:    "none",
		Docker:      false,
	}

	if err := scaffold.Generate(dir, cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mustExist(t, dir,
		"go.mod",
		"cmd/web/main.go",
		"internal/handler/home.go",
		"web/static/css/style.css",
		"web/templates/layout/base.html",
	)
	mustNotExist(t, dir, "Dockerfile", "internal/model/user.go")
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
