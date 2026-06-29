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

// Generate writes a full project skeleton into destDir.
func Generate(destDir string, cfg Config) error {
	files := buildFileMap(cfg)
	for relPath, content := range files {
		fullPath := filepath.Join(destDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		rendered, err := renderTemplate(content, cfg)
		if err != nil {
			return fmt.Errorf("render %s: %w", relPath, err)
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
		Framework:    cfg.Framework,
		Architecture: cfg.Architecture,
		Database:     cfg.Database,
		ORM:          cfg.ORM,
		Auth:         cfg.Auth,
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
	_ = dbDriver(cfg)
	_ = ormImport(cfg)
	fw := frameworkImport(cfg)

	files := map[string]string{
		// --- go.mod ---
		"go.mod": goModTmpl,

		// --- main entry point ---
		"cmd/api/main.go": mainGoTmpl,

		// --- config ---
		"internal/config/config.go": configGoTmpl,

		// --- database ---
		"internal/database/database.go": databaseGoTmpl,

		// --- routes ---
		"internal/routes/routes.go": routesGoTmpl(cfg.Framework),

		// --- middleware ---
		"internal/middleware/cors.go": corsGoTmpl(cfg.Framework),

		// --- entity example ---
		"internal/entity/user.go": entityGoTmpl,

		// --- env ---
		".env":         envTmpl,
		".env.example": envTmpl,

		// --- gitignore ---
		".gitignore": gitignoreTmpl,

		// --- Makefile ---
		"Makefile": makefileTmpl,

		// --- README ---
		"README.md": readmeTmpl,
	}

	// Docker
	if cfg.Docker {
		files["Dockerfile"] = dockerfileTmpl(cfg.Framework)
		files["docker-compose.yml"] = dockerComposeTmpl
	}

	// Swagger
	if cfg.Swagger {
		files["docs/.gitkeep"] = ""
	}

	// Auth
	if cfg.Auth == "jwt" {
		files["internal/middleware/jwt.go"] = jwtMiddlewareTmpl(fw)
	}

	return files
}

// ============================================================
// Template strings
// ============================================================

const goModTmpl = `module {{.ModuleName}}

go 1.22
`

const mainGoTmpl = `package main

import (
	"log"

	"{{.ModuleName}}/internal/config"
	"{{.ModuleName}}/internal/database"
	"{{.ModuleName}}/internal/routes"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	r := routes.Setup(cfg, db)

	log.Printf("🚀 Server running on %s", cfg.AppAddr)
	if err := r.Run(cfg.AppAddr); err != nil {
		log.Fatal(err)
	}
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

const databaseGoTmpl = `package database

import (
	"database/sql"
	"fmt"

	"{{.ModuleName}}/internal/config"
	_ "github.com/lib/pq"           // postgres
	_ "github.com/go-sql-driver/mysql" // mysql
	_ "github.com/mattn/go-sqlite3"    // sqlite
)

// Connect opens a database connection using the config.
func Connect(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}
`

func routesGoTmpl(framework string) string {
	switch framework {
	case "fiber":
		return `package routes

import (
	"database/sql"

	"{{.ModuleName}}/internal/config"
	"{{.ModuleName}}/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// Setup registers all application routes.
func Setup(cfg *config.Config, db *sql.DB) *fiber.App {
	app := fiber.New()
	app.Use(middleware.CORS())

	api := app.Group("/api/v1")
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "app": cfg.AppName})
	})

	return app
}
`
	case "echo":
		return `package routes

import (
	"database/sql"
	"net/http"

	"{{.ModuleName}}/internal/config"
	"{{.ModuleName}}/internal/middleware"
	"github.com/labstack/echo/v4"
)

// Setup registers all application routes.
func Setup(cfg *config.Config, db *sql.DB) *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORS())

	api := e.Group("/api/v1")
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "app": cfg.AppName})
	})

	return e
}
`
	case "chi":
		return `package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"{{.ModuleName}}/internal/config"
	"{{.ModuleName}}/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// Setup registers all application routes.
func Setup(cfg *config.Config, db *sql.DB) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.CORS())

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok", "app": cfg.AppName})
		})
	})

	return r
}
`
	default: // gin
		return `package routes

import (
	"database/sql"
	"net/http"

	"{{.ModuleName}}/internal/config"
	"{{.ModuleName}}/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Setup registers all application routes and returns the Gin engine.
func Setup(cfg *config.Config, db *sql.DB) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "app": cfg.AppName})
	})

	// gostack:routes

	return r
}
`
	}
}

func corsGoTmpl(framework string) string {
	switch framework {
	case "fiber":
		return `package middleware

import "github.com/gofiber/fiber/v2/middleware/cors"

// CORS returns a Fiber CORS middleware.
func CORS() func(*fiber.Ctx) error {
	return cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	})
}
`
	case "echo":
		return `package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// CORS returns an Echo CORS middleware.
func CORS() echo.MiddlewareFunc {
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	})
}
`
	case "chi":
		return `package middleware

import "net/http"

// CORS returns a simple CORS middleware for chi.
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
`
	default: // gin
		return `package middleware

import "github.com/gin-gonic/gin"

// CORS returns a Gin CORS middleware.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
`
	}
}

func jwtMiddlewareTmpl(_ string) string {
	return `package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTSecret should be loaded from config in production.
var JWTSecret = []byte("change-me")

// JWT validates the Authorization: Bearer <token> header.
func JWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		_, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
`
}

const entityGoTmpl = `package entity

import "time"

// User represents a user in the system.
type User struct {
	ID        int       ` + "`" + `json:"id" db:"id"` + "`" + `
	Name      string    ` + "`" + `json:"name" db:"name"` + "`" + `
	Email     string    ` + "`" + `json:"email" db:"email"` + "`" + `
	CreatedAt time.Time ` + "`" + `json:"created_at" db:"created_at"` + "`" + `
	UpdatedAt time.Time ` + "`" + `json:"updated_at" db:"updated_at"` + "`" + `
}
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

.PHONY: run build tidy lint test

run:
	go run ./cmd/api

build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/api

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

test:
	go test ./... -v

migrate-up:
	goose up

migrate-down:
	goose down
`

const readmeTmpl = `# {{.ProjectName}}

Generated with [GoStack CLI](https://github.com/alifkhasan01/gostack-cli).

## Stack

| Layer      | Choice              |
|------------|---------------------|
| Framework  | {{.Framework}}      |
| Database   | {{.Database}}       |
| ORM        | {{.ORM}}            |
| Auth       | {{.Auth}}           |

## Getting Started

` + "```bash" + `
cp .env.example .env
# edit .env with your DB credentials

go mod tidy
go run ./cmd/api
` + "```" + `

## Structure

` + "```" + `
.
├── cmd/api/          # entry point
├── internal/
│   ├── config/       # environment config
│   ├── database/     # db connection
│   ├── entity/       # domain models
│   ├── handler/      # HTTP handlers
│   ├── middleware/   # HTTP middlewares
│   ├── repository/   # data access layer
│   ├── routes/       # route registration
│   └── service/      # business logic
├── migrations/       # SQL migrations
└── docs/             # Swagger docs
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
RUN go build -o /app/server ./cmd/api

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
