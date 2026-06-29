package wizard

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// ProjectConfig holds all user selections from the wizard.
type ProjectConfig struct {
	ProjectName  string
	ModuleName   string
	Framework    string
	Architecture string
	Database     string
	ORM          string
	Auth         string
	Docker       bool
	Swagger      bool
}

// Run shows the interactive wizard and returns a filled ProjectConfig.
func Run(projectName string) (*ProjectConfig, error) {
	cfg := &ProjectConfig{
		ProjectName: projectName,
	}

	// --- Step 1: Project name (if not provided via CLI arg) ---
	nameGroup := huh.NewGroup(
		huh.NewInput().
			Title("Project Name").
			Placeholder("my-api").
			Value(&cfg.ProjectName).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("project name cannot be empty")
				}
				return nil
			}),
		huh.NewInput().
			Title("Module Name").
			Description("Go module path (e.g. github.com/yourname/my-api)").
			Placeholder("github.com/yourname/my-api").
			Value(&cfg.ModuleName).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("module name cannot be empty")
				}
				return nil
			}),
	)

	// --- Step 2: Framework ---
	frameworkGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Framework").
			Options(
				huh.NewOption("Gin", "gin"),
				huh.NewOption("Fiber", "fiber"),
				huh.NewOption("Echo", "echo"),
				huh.NewOption("Chi", "chi"),
			).
			Value(&cfg.Framework),
	)

	// --- Step 3: Architecture ---
	archGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Architecture").
			Options(
				huh.NewOption("Standard", "standard"),
				huh.NewOption("Clean", "clean"),
				huh.NewOption("Hexagonal", "hexagonal"),
				huh.NewOption("DDD", "ddd"),
			).
			Value(&cfg.Architecture),
	)

	// --- Step 4: Database ---
	dbGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Database").
			Options(
				huh.NewOption("PostgreSQL", "postgres"),
				huh.NewOption("MySQL", "mysql"),
				huh.NewOption("SQLite", "sqlite"),
			).
			Value(&cfg.Database),
	)

	// --- Step 5: ORM ---
	ormGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("ORM / Query Builder").
			Options(
				huh.NewOption("GORM", "gorm"),
				huh.NewOption("Bun", "bun"),
				huh.NewOption("SQLX", "sqlx"),
			).
			Value(&cfg.ORM),
	)

	// --- Step 6: Authentication ---
	authGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Authentication").
			Options(
				huh.NewOption("JWT", "jwt"),
				huh.NewOption("Session", "session"),
				huh.NewOption("None", "none"),
			).
			Value(&cfg.Auth),
	)

	// --- Step 7: Docker & Swagger ---
	extrasGroup := huh.NewGroup(
		huh.NewConfirm().
			Title("Include Docker?").
			Value(&cfg.Docker),
		huh.NewConfirm().
			Title("Include Swagger?").
			Value(&cfg.Swagger),
	)

	form := huh.NewForm(
		nameGroup,
		frameworkGroup,
		archGroup,
		dbGroup,
		ormGroup,
		authGroup,
		extrasGroup,
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return nil, err
	}

	// Default module name to project name if still empty
	if cfg.ModuleName == "" {
		cfg.ModuleName = cfg.ProjectName
	}

	return cfg, nil
}
