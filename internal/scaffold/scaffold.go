// Package scaffold generates a project structure locally when a remote
// template is not available (or as a standalone fallback).
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alifkhasan01/gostack-cli/internal/project"
)

// Config mirrors wizard.ProjectConfig — kept separate to avoid circular deps.
type Config struct {
	ProjectName    string
	ModuleName     string
	Type           string
	Framework      string
	Architecture   string
	Database       string
	ORM            string
	Auth           string
	Docker         bool
	Swagger        bool
	Version        string
	CLILib         string
	Services       string
	CSSFramework   string
	TemplateEngine string
}

// ServicesList splits cfg.Services by comma and trims whitespace.
func (cfg Config) ServicesList() []string {
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

// Generate writes a full project skeleton into destDir.
func Generate(destDir string, cfg Config) error {
	files := buildFileMap(cfg)
	for relPath, content := range files {
		fullPath := filepath.Join(destDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}

		var rendered string
		var err error
		if strings.HasSuffix(relPath, ".goreleaser.yml") {
			// Pre-rendered with fmt.Sprintf — skip template render
			rendered = content
		} else {
			rendered, err = renderTemplate(content, cfg)
			if err != nil {
				return fmt.Errorf("render %s: %w", relPath, err)
			}
		}

		if err := os.WriteFile(fullPath, []byte(rendered), 0644); err != nil {
			return err
		}
		fmt.Printf("  create  %s\n", relPath)
	}

	// Write gostack.json for generator commands to read later
	return project.Write(destDir, project.Meta{
		ProjectName:  cfg.ProjectName,
		ModuleName:   cfg.ModuleName,
		Type:         cfg.Type,
		Framework:    cfg.Framework,
		Architecture: cfg.Architecture,
		Database:     cfg.Database,
		ORM:          cfg.ORM,
		Auth:         cfg.Auth,
		GoStackVer:   cfg.Version,
	})
}

// renderTemplate executes a Go text/template string with cfg as data.
func renderTemplate(tmplStr string, cfg Config) (string, error) {
	t, err := template.New("").Funcs(template.FuncMap{
		"title": strings.Title, //nolint:staticcheck
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}).Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := t.Execute(&sb, cfg); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// dbDriver returns the Go import path for the chosen database driver.
func dbDriver(cfg Config) string {
	switch cfg.Database {
	case "mysql":
		return "github.com/go-sql-driver/mysql"
	case "sqlite":
		return "github.com/mattn/go-sqlite3"
	default: // postgres
		return "github.com/lib/pq"
	}
}

// ormImport returns the ORM import path.
func ormImport(cfg Config) string {
	switch cfg.ORM {
	case "bun":
		return "github.com/uptrace/bun"
	case "sqlx":
		return "github.com/jmoiron/sqlx"
	default: // gorm
		return "gorm.io/gorm"
	}
}

// frameworkImport returns the HTTP framework import path.
func frameworkImport(cfg Config) string {
	switch cfg.Framework {
	case "fiber":
		return "github.com/gofiber/fiber/v2"
	case "echo":
		return "github.com/labstack/echo/v4"
	case "chi":
		return "github.com/go-chi/chi/v5"
	default: // gin
		return "github.com/gin-gonic/gin"
	}
}

// buildFileMap returns all files to be created, keyed by relative path.
func buildFileMap(cfg Config) map[string]string {
	switch cfg.Type {
	case "cli":
		return buildCLIFileMap(cfg)
	case "microservice":
		return buildMicroserviceFileMap(cfg)
	case "fullstack":
		return buildFullstackFileMap(cfg)
	default:
		return buildRESTAPIFileMap(cfg)
	}
}

func buildRESTAPIFileMap(cfg Config) map[string]string {
	files := map[string]string{
		"go.mod": goModTmpl,
		"cmd/server/main.go": mainGoTmpl,
		"internal/config/config.go": configGoTmpl,
		"internal/model/user.go": modelGoTmplFn(cfg),
		"internal/handler/user.go": handlerGoTmpl(cfg.Framework),
		"internal/service/user.go": serviceGoTmpl,
		"internal/repository/user.go": repositoryGoTmpl,
		"internal/middleware/logger.go": loggerMidTmpl(cfg.Framework),
		"internal/middleware/auth.go":   authMidTmpl(cfg.Framework),
		"internal/handler/user_test.go": handlerTestGoTmpl,
		"internal/service/user_test.go": serviceTestTmpl,
		"pkg/response/response.go":   responsePkgTmpl,
		"pkg/validator/validator.go": validatorPkgTmpl,
		"migrations/.gitkeep": "",
		"docs/.gitkeep": "",
		".env":         envTmpl,
		".env.example": envTmpl,
		".gitignore": gitignoreTmpl,
		"Makefile": makefileTmpl,
		"README.md": readmeTmpl,
	}

	if cfg.Docker {
		files["Dockerfile"] = dockerfileTmpl(cfg.Framework)
		files["docker-compose.yml"] = dockerComposeTmpl
	}

	files[".github/workflows/ci.yml"] = ciGoTmpl

	if cfg.Auth == "jwt" {
		files["pkg/jwt/jwt.go"] = jwtPkgTmpl
	}

	return files
}

// ---------------------------------------------------------------------------
// CLI tool scaffold
// ---------------------------------------------------------------------------

func buildCLIFileMap(cfg Config) map[string]string {
	cliLib := cfg.CLILib
	if cliLib == "" {
		cliLib = "cobra"
	}

	files := map[string]string{
		"go.mod": goModTmpl,
		".gitignore":    gitignoreTmpl,
		"Makefile":      cliMakefileTmpl,
		"README.md":     cliReadmeTmpl,
		".goreleaser.yml": goreleaserTmplFunc(cfg.ProjectName, cfg.Version),
	}

	if cliLib == "cobra" {
		files["cmd/root.go"] = cobraRootTmpl
		files["cmd/init.go"] = cobraInitTmpl
		files["cmd/build.go"] = cobraBuildTmpl
		files["cmd/run.go"] = cobraRunTmpl
		files["main.go"] = cobraMainTmpl
	} else {
		files["cmd/root.go"] = plainRootTmpl
		files["main.go"] = plainMainTmpl
	}

	files["internal/config/config.go"] = cliConfigTmpl
	files["internal/runner/runner.go"] = cliRunnerTmpl
	files["internal/output/printer.go"] = cliPrinterTmpl
	files["pkg/util/file.go"] = cliUtilFileTmpl

	if cfg.Docker {
		files["Dockerfile"] = cliDockerfileTmpl
	}

	files[".github/workflows/ci.yml"] = ciGoTmpl

	return files
}

// ---------------------------------------------------------------------------
// Microservice scaffold
// ---------------------------------------------------------------------------

func buildMicroserviceFileMap(cfg Config) map[string]string {
	files := map[string]string{
		"go.mod": goModTmpl,
		"Makefile": msMakefileTmpl,
		".gitignore":  gitignoreTmpl,
		"README.md":   msReadmeTmpl,
	}

	// Shared library
	files["shared/go.mod"] = sharedGoModTmpl
	files["shared/logger/logger.go"] = sharedLoggerTmpl
	files["shared/middleware/auth.go"] = sharedMidAuthTmpl
	files["shared/proto/.gitkeep"] = ""

	services := cfg.ServicesList()
	if len(services) == 0 {
		services = []string{"user", "order", "notification"}
	}

	for _, svc := range services {
		prefix := "services/" + svc + "-service"
		svcModule := cfg.ModuleName + "/" + prefix

		files[prefix+"/cmd/main.go"] = msMainTmpl(svcModule, svc)
		files[prefix+"/internal/handler/"+svc+".go"] = msHandlerTmpl(svcModule, svc)
		files[prefix+"/internal/service/"+svc+".go"] = msServiceTmpl(svcModule, svc)
		files[prefix+"/internal/repository/"+svc+".go"] = msRepoTmpl(svcModule, svc)
		files[prefix+"/internal/model/"+svc+".go"] = msModelTmpl(svc)
		files[prefix+"/Dockerfile"] = msSvcDockerfileTmpl
		files[prefix+"/go.mod"] = fmt.Sprintf(msSvcGoModTmpl, svcModule)
	}

	// Gateway
	files["gateway/cmd/main.go"] = gwMainTmpl
	files["gateway/internal/proxy/proxy.go"] = gwProxyTmpl
	files["gateway/go.mod"] = fmt.Sprintf(msSvcGoModTmpl, cfg.ModuleName+"/gateway")

	// Infra
	files["infra/docker-compose.yml"] = msDockerComposeTmpl

	// Kubernetes manifests
	for _, svc := range services {
		files["infra/k8s/"+svc+"-deployment.yaml"] = k8sDeploymentTmpl(svc)
	}
	files["infra/nginx/nginx.conf"] = nginxConfTmpl

	return files
}

// ---------------------------------------------------------------------------
// Fullstack scaffold
// ---------------------------------------------------------------------------

func buildFullstackFileMap(cfg Config) map[string]string {
	files := map[string]string{
		"go.mod": goModTmpl,
		"cmd/web/main.go": fwMainGoTmpl,
		"internal/handler/home.go": fwHomeHandlerTmpl,
		"internal/handler/auth.go": fwAuthHandlerTmpl,
		"internal/template/renderer.go": fwRendererTmpl,
		"web/static/css/style.css": fwCSSTmpl(cfg.CSSFramework),
		"web/static/js/app.js": fwJSTmpl,
		"web/templates/layout/base.html": fwBaseTmpl,
		"web/templates/pages/home.html": fwHomePageTmpl,
		"web/templates/components/navbar.html": fwNavbarTmpl,
		"migrations/.gitkeep": "",
		".env.example": fwEnvTmpl,
		".env": fwEnvTmpl,
		".gitignore": gitignoreTmpl,
		"Makefile": fwMakefileTmpl,
		"README.md": fwReadmeTmpl,
	}

	if cfg.Database != "none" {
		databaseCfg := cfg
		databaseCfg.Framework = "gin"
		files["internal/model/user.go"] = modelGoTmplFn(databaseCfg)
		files["internal/repository/user.go"] = repositoryGoTmpl
		files["internal/service/user.go"] = serviceGoTmpl
	}

	if cfg.Docker {
		files["Dockerfile"] = fwDockerfileTmpl
	}

	files[".github/workflows/ci.yml"] = ciGoTmpl

	return files
}

// ============================================================
// Template strings
// ============================================================

const goModTmpl = `module {{.ModuleName}}

go 1.22
`

var mainGoTmpl = `package main

import (
	"log"

	"{{.ModuleName}}/internal/config"
	"{{.ModuleName}}/internal/model"
	"{{.ModuleName}}/internal/handler"
	"{{.ModuleName}}/internal/middleware"
	"{{.ModuleName}}/internal/repository"
	"{{.ModuleName}}/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := model.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)

	r := handler.SetupRoutes(cfg, db, h)

	log.Printf("Server running on %s", cfg.AppAddr)
	handler.StartServer(r, cfg.AppAddr)
}
`

const configGoTmpl = `package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	AppName string
	AppAddr string
	DBDriver string
	DBDSN    string
}

// Load reads environment variables and returns a Config.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env")
	}

	return &Config{
		AppName:  getEnv("APP_NAME", "{{.ProjectName}}"),
		AppAddr:  ":" + getEnv("APP_PORT", "8080"),
		DBDriver: getEnv("DB_DRIVER", "{{.Database}}"),
		DBDSN:    getEnv("DB_DSN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
`

func modelGoTmplFn(cfg Config) string {
	return fmt.Sprintf(`package model

import (
	"database/sql"
	"fmt"
	"time"

	"{{.ModuleName}}/internal/config"
	_ %q
)

type User struct {
	ID        int       `+"`"+`json:"id" db:"id"`+"`"+`
	Name      string    `+"`"+`json:"name" db:"name"`+"`"+`
	Email     string    `+"`"+`json:"email" db:"email"`+"`"+`
	CreatedAt time.Time `+"`"+`json:"created_at" db:"created_at"`+"`"+`
	UpdatedAt time.Time `+"`"+`json:"updated_at" db:"updated_at"`+"`"+`
}

func ConnectDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %%w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %%w", err)
	}
	return db, nil
}
`, dbDriverImport(cfg))
}

func dbDriverImport(cfg Config) string {
	switch cfg.Database {
	case "mysql":
		return "github.com/go-sql-driver/mysql"
	case "sqlite":
		return "github.com/mattn/go-sqlite3"
	default:
		return "github.com/lib/pq"
	}
}

// ---------------------------------------------------------------------------
// Middleware templates
// ---------------------------------------------------------------------------

func loggerMidTmpl(framework string) string {
	switch framework {
	case "fiber":
		return `package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Printf("[%d] %s %s %s", c.Response().StatusCode(), c.Method(), c.Path(), time.Since(start))
		return err
	}
}
`
	case "echo":
		return `package middleware

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
)

func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			log.Printf("[%d] %s %s %s", c.Response().Status, c.Request().Method, c.Path(), time.Since(start))
			return err
		}
	}
}
`
	case "chi":
		return `package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%d] %s %s %s", 200, r.Method, r.URL.Path, time.Since(start))
	})
}
`
	default: // gin
		return `package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		log.Printf("[%d] %s %s %s", status, method, path, latency)
	}
}
`
	}
}

func authMidTmpl(framework string) string {
	switch framework {
	case "fiber":
		return `package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get("Authorization") == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Next()
	}
}
`
	case "echo":
		return `package middleware

import (
	"github.com/labstack/echo/v4"
)

func Auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Header.Get("Authorization") == "" {
				return echo.NewHTTPError(401, "unauthorized")
			}
			return next(c)
		}
	}
}
`
	case "chi":
		return `package middleware

import (
	"net/http"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
`
	default: // gin
		return `package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
`
	}
}

// ---------------------------------------------------------------------------
// Handler template with SetupRoutes & StartServer
// ---------------------------------------------------------------------------

func handlerGoTmpl(framework string) string {
	switch framework {
	case "fiber":
		return `package handler

import (
	"database/sql"
	"log"

	"{{.ModuleName}}/internal/config"
	"{{.ModuleName}}/internal/middleware"
	"{{.ModuleName}}/internal/service"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func SetupRoutes(cfg *config.Config, db *sql.DB, h *UserHandler) *fiber.App {
	app := fiber.New()
	app.Use(middleware.Logger())
	app.Use(middleware.Auth())

	api := app.Group("/api/v1")
	api.Get("/health", func(c *fiber.Ctx) error {
		if err := db.Ping(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not ready"})
		}
		return c.JSON(fiber.Map{"status": "ok", "app": cfg.AppName})
	})

	// gostack:routes

	return app
}

func StartServer(app *fiber.App, addr string) {
	log.Printf("Server running on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func (h *UserHandler) List(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "list users"})
}
`
	case "echo":
		return `package handler

import (
	"database/sql"
	"log"
	"net/http"

	"{{.ModuleName}}/internal/config"
	"{{.ModuleName}}/internal/middleware"
	"{{.ModuleName}}/internal/service"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func SetupRoutes(cfg *config.Config, db *sql.DB, h *UserHandler) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Auth())

	e.GET("/api/v1/health", func(c echo.Context) error {
		if err := db.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "app": cfg.AppName})
	})

	// gostack:routes

	return e
}

func StartServer(e *echo.Echo, addr string) {
	log.Printf("Server running on %s", addr)
	if err := e.Start(addr); err != nil {
		log.Fatal(err)
	}
}

func (h *UserHandler) List(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "list users"})
}
`
	case "chi":
		return `package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"{{.ModuleName}}/internal/config"
	"{{.ModuleName}}/internal/middleware"
	"{{.ModuleName}}/internal/service"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func SetupRoutes(cfg *config.Config, db *sql.DB, h *UserHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Auth)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := db.Ping(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{"status": "not ready"})
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ok", "app": cfg.AppName})
		})
	})

	// gostack:routes

	return r
}

func StartServer(r *chi.Mux, addr string) {
	log.Printf("Server running on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "list users"})
}
`
	default: // gin
		return `package handler

import (
	"database/sql"
	"log"
	"net/http"

	"{{.ModuleName}}/internal/config"
	"{{.ModuleName}}/internal/middleware"
	"{{.ModuleName}}/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func SetupRoutes(cfg *config.Config, db *sql.DB, h *UserHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Auth())

	r.GET("/api/v1/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "app": cfg.AppName})
	})

	// gostack:routes

	return r
}

func StartServer(r *gin.Engine, addr string) {
	log.Printf("Server running on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func (h *UserHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "list users"})
}
`
	}
}

// ---------------------------------------------------------------------------
// Service & Repository templates
// ---------------------------------------------------------------------------

const serviceGoTmpl = `package service

import (
	"{{.ModuleName}}/internal/repository"
)

type UserService interface {
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}
`

const repositoryGoTmpl = `package repository

import (
	"database/sql"
)

type UserRepository interface {
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}
`

// ---------------------------------------------------------------------------
// Test templates
// ---------------------------------------------------------------------------

const handlerTestGoTmpl = `package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserHandler_List(t *testing.T) {
	h := &UserHandler{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/users", nil)

	h.List(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["message"] != "list users" {
		t.Errorf("unexpected message: %s", resp["message"])
	}
}
`

const serviceTestTmpl = `package service

import "testing"

func TestNewUserService(t *testing.T) {
	_ = NewUserService(nil)
}
`

// ---------------------------------------------------------------------------
// pkg templates
// ---------------------------------------------------------------------------

const responsePkgTmpl = `package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Status  string      ` + "`json:\"status\"`" + `
	Message string      ` + "`json:\"message,omitempty\"`" + `
	Data    interface{} ` + "`json:\"data,omitempty\"`" + `
	Meta    *Meta       ` + "`json:\"meta,omitempty\"`" + `
	Errors  interface{} ` + "`json:\"errors,omitempty\"`" + `
}

type Meta struct {
	Page       int ` + "`json:\"page\"`" + `
	PerPage    int ` + "`json:\"per_page\"`" + `
	Total      int ` + "`json:\"total\"`" + `
	TotalPages int ` + "`json:\"total_pages\"`" + `
}

func writeJSON(w http.ResponseWriter, status int, data APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, APIResponse{Status: "success", Data: data})
}

func Created(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusCreated, APIResponse{Status: "success", Data: data})
}

func ErrorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse{Status: "error", Message: message})
}

func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "resource not found"
	}
	ErrorJSON(w, http.StatusNotFound, message)
}

func InternalError(w http.ResponseWriter) {
	ErrorJSON(w, http.StatusInternalServerError, "internal server error")
}
`

const validatorPkgTmpl = `package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type ValidationError struct {
	Field   string ` + "`json:\"field\"`" + `
	Message string ` + "`json:\"message\"`" + `
}

func Validate(i interface{}) []ValidationError {
	var errs []ValidationError
	err := validate.Struct(i)
	if err == nil {
		return nil
	}
	for _, verr := range err.(validator.ValidationErrors) {
		errs = append(errs, ValidationError{
			Field:   toSnakeCase(verr.Field()),
			Message: messageForTag(verr.Tag(), verr.Param()),
		})
	}
	return errs
}

func messageForTag(tag, param string) string {
	switch tag {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", param)
	case "max":
		return fmt.Sprintf("must be at most %s characters", param)
	default:
		return fmt.Sprintf("validation failed on '%s'", tag)
	}
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
`

const jwtPkgTmpl = `package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var secret = []byte("change-me-in-production")

type Claims struct {
	UserID int ` + "`json:\"user_id\"`" + `
	jwtlib.RegisteredClaims
}

func GenerateToken(userID int) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtlib.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
`

const ciGoTmpl = `name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - run: go mod tidy
      - run: go vet ./...
      - run: go test ./... -v -count=1
`

const envTmpl = `APP_NAME={{.ProjectName}}
APP_PORT=8080

DB_DRIVER={{.Database}}
DB_DSN=

JWT_SECRET=change-me-in-production
`

const gitignoreTmpl = `.env
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
/bin/
/dist/
vendor/
`

const makefileTmpl = `APP_NAME={{.ProjectName}}
BIN_DIR=./bin

.PHONY: run build tidy lint test clean

run:
	go run ./cmd/server

build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/server

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

test:
	go test ./... -v

clean:
	rm -rf $(BIN_DIR)
`

const readmeTmpl = `# {{.ProjectName}}

Generated with [GoStack CLI](https://github.com/alifkhasan01/gostack-cli).

## Stack

| Layer      | Choice              |
|------------|---------------------|
| Framework  | {{.Framework}}      |
| Database   | {{.Database}}       |
| Auth       | {{.Auth}}           |

## Getting Started

` + "```bash" + `
cp .env.example .env
# edit .env with your DB credentials

go mod tidy
go run ./cmd/server
` + "```" + `

## Structure

` + "```" + `
.
├── cmd/server/       # entry point
├── internal/
│   ├── config/       # environment config
│   ├── handler/      # HTTP handlers + routes
│   ├── middleware/    # HTTP middlewares
│   ├── model/        # domain models + DB connection
│   ├── repository/   # data access layer
│   └── service/      # business logic
├── pkg/
│   ├── jwt/          # JWT helpers
│   ├── response/     # standard API response
│   └── validator/    # request validation
├── migrations/       # SQL migrations
└── docs/             # API docs
` + "```" + `
`

func dockerfileTmpl(framework string) string {
	_ = framework
	return `# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/server ./cmd/server

# Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
COPY .env.example .env
EXPOSE 8080
CMD ["./server"]
`
}

const dockerComposeTmpl = `version: "3.9"

services:
  app:
    build: .
    ports:
      - "8080:8080"
    env_file:
      - .env
    depends_on:
      - db

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: {{.ProjectName}}
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
`

// =========================================================================
// CLI tool templates
// =========================================================================

const cobraRootTmpl = `package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "{{.ProjectName}}",
	Short: "{{.ProjectName}} - a CLI tool built with GoStack",
	Long: ` + "`" + `{{.ProjectName}} is a CLI application generated by GoStack CLI.` + "`" + `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("{{.ProjectName}} — use --help to see available commands")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
`

const cobraInitTmpl = `package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing project...")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
`

const cobraBuildTmpl = `package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building project...")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
`

const cobraRunTmpl = `package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the project",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running project...")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
`

const cobraMainTmpl = `package main

import (
	"{{.ModuleName}}/cmd"
)

func main() {
	cmd.Execute()
}
`

const plainRootTmpl = `package cmd

import (
	"flag"
	"fmt"
	"os"
)

func Execute() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: {{.ProjectName}} <command>")
		fmt.Println("Available commands: init, build, run")
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "init":
		InitCmd()
	case "build":
		BuildCmd()
	case "run":
		RunCmd()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %%s\n", cmd)
		os.Exit(1)
	}
}

func InitCmd() {
	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	name := initFlags.String("name", "", "project name")
	initFlags.Parse(os.Args[2:])
	fmt.Printf("Initializing project: %%s\n", *name)
}

func BuildCmd() {
	fmt.Println("Building project...")
}

func RunCmd() {
	fmt.Println("Running project...")
}
`

const plainMainTmpl = `package main

import (
	"{{.ModuleName}}/cmd"
)

func main() {
	cmd.Execute()
}
`

const cliConfigTmpl = `package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	AppName string ` + "`json:\"app_name\"`" + `
	Version string ` + "`json:\"version\"`" + `
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cfgPath := filepath.Join(home, "."+appName()+".json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return Default(), nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), nil
	}
	return &cfg, nil
}

func Default() *Config {
	return &Config{
		AppName: "{{.ProjectName}}",
		Version: "0.1.0",
	}
}

func appName() string {
	return "{{.ProjectName}}"
}

func (c *Config) Save() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cfgPath := filepath.Join(home, "."+appName()+".json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	//nolint:gosec
	return os.WriteFile(cfgPath, data, 0644)
}

func init() {
	// Auto-load config on init
	if _, err := Load(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: %%v\n", err)
	}
}
`

const cliRunnerTmpl = `package runner

import (
	"fmt"
	"os/exec"
)

func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("run %%s: %%w", name, err)
	}
	return string(out), nil
}
`

const cliPrinterTmpl = `package output

import (
	"fmt"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

func Info(format string, a ...interface{}) {
	fmt.Print(string(colorCyan))
	fmt.Printf(format, a...)
	fmt.Println(string(colorReset))
}

func Success(format string, a ...interface{}) {
	fmt.Print(string(colorGreen))
	fmt.Printf("✓ "+format, a...)
	fmt.Println(string(colorReset))
}

func Error(format string, a ...interface{}) {
	fmt.Print(string(colorRed))
	fmt.Printf("✗ "+format, a...)
	fmt.Println(string(colorReset))
}

func Warning(format string, a ...interface{}) {
	fmt.Print(string(colorYellow))
	fmt.Printf("! "+format, a...)
	fmt.Println(string(colorReset))
}

func Table(headers []string, rows [][]string) {
	cols := len(headers)
	if cols == 0 {
		return
	}
	widths := make([]int, cols)
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < cols && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	printRow(headers, widths, colorBlue)
	fmt.Println(strings.Repeat("─", sum(widths)+cols*3+1))
	for _, row := range rows {
		printRow(row, widths, "")
	}
}

func printRow(cells []string, widths []int, color string) {
	fmt.Print(color)
	for i, cell := range cells {
		if i >= len(widths) {
			break
		}
		fmt.Printf(" %%-#*s", widths[i], cell)
	}
	fmt.Println(colorReset)
}

func sum(vals []int) int {
	s := 0
	for _, v := range vals {
		s += v
	}
	return s
}
`

const cliUtilFileTmpl = `package util

import (
	"io"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func WriteFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
`

func goreleaserTmplFunc(projectName, version string) string {
	nameTmpl := `{{ .ProjectName }}_
{{- title .Os }}_
{{- if eq .Arch "amd64" }}x86_64
{{- else if eq .Arch "386" }}i386
{{- else }}{{ .Arch }}{{ end }}`
	return fmt.Sprintf(`# .goreleaser.yml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - binary: "%s"
    main: .
    ldflags:
      - -s -w -X main.version=%s
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: >-
      %s
    format_overrides:
      - goos: windows
        format: zip

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
`, projectName, version, nameTmpl)
}

const cliMakefileTmpl = `APP_NAME={{.ProjectName}}
BIN_DIR=./bin

.PHONY: run build tidy lint test clean release

run:
	go run .

build:
	go build -o $(BIN_DIR)/$(APP_NAME) .

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

test:
	go test ./... -v

release:
	goreleaser release --snapshot --skip-publish --clean

clean:
	rm -rf $(BIN_DIR)
`

const cliReadmeTmpl = `# {{.ProjectName}}

A CLI tool built with [GoStack CLI](https://github.com/alifkhasan01/gostack-cli).

## Commands

- **init** — Initialize a new project
- **build** — Build the project
- **run** — Run the project

## Getting Started

` + "```bash" + `
go mod tidy
go build -o bin/{{.ProjectName}} .
./bin/{{.ProjectName}} --help
` + "```" + `

## Structure

` + "```" + `
.
├── cmd/              # CLI commands (cobra)
├── internal/
│   ├── config/       # Configuration loader
│   ├── runner/       # Core logic
│   └── output/       # Terminal output helpers
├── pkg/
│   └── util/         # File utilities
├── main.go           # Entry point
└── .goreleaser.yml   # Release configuration
` + "```" + `
`

const cliDockerfileTmpl = `# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/{{.ProjectName}} .

# Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/{{.ProjectName}} .
ENTRYPOINT ["/app/{{.ProjectName}}"]
`

// =========================================================================
// Microservice templates
// =========================================================================

const msMakefileTmpl = `SERVICES := $(wildcard services/*)

.PHONY: run-all build-all tidy lint test-all clean

run-all:
	@for svc in $(SERVICES); do \
		echo "Starting $$svc..."; \
		(cd $$svc && go run ./cmd) & \
	done; \
	wait

build-all:
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		(cd $$svc && go build -o bin/ ./cmd); \
	done

tidy:
	@for svc in $(SERVICES); do \
		(cd $$svc && go mod tidy); \
	done
	cd gateway && go mod tidy

lint:
	golangci-lint run ./...

test-all:
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		(cd $$svc && go test ./... -v); \
	done

clean:
	rm -rf services/*/bin gateway/bin
`

const msReadmeTmpl = `# {{.ProjectName}}

A microservice project generated with [GoStack CLI](https://github.com/alifkhasan01/gostack-cli).

## Services

- **User Service** — manages users
- **Order Service** — manages orders
- **Notification Service** — sends notifications

## Getting Started

` + "```bash" + `
make run-all
` + "```" + `

## Structure

` + "```" + `
.
├── services/             # individual service modules
│   ├── user-service/
│   ├── order-service/
│   └── notification-service/
├── shared/               # shared library
├── gateway/              # API gateway
├── infra/                # k8s, nginx, docker-compose
└── Makefile
` + "```" + `
`

const sharedGoModTmpl = `module {{.ModuleName}}/shared

go 1.22
`

const sharedLoggerTmpl = `package logger

import (
	"io"
	"log"
	"os"
)

var (
	Info  = log.New(os.Stdout, "INFO ", log.LstdFlags)
	Warn  = log.New(os.Stdout, "WARN ", log.LstdFlags)
	Error = log.New(os.Stderr, "ERROR ", log.LstdFlags)
	Debug = log.New(os.Stdout, "DEBUG ", log.LstdFlags)
)

func Init(debug bool) {
	if !debug {
		Debug.SetOutput(io.Discard)
	}
}
`

const sharedMidAuthTmpl = `package middleware

import (
	"log"
	"net/http"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
`

func msMainTmpl(modPath, svc string) string {
	return fmt.Sprintf(`package main

import (
	"log"

	"%s/internal/handler"
	"%s/internal/repository"
	"%s/internal/service"
)

func main() {
	repo := repository.NewRepository()
	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	log.Println("%s service starting on :8080")
	_ = h
}
`, modPath, modPath, modPath, svc)
}

func msHandlerTmpl(modPath, svc string) string {
	svcUpper := strings.ToUpper(svc[:1]) + svc[1:]
	return fmt.Sprintf(`package handler

import (
	"net/http"

	"%s/internal/service"
)

type %sHandler struct {
	svc service.%sService
}

func NewHandler(svc service.%sService) *%sHandler {
	return &%sHandler{svc: svc}
}

func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[]"))
}
`, modPath, svcUpper, svcUpper, svcUpper, svcUpper, svcUpper, svcUpper)
}

func msServiceTmpl(modPath, svc string) string {
	svcUpper := strings.ToUpper(svc[:1]) + svc[1:]
	return fmt.Sprintf(`package service

import (
	"%s/internal/repository"
)

type %sService interface {
}

type %sService struct {
	repo repository.%sRepository
}

func NewService(repo repository.%sRepository) %sService {
	return &%sService{repo: repo}
}
`, modPath, svcUpper, svcUpper, svcUpper, svcUpper, svcUpper, svcUpper)
}

func msRepoTmpl(modPath, svc string) string {
	svcUpper := strings.ToUpper(svc[:1]) + svc[1:]
	return fmt.Sprintf(`package repository

type %sRepository interface {
}

type %sRepository struct {
}

func NewRepository() %sRepository {
	return &%sRepository{}
}
`, svcUpper, svcUpper, svcUpper, svcUpper)
}

func msModelTmpl(svc string) string {
	svcUpper := strings.ToUpper(svc[:1]) + svc[1:]
	return fmt.Sprintf(`package model

import "time"

type %s struct {
	ID        int       `+"`"+`json:"id" db:"id"`+"`"+`
	CreatedAt time.Time `+"`"+`json:"created_at" db:"created_at"`+"`"+`
	UpdatedAt time.Time `+"`"+`json:"updated_at" db:"updated_at"`+"`"+`
}
`, svcUpper)
}

const msSvcGoModTmpl = `module %s

go 1.22
`

const msSvcDockerfileTmpl = `FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/service ./cmd

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/service .
EXPOSE 8080
CMD ["./service"]
`

const gwMainTmpl = `package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	userURL, _ := url.Parse("http://localhost:8081")
	orderURL, _ := url.Parse("http://localhost:8082")

	mux := http.NewServeMux()
	mux.Handle("/api/v1/users/", httputil.NewSingleHostReverseProxy(userURL))
	mux.Handle("/api/v1/orders/", httputil.NewSingleHostReverseProxy(orderURL))

	log.Println("API Gateway running on :8000")
	log.Fatal(http.ListenAndServe(":8000", mux))
}
`

const gwProxyTmpl = `package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(target string) *httputil.ReverseProxy {
	url, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(url)
}

type Gateway struct {
	routes map[string]*httputil.ReverseProxy
}

func NewGateway() *Gateway {
	return &Gateway{routes: make(map[string]*httputil.ReverseProxy)}
}

func (g *Gateway) Register(path, target string) {
	g.routes[path] = NewReverseProxy(target)
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy, ok := g.routes[r.URL.Path]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	proxy.ServeHTTP(w, r)
}
`

const msDockerComposeTmpl = `version: "3.9"

services:
  user-service:
    build: ./services/user-service
    ports:
      - "8081:8080"

  order-service:
    build: ./services/order-service
    ports:
      - "8082:8080"

  notification-service:
    build: ./services/notification-service
    ports:
      - "8083:8080"

  gateway:
    build: ./gateway
    ports:
      - "8000:8000"
    depends_on:
      - user-service
      - order-service
`

func k8sDeploymentTmpl(svc string) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: %s-service
  template:
    metadata:
      labels:
        app: %s-service
    spec:
      containers:
        - name: %s-service
          image: %s-service:latest
          ports:
            - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: %s-service
spec:
  selector:
    app: %s-service
  ports:
    - port: 8080
`, svc, svc, svc, svc, svc, svc, svc)
}

const nginxConfTmpl = `upstream user-service {
    server user-service:8080;
}

upstream order-service {
    server order-service:8080;
}

server {
    listen 80;

    location /api/v1/users/ {
        proxy_pass http://user-service;
    }

    location /api/v1/orders/ {
        proxy_pass http://order-service;
    }
}
`

// =========================================================================
// Fullstack templates
// =========================================================================

const fwMainGoTmpl = `package main

import (
	"log"
	"net/http"

	"{{.ModuleName}}/internal/handler"
	"{{.ModuleName}}/internal/template"
)

func main() {
	tmpl := template.NewRenderer()

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	h := handler.NewHomeHandler(tmpl)
	mux.HandleFunc("/", h.Home)
	mux.HandleFunc("/dashboard", h.Dashboard)

	auth := handler.NewAuthHandler(tmpl)
	mux.HandleFunc("/login", auth.Login)
	mux.HandleFunc("/register", auth.Register)

	addr := ":8080"
	log.Printf("Server running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
`

const fwHomeHandlerTmpl = `package handler

import (
	"net/http"

	"{{.ModuleName}}/internal/template"
)

type HomeHandler struct {
	tmpl *template.Renderer
}

func NewHomeHandler(tmpl *template.Renderer) *HomeHandler {
	return &HomeHandler{tmpl: tmpl}
}

func (h *HomeHandler) Home(w http.ResponseWriter, r *http.Request) {
	h.tmpl.Render(w, "home.html", nil)
}

func (h *HomeHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Dashboard",
		"User":  "Guest",
	}
	h.tmpl.Render(w, "home.html", data)
}
`

const fwAuthHandlerTmpl = `package handler

import (
	"net/http"

	"{{.ModuleName}}/internal/template"
)

type AuthHandler struct {
	tmpl *template.Renderer
}

func NewAuthHandler(tmpl *template.Renderer) *AuthHandler {
	return &AuthHandler{tmpl: tmpl}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	h.tmpl.Render(w, "home.html", map[string]interface{}{
		"Title": "Login",
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	h.tmpl.Render(w, "home.html", map[string]interface{}{
		"Title": "Register",
	})
}
`

const fwRendererTmpl = `package template

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	templates *template.Template
}

func NewRenderer() *Renderer {
	tmpl := template.Must(template.ParseGlob("web/templates/**/*.html"))
	return &Renderer{templates: tmpl}
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data interface{}) {
	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
`

func fwCSSTmpl(framework string) string {
	switch framework {
	case "tailwind":
		return `@tailwind base;
@tailwind components;
@tailwind utilities;

body {
	@apply bg-gray-50 text-gray-900 antialiased;
}
`
	case "bootstrap":
		return `/* Bootstrap is loaded via CDN in base.html */
body {
	font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
}
`
	default:
		return `* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}

body {
	font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	line-height: 1.6;
	color: #333;
	background: #fafafa;
}

.container {
	max-width: 1200px;
	margin: 0 auto;
	padding: 0 1rem;
}
`
	}
}

const fwJSTmpl = `// App JavaScript
document.addEventListener("DOMContentLoaded", function () {
	console.log("{{.ProjectName}} app loaded");
});
`

const fwBaseTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{"{{"}}if .Title}}{{"{{"}}.Title}} - {{"{{"}}end}}{{"{{"}}.ProjectName}}</title>
	<link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
	{{"{{"}}template "navbar.html" .}}

	<main class="container">
		{{"{{"}}block "content" .}}{{"{{"}}end}}
	</main>

	<script src="/static/js/app.js"></script>
</body>
</html>
`

const fwHomePageTmpl = `{{"{{"}}define "content"}}
<div class="hero">
	<h1>Welcome to {{"{{"}}.ProjectName}}</h1>
	<p>A full-stack web application built with Go.</p>
	<a href="/dashboard">Get Started</a>
</div>
{{"{{"}}end}}
`

const fwNavbarTmpl = `<nav>
	<div class="container">
		<a href="/">{{"{{"}}.ProjectName}}</a>
		<ul>
			<li><a href="/">Home</a></li>
			<li><a href="/dashboard">Dashboard</a></li>
			<li><a href="/login">Login</a></li>
		</ul>
	</div>
</nav>
`

const fwEnvTmpl = `APP_NAME={{.ProjectName}}
APP_PORT=8080
`

const fwMakefileTmpl = `APP_NAME={{.ProjectName}}
BIN_DIR=./bin

.PHONY: run build tidy lint test clean

run:
	go run ./cmd/web

build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/web

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

test:
	go test ./... -v

clean:
	rm -rf $(BIN_DIR)
`

const fwReadmeTmpl = `# {{.ProjectName}}

A full-stack web application built with [GoStack CLI](https://github.com/alifkhasan01/gostack-cli).

## Stack

| Layer    | Choice              |
|----------|---------------------|
| Backend  | Go (net/http)       |
| Templ    | html/template       |
| Database | {{.Database}}       |
| CSS      | {{.CSSFramework}}   |

## Getting Started

` + "```bash" + `
cp .env.example .env
# edit .env with your DB credentials (if using database)

go mod tidy
go run ./cmd/web
` + "```" + `

## Structure

` + "```" + `
.
├── cmd/web/          # entry point
├── internal/
│   ├── handler/      # HTTP handlers
│   ├── service/      # business logic
│   ├── repository/   # data access
│   └── template/     # template rendering
├── web/
│   ├── static/       # CSS, JS, images
│   └── templates/    # HTML templates
├── migrations/       # SQL migrations
└── go.mod
` + "```" + `
`

const fwDockerfileTmpl = `# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/server ./cmd/web

# Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
COPY web ./web
COPY .env.example .env
EXPOSE 8080
CMD ["./server"]
`
