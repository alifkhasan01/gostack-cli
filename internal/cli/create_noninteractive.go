package cli

// flagConfig holds values passed via --flag for non-interactive mode.
type flagConfig struct {
	moduleName     string
	projectType    string
	framework      string
	architecture   string
	database       string
	orm            string
	auth           string
	docker         bool
	swagger        bool
	cliLib         string
	services       string
	cssFramework   string
	templateEngine string
}

func init() {
	createCmd.Flags().StringVar(&createFlags.moduleName, "module", "", "Go module path (e.g. github.com/you/my-api)")
	createCmd.Flags().StringVar(&createFlags.projectType, "type", "", "Project type: rest-api|cli|microservice|fullstack")
	createCmd.Flags().StringVar(&createFlags.framework, "framework", "", "Framework: gin|fiber|echo|chi (rest-api)")
	createCmd.Flags().StringVar(&createFlags.architecture, "arch", "", "Architecture: standard|clean|hexagonal|ddd (rest-api)")
	createCmd.Flags().StringVar(&createFlags.database, "database", "", "Database: postgres|mysql|sqlite|none")
	createCmd.Flags().StringVar(&createFlags.orm, "orm", "", "ORM: gorm|bun|sqlx (rest-api)")
	createCmd.Flags().StringVar(&createFlags.auth, "auth", "", "Auth: jwt|session|none (rest-api)")
	createCmd.Flags().StringVar(&createFlags.cliLib, "cli-lib", "", "CLI library: cobra|plain (cli)")
	createCmd.Flags().StringVar(&createFlags.services, "services", "", "Service names, comma-separated (microservice)")
	createCmd.Flags().StringVar(&createFlags.cssFramework, "css", "", "CSS framework: tailwind|bootstrap|none (fullstack)")
	createCmd.Flags().StringVar(&createFlags.templateEngine, "templ", "", "Template engine: html|templ (fullstack)")
	createCmd.Flags().BoolVar(&createFlags.docker, "docker", false, "Include Docker files")
	createCmd.Flags().BoolVar(&createFlags.swagger, "swagger", false, "Include Swagger setup")
}

// createFlags is populated from CLI flags.
var createFlags flagConfig
