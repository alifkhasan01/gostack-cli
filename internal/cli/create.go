package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gostack/cli/internal/printer"
	"github.com/gostack/cli/internal/replacer"
	"github.com/gostack/cli/internal/runner"
	"github.com/gostack/cli/internal/scaffold"
	"github.com/gostack/cli/internal/template"
	"github.com/gostack/cli/internal/wizard"
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
	// Pre-fill project name if passed as argument
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

	// Guard: refuse to overwrite existing directory
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("directory '%s' already exists", cfg.ProjectName)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}

	// --- Try remote template first, fall back to built-in scaffold ---
	printer.Step("📦", fmt.Sprintf("Fetching template: %s-%s ...", cfg.Framework, cfg.Architecture))

	remoteErr := template.Download(cfg.Framework, cfg.Architecture, destDir)
	if remoteErr != nil {
		printer.Warn(fmt.Sprintf("Remote template unavailable (%v), using built-in scaffold ...", remoteErr))

		if err := scaffold.Generate(destDir, scaffold.Config{
			ProjectName:  cfg.ProjectName,
			ModuleName:   cfg.ModuleName,
			Framework:    cfg.Framework,
			Architecture: cfg.Architecture,
			Database:     cfg.Database,
			ORM:          cfg.ORM,
			Auth:         cfg.Auth,
			Docker:       cfg.Docker,
			Swagger:      cfg.Swagger,
		}); err != nil {
			os.RemoveAll(destDir)
			return fmt.Errorf("scaffold: %w", err)
		}
	} else {
		// Replace placeholders in downloaded template
		printer.Step("🔧", "Replacing placeholders ...")
		if err := replacer.ReplaceAll(destDir, replacer.Config{
			ModuleName:  cfg.ModuleName,
			ProjectName: cfg.ProjectName,
		}); err != nil {
			return fmt.Errorf("replace placeholders: %w", err)
		}
	}

	// --- git init ---
	printer.Step("🗂 ", "Initializing git ...")
	if err := runner.GitInit(destDir); err != nil {
		printer.Warn("git init skipped: " + err.Error())
	}

	// --- go mod tidy ---
	printer.Step("🧹", "Running go mod tidy ...")
	if err := runner.GoModTidy(destDir); err != nil {
		printer.Warn("go mod tidy skipped (run it manually): " + err.Error())
	}

	// --- Done ---
	printer.Summary(cfg.ProjectName, destDir)
	return nil
}

// resolveConfig decides whether to use wizard or CLI flags.
func resolveConfig(projectName string) (*wizard.ProjectConfig, error) {
	f := createFlags

	// Non-interactive: all required flags provided
	if f.framework != "" && f.architecture != "" && f.database != "" && f.orm != "" {
		name := projectName
		if name == "" {
			return nil, fmt.Errorf("project name required when using flags (e.g. gostack create my-api --framework gin ...)")
		}
		moduleName := f.moduleName
		if moduleName == "" {
			moduleName = name
		}
		return &wizard.ProjectConfig{
			ProjectName:  name,
			ModuleName:   moduleName,
			Framework:    f.framework,
			Architecture: f.architecture,
			Database:     f.database,
			ORM:          f.orm,
			Auth:         f.auth,
			Docker:       f.docker,
			Swagger:      f.swagger,
		}, nil
	}

	// Interactive wizard
	printer.Step("🧙", "Starting project wizard ...")
	cfg, err := wizard.Run(projectName)
	if err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}
	return cfg, nil
}
