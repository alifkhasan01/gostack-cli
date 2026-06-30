package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alifkhasan01/gostack-cli/internal/printer"
	"github.com/alifkhasan01/gostack-cli/internal/replacer"
	"github.com/alifkhasan01/gostack-cli/internal/runner"
	"github.com/alifkhasan01/gostack-cli/internal/scaffold"
	"github.com/alifkhasan01/gostack-cli/internal/template"
	"github.com/alifkhasan01/gostack-cli/internal/wizard"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new Go project from a starter template",
	Example: `  gostack create
  gostack create my-api
  gostack create my-api --framework gin --arch clean --database postgres --orm gorm --auth jwt --docker --swagger`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCreate,
}

func runCreate(_ *cobra.Command, args []string) error {
	projectName := ""
	if len(args) > 0 {
		projectName = args[0]
	}

	// --- Resolve config: flags (non-interactive) OR wizard ---
	cfg, err := resolveConfig(projectName)
	if err != nil {
		return err
	}

	// --- Resolve destination ---
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	destDir := filepath.Join(cwd, cfg.ProjectName)

	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("directory '%s' already exists", cfg.ProjectName)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}

	// --- Download or scaffold ---
	var remoteErr error
	if cfg.Type == "" || cfg.Type == "rest-api" {
		sp := printer.NewSpinner(fmt.Sprintf("Fetching template: %s-%s", cfg.Framework, cfg.Architecture))
		sp.Start()
		remoteErr = template.Download(cfg.Framework, cfg.Architecture, destDir)
		if remoteErr != nil {
			sp.Fail(fmt.Sprintf("Remote template unavailable, using built-in scaffold (%s-%s)", cfg.Framework, cfg.Architecture))
		} else {
			sp.Done(fmt.Sprintf("Template downloaded: %s-%s", cfg.Framework, cfg.Architecture))
		}
	} else {
		remoteErr = fmt.Errorf("skip remote template")
	}

	if remoteErr != nil {
		sp2 := printer.NewSpinner("Generating project structure ...")
		sp2.Start()
		if err := scaffold.Generate(destDir, scaffold.Config{
			ProjectName:    cfg.ProjectName,
			ModuleName:     cfg.ModuleName,
			Type:           cfg.Type,
			Framework:      cfg.Framework,
			Architecture:   cfg.Architecture,
			Database:       cfg.Database,
			ORM:            cfg.ORM,
			Auth:           cfg.Auth,
			Docker:         cfg.Docker,
			Swagger:        cfg.Swagger,
			Version:        Version,
			CLILib:         cfg.CLILib,
			Services:       cfg.Services,
			CSSFramework:   cfg.CSSFramework,
			TemplateEngine: cfg.TemplateEngine,
		}); err != nil {
			sp2.Fail("Scaffold failed")
			os.RemoveAll(destDir)
			return fmt.Errorf("scaffold: %w", err)
		}
		sp2.Done("Project structure generated")
	} else {
		sp2 := printer.NewSpinner("Replacing placeholders ...")
		sp2.Start()
		if err := replacer.ReplaceAll(destDir, replacer.Config{
			ModuleName:  cfg.ModuleName,
			ProjectName: cfg.ProjectName,
		}); err != nil {
			sp2.Fail("Replace failed")
			return fmt.Errorf("replace placeholders: %w", err)
		}
		sp2.Done("Placeholders replaced")

		// Fix database imports — remote templates often import all drivers
		spDb := printer.NewSpinner("Fixing database imports ...")
		spDb.Start()
		if err := fixDatabaseImports(destDir, cfg.Database); err != nil {
			spDb.Fail("Database import fix skipped: " + err.Error())
		} else {
			spDb.Done("Database imports fixed")
		}
	}

	// --- git init ---
	sp3 := printer.NewSpinner("Initializing git repository ...")
	sp3.Start()
	if err := runner.GitInit(destDir); err != nil {
		sp3.Fail("git init skipped: " + err.Error())
	} else {
		sp3.Done("Git repository initialized")
	}

	// --- go mod tidy ---
	sp4 := printer.NewSpinner("Running go mod tidy ...")
	sp4.Start()
	if err := runner.GoModTidy(destDir); err != nil {
		sp4.Fail("go mod tidy skipped — run it manually inside the project")
	} else {
		sp4.Done("Dependencies resolved")
	}

	// --- Done ---
	printer.Summary(cfg.ProjectName, destDir)
	return nil
}

// resolveConfig decides whether to use wizard or CLI flags.
func resolveConfig(projectName string) (*wizard.ProjectConfig, error) {
	f := createFlags

	// Non-interactive: --type or --framework triggers flag mode
	if f.projectType != "" || f.framework != "" {
		name := projectName
		if name == "" {
			return nil, fmt.Errorf("project name required when using flags (e.g. gostack create my-api --framework gin ...)")
		}
		moduleName := f.moduleName
		if moduleName == "" {
			moduleName = name
		}

		projectType := f.projectType
		if projectType == "" {
			projectType = "rest-api"
		}

		validTypes := map[string]bool{"rest-api": true, "cli": true, "microservice": true, "fullstack": true}
		if !validTypes[projectType] {
			return nil, fmt.Errorf("invalid type %q, valid: rest-api, cli, microservice, fullstack", projectType)
		}

		switch projectType {
		case "cli":
			validCLILibs := map[string]bool{"cobra": true, "plain": true}
			if f.cliLib != "" && !validCLILibs[f.cliLib] {
				return nil, fmt.Errorf("invalid cli-lib %q, valid: cobra, plain", f.cliLib)
			}
			return &wizard.ProjectConfig{
				ProjectName: name,
				ModuleName:  moduleName,
				Type:        projectType,
				CLILib:      f.cliLib,
				Docker:      f.docker,
			}, nil

		case "microservice":
			if f.services != "" && f.services == "" {
				return nil, fmt.Errorf("services required for microservice type")
			}
			return &wizard.ProjectConfig{
				ProjectName: name,
				ModuleName:  moduleName,
				Type:        projectType,
				Services:    f.services,
				Docker:      f.docker,
			}, nil

		case "fullstack":
			return &wizard.ProjectConfig{
				ProjectName:    name,
				ModuleName:     moduleName,
				Type:           projectType,
				Database:       ifEmpty(f.database, "none"),
				CSSFramework:   f.cssFramework,
				TemplateEngine: f.templateEngine,
				Docker:         f.docker,
			}, nil

		default: // rest-api
			if f.framework == "" || f.architecture == "" || f.database == "" || f.orm == "" {
				return nil, fmt.Errorf("--framework, --arch, --database, --orm are required for rest-api")
			}

			validFrameworks := map[string]bool{"gin": true, "fiber": true, "echo": true, "chi": true}
			validArchs := map[string]bool{"standard": true, "clean": true, "hexagonal": true, "ddd": true}
			validDBs := map[string]bool{"postgres": true, "mysql": true, "sqlite": true}
			validORMs := map[string]bool{"gorm": true, "bun": true, "sqlx": true}
			validAuths := map[string]bool{"jwt": true, "session": true, "none": true}

			if !validFrameworks[f.framework] {
				return nil, fmt.Errorf("invalid framework %q, valid: gin, fiber, echo, chi", f.framework)
			}
			if !validArchs[f.architecture] {
				return nil, fmt.Errorf("invalid architecture %q, valid: standard, clean, hexagonal, ddd", f.architecture)
			}
			if !validDBs[f.database] {
				return nil, fmt.Errorf("invalid database %q, valid: postgres, mysql, sqlite", f.database)
			}
			if !validORMs[f.orm] {
				return nil, fmt.Errorf("invalid orm %q, valid: gorm, bun, sqlx", f.orm)
			}
			if f.auth != "" && !validAuths[f.auth] {
				return nil, fmt.Errorf("invalid auth %q, valid: jwt, session, none", f.auth)
			}

			return &wizard.ProjectConfig{
				ProjectName:  name,
				ModuleName:   moduleName,
				Type:         projectType,
				Framework:    f.framework,
				Architecture: f.architecture,
				Database:     f.database,
				ORM:          f.orm,
				Auth:         f.auth,
				Docker:       f.docker,
				Swagger:      f.swagger,
			}, nil
		}
	}

	// Interactive wizard
	printer.Step("🧙", "Starting project wizard ...")
	cfg, err := wizard.Run(projectName)
	if err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}
	return cfg, nil
}

func ifEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// fixDatabaseImports removes unused database driver imports from generated
// projects. Remote templates often include all three drivers (pq, mysql,
// sqlite3) — only the selected one should remain.
func fixDatabaseImports(projectDir, database string) error {
	// Which driver import to keep
	var keepImport string
	switch database {
	case "mysql":
		keepImport = "github.com/go-sql-driver/mysql"
	case "sqlite":
		keepImport = "github.com/mattn/go-sqlite3"
	default:
		keepImport = "github.com/lib/pq"
	}

	allDrivers := []string{
		"github.com/lib/pq",
		"github.com/go-sql-driver/mysql",
		"github.com/mattn/go-sqlite3",
	}

	return filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		content := string(data)

		// Check if this file has any driver imports
		hasAny := false
		for _, drv := range allDrivers {
			if strings.Contains(content, `"`+drv+`"`) {
				hasAny = true
				break
			}
		}
		if !hasAny {
			return nil
		}

		// Remove all driver imports except the selected one
		for _, drv := range allDrivers {
			if drv != keepImport {
				content = removeLineContaining(content, `"`+drv+`"`)
			}
		}

		// Clean up blank lines in import blocks left by removals
		content = cleanImportBlock(content)

		return os.WriteFile(path, []byte(content), 0644)
	})
}

// removeLineContaining removes any line from content that contains the given
// substring. It handles leading whitespace and trailing newlines.
func removeLineContaining(content, substr string) string {
	lines := strings.Split(content, "\n")
	var kept []string
	for _, line := range lines {
		if strings.Contains(line, substr) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// cleanImportBlock removes unnecessary blank lines inside import (...) blocks
// that may remain after removing imports.
func cleanImportBlock(content string) string {
	idx := strings.Index(content, "import (")
	if idx < 0 {
		return content
	}

	before := content[:idx]
	rest := content[idx:]

	// Find closing paren
	closeIdx := strings.Index(rest, "\n)")
	if closeIdx < 0 {
		return content
	}
	importBlock := rest[:closeIdx+2] // include the "\n)"

	// Remove double-blank lines inside the block
	cleaned := strings.ReplaceAll(importBlock, "\n\n\n", "\n\n")
	cleaned = strings.ReplaceAll(cleaned, "\n\n\t\n", "\n\n")

	return before + cleaned + rest[closeIdx+2:]
}
