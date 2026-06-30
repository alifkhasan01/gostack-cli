package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// ProjectConfig holds all user selections from the wizard.
type ProjectConfig struct {
	ProjectName    string
	ModuleName     string
	Type           string // rest-api, cli, microservice, fullstack
	Framework      string
	Architecture   string
	Database       string
	ORM            string
	Auth           string
	Docker         bool
	Swagger        bool
	CLILib         string // cobra, plain — for cli type
	Services       string // comma-separated — for microservice type
	CSSFramework   string // tailwind, bootstrap, none — for fullstack
	TemplateEngine string // html, templ — for fullstack
}

// hide is a helper that returns a hide-func checking cfg.Type.
func hide(t string, cfg *ProjectConfig) func() bool {
	return func() bool { return cfg.Type != t }
}

// Run shows the interactive wizard and returns a filled ProjectConfig.
func Run(projectName string) (*ProjectConfig, error) {
	cfg := &ProjectConfig{
		ProjectName: projectName,
	}

	groups := []*huh.Group{
		// --- Step 1: Project name ---
		huh.NewGroup(
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
		),

		// --- Step 2: Project Type ---
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Project Type").
				Options(
					huh.NewOption("REST API", "rest-api"),
					huh.NewOption("CLI Tool", "cli"),
					huh.NewOption("Microservice", "microservice"),
					huh.NewOption("Full-Stack Web", "fullstack"),
				).
				Value(&cfg.Type),
		),

		// --- REST API groups ---
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Framework").
				Options(
					huh.NewOption("Gin", "gin"),
					huh.NewOption("Fiber", "fiber"),
					huh.NewOption("Echo", "echo"),
					huh.NewOption("Chi", "chi"),
				).
				Value(&cfg.Framework),
		).WithHideFunc(hide("rest-api", cfg)),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Architecture").
				Options(
					huh.NewOption("Standard", "standard"),
					huh.NewOption("Clean", "clean"),
					huh.NewOption("Hexagonal", "hexagonal"),
					huh.NewOption("DDD", "ddd"),
				).
				Value(&cfg.Architecture),
		).WithHideFunc(hide("rest-api", cfg)),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Database").
				Options(
					huh.NewOption("PostgreSQL", "postgres"),
					huh.NewOption("MySQL", "mysql"),
					huh.NewOption("SQLite", "sqlite"),
				).
				Value(&cfg.Database),
		).WithHideFunc(hide("rest-api", cfg)),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("ORM / Query Builder").
				Options(
					huh.NewOption("GORM", "gorm"),
					huh.NewOption("Bun", "bun"),
					huh.NewOption("SQLX", "sqlx"),
				).
				Value(&cfg.ORM),
		).WithHideFunc(hide("rest-api", cfg)),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Authentication").
				Options(
					huh.NewOption("JWT", "jwt"),
					huh.NewOption("Session", "session"),
					huh.NewOption("None", "none"),
				).
				Value(&cfg.Auth),
		).WithHideFunc(hide("rest-api", cfg)),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Include Docker?").
				Value(&cfg.Docker),
			huh.NewConfirm().
				Title("Include Swagger?").
				Value(&cfg.Swagger),
		).WithHideFunc(hide("rest-api", cfg)),

		// --- CLI groups ---
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("CLI Library").
				Options(
					huh.NewOption("Cobra", "cobra"),
					huh.NewOption("Plain (no dependencies)", "plain"),
				).
				Value(&cfg.CLILib),
		).WithHideFunc(hide("cli", cfg)),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include Docker?").
				Value(&cfg.Docker),
		).WithHideFunc(hide("cli", cfg)),

		// --- Microservice groups ---
		huh.NewGroup(
			huh.NewInput().
				Title("Service Names").
				Description("Comma-separated list of service names").
				Placeholder("user,order,notification").
				Value(&cfg.Services).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("at least one service is required")
					}
					return nil
				}),
		).WithHideFunc(hide("microservice", cfg)),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include API Gateway?").
				Value(&cfg.Docker),
			huh.NewConfirm().
				Title("Include Kubernetes manifests?").
				Value(&cfg.Swagger),
		).WithHideFunc(hide("microservice", cfg)),

		// --- Fullstack groups ---
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Template Engine").
				Options(
					huh.NewOption("html/template (stdlib)", "html"),
					huh.NewOption("Templ", "templ"),
				).
				Value(&cfg.TemplateEngine),
		).WithHideFunc(hide("fullstack", cfg)),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("CSS Framework").
				Options(
					huh.NewOption("Tailwind CSS", "tailwind"),
					huh.NewOption("Bootstrap", "bootstrap"),
					huh.NewOption("None (plain CSS)", "none"),
				).
				Value(&cfg.CSSFramework),
		).WithHideFunc(hide("fullstack", cfg)),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Database").
				Options(
					huh.NewOption("PostgreSQL", "postgres"),
					huh.NewOption("MySQL", "mysql"),
					huh.NewOption("SQLite", "sqlite"),
					huh.NewOption("None", "none"),
				).
				Value(&cfg.Database),
		).WithHideFunc(hide("fullstack", cfg)),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include Docker?").
				Value(&cfg.Docker),
		).WithHideFunc(hide("fullstack", cfg)),
	}

	form := huh.NewForm(groups...).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return nil, err
	}

	// Default module name to project name if still empty
	if cfg.ModuleName == "" {
		cfg.ModuleName = cfg.ProjectName
	}

	// Set defaults for fields not covered by the selected type's groups
	if cfg.Type == "cli" {
		if cfg.Database == "" {
			cfg.Database = "none"
		}
		if cfg.ORM == "" {
			cfg.ORM = "none"
		}
		if cfg.Auth == "" {
			cfg.Auth = "none"
		}
	}
	if cfg.Type == "microservice" {
		if cfg.Services == "" {
			cfg.Services = "user,order,notification"
		}
		if cfg.Database == "" {
			cfg.Database = "none"
		}
		if cfg.ORM == "" {
			cfg.ORM = "none"
		}
		if cfg.Auth == "" {
			cfg.Auth = "none"
		}
		if cfg.Framework == "" {
			cfg.Framework = "none"
		}
	}
	if cfg.Type == "fullstack" {
		if cfg.ORM == "" {
			cfg.ORM = "none"
		}
		if cfg.Auth == "" {
			cfg.Auth = "none"
		}
		if cfg.Framework == "" {
			cfg.Framework = "none"
		}
	}

	return cfg, nil
}

// ServicesList splits cfg.Services by comma and trims whitespace.
func (cfg *ProjectConfig) ServicesList() []string {
	if cfg.Services == "" {
		return nil
	}
	parts := strings.Split(cfg.Services, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
