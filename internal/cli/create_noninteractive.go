package cli

// flagConfig holds values passed via --flag for non-interactive mode.
type flagConfig struct {
	moduleName   string
	framework    string
	architecture string
	database     string
	orm          string
	auth         string
	docker       bool
	swagger      bool
}

func init() {
	createCmd.Flags().StringVar(&createFlags.moduleName, "module", "", "Go module path (e.g. github.com/you/my-api)")
	createCmd.Flags().StringVar(&createFlags.framework, "framework", "", "Framework: gin|fiber|echo|chi")
	createCmd.Flags().StringVar(&createFlags.architecture, "arch", "", "Architecture: standard|clean|hexagonal|ddd")
	createCmd.Flags().StringVar(&createFlags.database, "database", "", "Database: postgres|mysql|sqlite")
	createCmd.Flags().StringVar(&createFlags.orm, "orm", "", "ORM: gorm|bun|sqlx")
	createCmd.Flags().StringVar(&createFlags.auth, "auth", "", "Auth: jwt|session|none")
	createCmd.Flags().BoolVar(&createFlags.docker, "docker", false, "Include Docker files")
	createCmd.Flags().BoolVar(&createFlags.swagger, "swagger", false, "Include Swagger setup")
}

// createFlags is populated from CLI flags.
var createFlags flagConfig
