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
		"internal/middleware/cors.go":       corsGoTmpl(cfg.Framework),
		"internal/middleware/requestid.go":  requestIDGoTmpl(cfg.Framework),
		"internal/middleware/logger.go":     loggerGoTmpl(cfg.Framework),
		"internal/middleware/recovery.go":   recoveryGoTmpl(cfg.Framework),
		"internal/middleware/timeout.go":    timeoutGoTmpl(cfg.Framework),
		"internal/middleware/ratelimit.go":  rateLimitGoTmpl(cfg.Framework),

		// --- entity example ---
		"internal/entity/user.go": entityGoTmpl,

		// --- handler example ---
		"internal/handler/user_handler.go": handlerGoTmpl(cfg.Framework),

		// --- service example ---
		"internal/service/user_service.go": serviceGoTmpl,

		// --- repository example ---
		"internal/repository/user_repository.go": repositoryGoTmpl,

		// --- example tests ---
		"internal/handler/user_handler_test.go": handlerTestGoTmpl,
		"internal/service/user_service_test.go": serviceTestGoTmpl,

		// --- utility packages ---
		"internal/logger/logger.go":         loggerGoTmplFile,
		"internal/response/response.go":     responseGoTmplFile,
		"internal/validator/validator.go":   validatorGoTmplFile,
		"internal/errors/errors.go":         errorsGoTmplFile,

		// --- migrations ---
		"migrations/.gitkeep": "",

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

	// CI
	files[".github/workflows/ci.yml"] = ciGoTmpl

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
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())

	api := app.Group("/api/v1")
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "app": cfg.AppName})
	})
	api.Get("/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "alive"})
	})
	api.Get("/ready", func(c *fiber.Ctx) error {
		if err := db.Ping(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not ready"})
		}
		return c.JSON(fiber.Map{"status": "ready"})
	})

	// gostack:routes

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
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recovery())
	e.Use(middleware.CORS())

	api := e.Group("/api/v1")
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "app": cfg.AppName})
	})
	api.GET("/live", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "alive"})
	})
	api.GET("/ready", func(c echo.Context) error {
		if err := db.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	// gostack:routes

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
)

// Setup registers all application routes.
func Setup(cfg *config.Config, db *sql.DB) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS())

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok", "app": cfg.AppName})
		})
		r.Get("/live", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
		})
		r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := db.Ping(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{"status": "not ready"})
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
		})
	})

	// gostack:routes

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
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "app": cfg.AppName})
	})
	api.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})
	api.GET("/ready", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
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

// ---------------------------------------------------------------------------
// Middleware templates
// ---------------------------------------------------------------------------

func requestIDGoTmpl(framework string) string {
	switch framework {
	case "fiber":
		return `package middleware

import "github.com/gofiber/fiber/v2"

func RequestID() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = c.IP() + "-" + c.IP()
		}
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}
`
	case "echo":
		return `package middleware

import "github.com/labstack/echo/v4"

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := c.Request().Header.Get("X-Request-ID")
			if id == "" {
				id = c.RealIP() + "-" + c.Response().Header().Get(echo.HeaderXRequestID)
			}
			c.Response().Header().Set("X-Request-ID", id)
			return next(c)
		}
	}
}
`
	case "chi":
		return `package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 16)
			rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}
`
	default: // gin
		return `package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
`
	}
}

func loggerGoTmpl(framework string) string {
	switch framework {
	case "fiber":
		return `package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Logger() func(*fiber.Ctx) error {
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

func recoveryGoTmpl(framework string) string {
	switch framework {
	case "fiber":
		return `package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func Recovery() func(*fiber.Ctx) error {
	return recover.New()
}
`
	case "echo":
		return `package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func Recovery() echo.MiddlewareFunc {
	return echomw.Recover()
}
`
	case "chi":
		return `package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v\n%s", err, debug.Stack())
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
`
	default: // gin
		return `package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n%s", err, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
`
	}
}

func timeoutGoTmpl(framework string) string {
	switch framework {
	case "fiber":
		return `package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
)

func Timeout(duration time.Duration) fiber.Handler {
	return timeout.New(func(c *fiber.Ctx) error {
		return c.Next()
	}, duration)
}
`
	case "echo":
		return `package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func Timeout(duration time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), duration)
			defer cancel()
			c.SetRequest(c.Request().WithContext(ctx))

			done := make(chan error, 1)
			go func() {
				done <- next(c)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return c.JSON(http.StatusRequestTimeout, map[string]string{"error": "request timeout"})
			}
		}
	}
}
`
	case "chi":
		return `package middleware

import (
	"context"
	"net/http"
	"time"
)

func Timeout(duration time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()
			r = r.WithContext(ctx)

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				http.Error(w, ` + "`" + `{"error":"request timeout"}` + "`" + `, http.StatusRequestTimeout)
			}
		})
	}
}
`
	default: // gin
		return `package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"error": "request timeout",
			})
		}
	}
}
`
	}
}

func rateLimitGoTmpl(framework string) string {
	switch framework {
	case "fiber":
		return `package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RateLimit(limit int, window time.Duration) func(*fiber.Ctx) error {
	return limiter.New(limiter.Config{
		Max:        limit,
		Expiration: window,
	})
}
`
	case "echo":
		return `package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func RateLimit(limit int, window time.Duration) echo.MiddlewareFunc {
	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Rate:       limit,
		Burst:      limit,
		ExpiresIn:  window,
	})
}
`
	case "chi":
		return `package middleware

import (
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[ip]
	if !ok {
		rl.visitors[ip] = &visitor{lastSeen: time.Now(), count: 1}
		return true
	}

	if time.Since(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = time.Now()
		return true
	}

	v.lastSeen = time.Now()
	v.count++
	return v.count <= rl.limit
}

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := NewRateLimiter(limit, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow(r.RemoteAddr) {
				http.Error(w, ` + "`" + `{"error":"too many requests"}` + "`" + `, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
`
	default: // gin
		return `package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[ip]
	if !ok {
		rl.visitors[ip] = &visitor{lastSeen: time.Now(), count: 1}
		return true
	}

	if time.Since(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = time.Now()
		return true
	}

	v.lastSeen = time.Now()
	v.count++
	return v.count <= rl.limit
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	rl := NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}
		c.Next()
	}
}
`
	}
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

func handlerGoTmpl(framework string) string {
	_ = framework
	return `package handler

import (
	"net/http"

	"{{.ModuleName}}/internal/service"
)

// UserHandler handles HTTP requests for user resources.
type UserHandler struct {
	svc service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// List handles GET /users.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(` + "`" + `{"message":"list users"}` + "`" + `))
}
`
}

const serviceGoTmpl = `package service

// UserService defines business logic for user.
type UserService interface {
}

// userService implements UserService.
type userService struct {
}

// NewUserService creates a new UserService.
func NewUserService() UserService {
	return &userService{}
}
`

const repositoryGoTmpl = `package repository

// UserRepository defines data-access operations for user.
type UserRepository interface {
}

// userRepository implements UserRepository.
type userRepository struct {
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository() UserRepository {
	return &userRepository{}
}
`

// ---------------------------------------------------------------------------
// Example test templates
// ---------------------------------------------------------------------------

const handlerTestGoTmpl = `package handler

import (
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
}
`

const serviceTestGoTmpl = `package service

import "testing"

func TestNewUserService(t *testing.T) {
	svc := NewUserService()
	if svc == nil {
		t.Error("expected non-nil service")
	}
}
`

// ---------------------------------------------------------------------------
// Utility package templates
// ---------------------------------------------------------------------------

const loggerGoTmplFile = `package logger

import (
	"context"
	"log/slog"
	"os"
)

var log *slog.Logger

func Init() {
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func InitWithLevel(level slog.Level) {
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

func Info(msg string, args ...any) {
	log.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	log.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	log.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	log.Debug(msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	log.InfoContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	log.ErrorContext(ctx, msg, args...)
}
`

const responseGoTmplFile = `package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Status  string      ` + "`" + `json:"status"` + "`" + `
	Message string      ` + "`" + `json:"message,omitempty"` + "`" + `
	Data    interface{} ` + "`" + `json:"data,omitempty"` + "`" + `
	Meta    *Meta       ` + "`" + `json:"meta,omitempty"` + "`" + `
	Errors  interface{} ` + "`" + `json:"errors,omitempty"` + "`" + `
}

type Meta struct {
	Page       int ` + "`" + `json:"page"` + "`" + `
	PerPage    int ` + "`" + `json:"per_page"` + "`" + `
	Total      int ` + "`" + `json:"total"` + "`" + `
	TotalPages int ` + "`" + `json:"total_pages"` + "`" + `
}

func writeJSON(w http.ResponseWriter, status int, data APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data:   data,
	})
}

func Created(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusCreated, APIResponse{
		Status: "success",
		Data:   data,
	})
}

func Error(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse{
		Status:  "error",
		Message: message,
	})
}

func ValidationError(w http.ResponseWriter, errors interface{}) {
	writeJSON(w, http.StatusUnprocessableEntity, APIResponse{
		Status: "error",
		Errors: errors,
	})
}

func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "resource not found"
	}
	Error(w, http.StatusNotFound, message)
}

func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal server error")
}

func Paginated(w http.ResponseWriter, data interface{}, page, perPage, total int) {
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}
	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data:   data,
		Meta: &Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
`

const validatorGoTmplFile = `package validator

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
	Field   string ` + "`" + `json:"field"` + "`" + `
	Message string ` + "`" + `json:"message"` + "`" + `
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
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", param)
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", param)
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

func FormatErrors(errs []ValidationError) map[string]string {
	formatted := make(map[string]string)
	for _, e := range errs {
		formatted[e.Field] = e.Message
	}
	return formatted
}

func FormatErrorsSlice(errs []ValidationError) []string {
	var msgs []string
	for _, e := range errs {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return msgs
}

func Valid(i interface{}) bool {
	return validate.Struct(i) == nil
}

func HasErrors(errs []ValidationError) bool {
	return len(errs) > 0
}

func ErrorSlice(errs []ValidationError) string {
	var sb strings.Builder
	for i, e := range errs {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return sb.String()
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

const errorsGoTmplFile = `package errors

import "net/http"

type AppError struct {
	Code        int    ` + "`" + `json:"-"` + "`" + `
	Message     string ` + "`" + `json:"message"` + "`" + `
	InternalErr error  ` + "`" + `json:"-"` + "`" + `
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.InternalErr
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, InternalErr: err}
}

var (
	ErrNotFound       = New(http.StatusNotFound, "resource not found")
	ErrBadRequest     = New(http.StatusBadRequest, "bad request")
	ErrUnauthorized   = New(http.StatusUnauthorized, "unauthorized")
	ErrForbidden      = New(http.StatusForbidden, "forbidden")
	ErrConflict       = New(http.StatusConflict, "resource already exists")
	ErrInternal       = New(http.StatusInternalServerError, "internal server error")
	ErrUnprocessable  = New(http.StatusUnprocessableEntity, "unprocessable entity")
	ErrTooManyRequest = New(http.StatusTooManyRequests, "too many requests")
)

func IsNotFound(err error) bool  { return isCode(err, http.StatusNotFound) }
func IsBadRequest(err error) bool { return isCode(err, http.StatusBadRequest) }
func IsConflict(err error) bool   { return isCode(err, http.StatusConflict) }

func isCode(err error, code int) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
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

.PHONY: run build tidy lint test clean migrate

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

clean:
	rm -rf $(BIN_DIR)

migrate:
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
│   ├── errors/       # custom error types
│   ├── handler/      # HTTP handlers
│   ├── logger/       # structured logger
│   ├── middleware/   # HTTP middlewares
│   ├── repository/   # data access layer
│   ├── response/     # standard API response
│   ├── routes/       # route registration
│   ├── service/      # business logic
│   └── validator/    # request validation
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
