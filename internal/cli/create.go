package cli

import (
	"fmt"
	"os"
	"path/filepath"

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
	sp := printer.NewSpinner(fmt.Sprintf("Fetching template: %s-%s", cfg.Framework, cfg.Architecture))
	sp.Start()

	remoteErr := template.Download(cfg.Framework, cfg.Architecture, destDir)
	if remoteErr != nil {
		sp.Fail(fmt.Sprintf("Remote template unavailable, using built-in scaffold (%s-%s)", cfg.Framework, cfg.Architecture))

		sp2 := printer.NewSpinner("Generating project structure ...")
		sp2.Start()
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
			sp2.Fail("Scaffold failed")
			os.RemoveAll(destDir)
			return fmt.Errorf("scaffold: %w", err)
		}
		sp2.Done("Project structure generated")
	} else {
		sp.Done(fmt.Sprintf("Template downloaded: %s-%s", cfg.Framework, cfg.Architecture))

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
