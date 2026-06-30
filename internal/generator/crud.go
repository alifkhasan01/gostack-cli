package generator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alifkhasan01/gostack-cli/internal/project"
)

// CRUDConfig holds everything needed to generate a full CRUD module.
type CRUDConfig struct {
	Name   string // e.g. "Product"
	Lower  string // e.g. "product"
	Plural string // e.g. "products"
	Meta   *project.Meta
	Root   string
}

// GenerateCRUD creates entity, handler, service, repository, migration,
// and injects routes — all in one shot.
func GenerateCRUD(name string) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}

	meta, err := project.ReadFromDir(root)
	if err != nil {
		// Fallback: minimal meta for projects not created with gostack
		meta = &project.Meta{
			Framework:    "gin",
			Architecture: "standard",
			ModuleName:   moduleName(root),
		}
	}

	cfg := CRUDConfig{
		Name:   capitalize(name),
		Lower:  strings.ToLower(name),
		Plural: strings.ToLower(name) + "s",
		Meta:   meta,
		Root:   root,
	}

	fmt.Printf("\n  Generating CRUD: %s\n\n", cfg.Name)

	steps := []struct {
		label string
		fn    func() error
	}{
		{"entity", func() error { return cfg.writeEntity() }},
		{"repository", func() error { return cfg.writeRepository() }},
		{"service", func() error { return cfg.writeService() }},
		{"handler", func() error { return cfg.writeHandler() }},
		{"migration", func() error { return cfg.writeMigration() }},
		{"routes", func() error { return cfg.injectRoutes() }},
	}

	for _, s := range steps {
		if err := s.fn(); err != nil {
			return fmt.Errorf("%s: %w", s.label, err)
		}
	}

	fmt.Printf("\n  ✅ CRUD '%s' generated successfully!\n", cfg.Name)
	fmt.Printf("     Routes registered: /api/v1/%s\n\n", cfg.Plural)
	return nil
}

// -------------------------------------------------------------------------
// Entity
// -------------------------------------------------------------------------

const crudEntityTmpl = `package entity

import "time"

// {{.Name}} represents the {{.Lower}} domain model.
type {{.Name}} struct {
	ID        int       ` + "`json:\"id\" db:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" db:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" db:\"updated_at\"`" + `
}

// Create{{.Name}}Request is the payload for creating a {{.Lower}}.
type Create{{.Name}}Request struct {
	// TODO: add fields
}

// Update{{.Name}}Request is the payload for updating a {{.Lower}}.
type Update{{.Name}}Request struct {
	// TODO: add fields
}
`

func (c *CRUDConfig) writeEntity() error {
	dir := filepath.Join(c.Root, "internal", "entity")
	return writeCRUDTemplate(crudEntityTmpl, dir, c.Lower+"_entity.go", c)
}

// -------------------------------------------------------------------------
// Repository
// -------------------------------------------------------------------------

const crudRepoTmpl = `package repository

import (
	"database/sql"
	"fmt"

	"{{.Meta.ModuleName}}/internal/entity"
)

// {{.Name}}Repository defines data-access for {{.Lower}}.
type {{.Name}}Repository interface {
	FindAll() ([]entity.{{.Name}}, error)
	FindByID(id int) (*entity.{{.Name}}, error)
	Save(req entity.Create{{.Name}}Request) (*entity.{{.Name}}, error)
	Update(id int, req entity.Update{{.Name}}Request) (*entity.{{.Name}}, error)
	Delete(id int) error
}

type {{.Lower}}Repository struct {
	db *sql.DB
}

// New{{.Name}}Repository creates a new {{.Name}}Repository backed by sql.DB.
func New{{.Name}}Repository(db *sql.DB) {{.Name}}Repository {
	return &{{.Lower}}Repository{db: db}
}

func (r *{{.Lower}}Repository) FindAll() ([]entity.{{.Name}}, error) {
	rows, err := r.db.Query(` + "`" + `SELECT id, created_at, updated_at FROM {{.Plural}}` + "`" + `)
	if err != nil {
		return nil, fmt.Errorf("{{.Lower}}.FindAll: %w", err)
	}
	defer rows.Close()

	var items []entity.{{.Name}}
	for rows.Next() {
		var item entity.{{.Name}}
		if err := rows.Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *{{.Lower}}Repository) FindByID(id int) (*entity.{{.Name}}, error) {
	var item entity.{{.Name}}
	err := r.db.QueryRow(` + "`" + `SELECT id, created_at, updated_at FROM {{.Plural}} WHERE id = $1` + "`" + `, id).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("{{.Lower}}.FindByID: %w", err)
	}
	return &item, nil
}

func (r *{{.Lower}}Repository) Save(_ entity.Create{{.Name}}Request) (*entity.{{.Name}}, error) {
	var item entity.{{.Name}}
	err := r.db.QueryRow(` + "`" + `INSERT INTO {{.Plural}} DEFAULT VALUES RETURNING id, created_at, updated_at` + "`" + `).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("{{.Lower}}.Save: %w", err)
	}
	return &item, nil
}

func (r *{{.Lower}}Repository) Update(id int, _ entity.Update{{.Name}}Request) (*entity.{{.Name}}, error) {
	var item entity.{{.Name}}
	err := r.db.QueryRow(` + "`" + `UPDATE {{.Plural}} SET updated_at = NOW() WHERE id = $1 RETURNING id, created_at, updated_at` + "`" + `, id).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("{{.Lower}}.Update: %w", err)
	}
	return &item, nil
}

func (r *{{.Lower}}Repository) Delete(id int) error {
	_, err := r.db.Exec(` + "`" + `DELETE FROM {{.Plural}} WHERE id = $1` + "`" + `, id)
	return err
}
`

func (c *CRUDConfig) writeRepository() error {
	dir := filepath.Join(c.Root, "internal", "repository")
	return writeCRUDTemplate(crudRepoTmpl, dir, c.Lower+"_repository.go", c)
}

// -------------------------------------------------------------------------
// Service
// -------------------------------------------------------------------------

const crudServiceTmpl = `package service

import (
	"{{.Meta.ModuleName}}/internal/entity"
	"{{.Meta.ModuleName}}/internal/repository"
)

// {{.Name}}Service defines business logic for {{.Lower}}.
type {{.Name}}Service interface {
	GetAll() ([]entity.{{.Name}}, error)
	GetByID(id int) (*entity.{{.Name}}, error)
	Create(req entity.Create{{.Name}}Request) (*entity.{{.Name}}, error)
	Update(id int, req entity.Update{{.Name}}Request) (*entity.{{.Name}}, error)
	Delete(id int) error
}

type {{.Lower}}Service struct {
	repo repository.{{.Name}}Repository
}

// New{{.Name}}Service creates a new {{.Name}}Service.
func New{{.Name}}Service(repo repository.{{.Name}}Repository) {{.Name}}Service {
	return &{{.Lower}}Service{repo: repo}
}

func (s *{{.Lower}}Service) GetAll() ([]entity.{{.Name}}, error) {
	return s.repo.FindAll()
}

func (s *{{.Lower}}Service) GetByID(id int) (*entity.{{.Name}}, error) {
	return s.repo.FindByID(id)
}

func (s *{{.Lower}}Service) Create(req entity.Create{{.Name}}Request) (*entity.{{.Name}}, error) {
	return s.repo.Save(req)
}

func (s *{{.Lower}}Service) Update(id int, req entity.Update{{.Name}}Request) (*entity.{{.Name}}, error) {
	return s.repo.Update(id, req)
}

func (s *{{.Lower}}Service) Delete(id int) error {
	return s.repo.Delete(id)
}
`

func (c *CRUDConfig) writeService() error {
	dir := filepath.Join(c.Root, "internal", "service")
	return writeCRUDTemplate(crudServiceTmpl, dir, c.Lower+"_service.go", c)
}

// -------------------------------------------------------------------------
// Handler (Gin — default; fallback for unknown frameworks)
// -------------------------------------------------------------------------

const crudHandlerGinTmpl = `package handler

import (
	"net/http"
	"strconv"

	"{{.Meta.ModuleName}}/internal/entity"
	"{{.Meta.ModuleName}}/internal/service"
	"github.com/gin-gonic/gin"
)

// {{.Name}}Handler handles HTTP requests for {{.Lower}} resources.
type {{.Name}}Handler struct {
	svc service.{{.Name}}Service
}

// New{{.Name}}Handler creates a new {{.Name}}Handler.
func New{{.Name}}Handler(svc service.{{.Name}}Service) *{{.Name}}Handler {
	return &{{.Name}}Handler{svc: svc}
}

// List godoc
// @Summary  List {{.Plural}}
// @Tags     {{.Plural}}
// @Produce  json
// @Success  200 {array} entity.{{.Name}}
// @Router   /{{.Plural}} [get]
func (h *{{.Name}}Handler) List(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Get godoc
// @Summary  Get {{.Lower}} by ID
// @Tags     {{.Plural}}
// @Produce  json
// @Param    id path int true "{{.Name}} ID"
// @Success  200 {object} entity.{{.Name}}
// @Router   /{{.Plural}}/{id} [get]
func (h *{{.Name}}Handler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	item, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Create godoc
// @Summary  Create {{.Lower}}
// @Tags     {{.Plural}}
// @Accept   json
// @Produce  json
// @Param    body body entity.Create{{.Name}}Request true "Payload"
// @Success  201 {object} entity.{{.Name}}
// @Router   /{{.Plural}} [post]
func (h *{{.Name}}Handler) Create(c *gin.Context) {
	var req entity.Create{{.Name}}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.svc.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// Update godoc
// @Summary  Update {{.Lower}}
// @Tags     {{.Plural}}
// @Accept   json
// @Produce  json
// @Param    id   path int                         true "{{.Name}} ID"
// @Param    body body entity.Update{{.Name}}Request true "Payload"
// @Success  200 {object} entity.{{.Name}}
// @Router   /{{.Plural}}/{id} [put]
func (h *{{.Name}}Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req entity.Update{{.Name}}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.svc.Update(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Delete godoc
// @Summary  Delete {{.Lower}}
// @Tags     {{.Plural}}
// @Param    id path int true "{{.Name}} ID"
// @Success  204
// @Router   /{{.Plural}}/{id} [delete]
func (h *{{.Name}}Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
`

func crudHandlerTmpl(framework string) string {
	switch framework {
	case "fiber":
		return crudHandlerFiberTmpl
	case "echo":
		return crudHandlerEchoTmpl
	case "chi":
		return crudHandlerChiTmpl
	default:
		return crudHandlerGinTmpl
	}
}

const crudHandlerFiberTmpl = `package handler

import (
	"strconv"

	"{{.Meta.ModuleName}}/internal/entity"
	"{{.Meta.ModuleName}}/internal/service"
	"github.com/gofiber/fiber/v2"
)

// {{.Name}}Handler handles HTTP requests for {{.Lower}} resources.
type {{.Name}}Handler struct {
	svc service.{{.Name}}Service
}

// New{{.Name}}Handler creates a new {{.Name}}Handler.
func New{{.Name}}Handler(svc service.{{.Name}}Service) *{{.Name}}Handler {
	return &{{.Name}}Handler{svc: svc}
}

// List handles GET /{{.Plural}}
func (h *{{.Name}}Handler) List(c *fiber.Ctx) error {
	items, err := h.svc.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

// Get handles GET /{{.Plural}}/:id
func (h *{{.Name}}Handler) Get(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	item, err := h.svc.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

// Create handles POST /{{.Plural}}
func (h *{{.Name}}Handler) Create(c *fiber.Ctx) error {
	var req entity.Create{{.Name}}Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	item, err := h.svc.Create(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// Update handles PUT /{{.Plural}}/:id
func (h *{{.Name}}Handler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req entity.Update{{.Name}}Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	item, err := h.svc.Update(id, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

// Delete handles DELETE /{{.Plural}}/:id
func (h *{{.Name}}Handler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.svc.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
`

func (c *CRUDConfig) writeHandler() error {
	dir := filepath.Join(c.Root, "internal", "handler")
	tmpl := crudHandlerTmpl(c.Meta.Framework)
	return writeCRUDTemplate(tmpl, dir, c.Lower+"_handler.go", c)
}

// -------------------------------------------------------------------------
// Migration
// -------------------------------------------------------------------------

const crudMigrationTmpl = `-- Migration: create_{{.Plural}}
-- Generated by GoStack CLI

-- +migrate Up
CREATE TABLE IF NOT EXISTS {{.Plural}} (
    id         SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS {{.Plural}};
`

func (c *CRUDConfig) writeMigration() error {
	dir := filepath.Join(c.Root, "migrations")
	fileName := nextMigrationFileName(dir, "create_"+c.Plural)
	return writeCRUDTemplate(crudMigrationTmpl, dir, fileName, c)
}

const crudHandlerEchoTmpl = `package handler

import (
	"net/http"
	"strconv"

	"{{.Meta.ModuleName}}/internal/entity"
	"{{.Meta.ModuleName}}/internal/service"
	"github.com/labstack/echo/v4"
)

type {{.Name}}Handler struct {
	svc service.{{.Name}}Service
}

func New{{.Name}}Handler(svc service.{{.Name}}Service) *{{.Name}}Handler {
	return &{{.Name}}Handler{svc: svc}
}

func (h *{{.Name}}Handler) List(c echo.Context) error {
	items, err := h.svc.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *{{.Name}}Handler) Get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	item, err := h.svc.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, item)
}

func (h *{{.Name}}Handler) Create(c echo.Context) error {
	var req entity.Create{{.Name}}Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	item, err := h.svc.Create(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *{{.Name}}Handler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	var req entity.Update{{.Name}}Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	item, err := h.svc.Update(id, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, item)
}

func (h *{{.Name}}Handler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	if err := h.svc.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
`

const crudHandlerChiTmpl = `package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"{{.Meta.ModuleName}}/internal/entity"
	"{{.Meta.ModuleName}}/internal/service"
)

type {{.Name}}Handler struct {
	svc service.{{.Name}}Service
}

func New{{.Name}}Handler(svc service.{{.Name}}Service) *{{.Name}}Handler {
	return &{{.Name}}Handler{svc: svc}
}

func (h *{{.Name}}Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.GetAll()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *{{.Name}}Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		return
	}
	item, err := h.svc.GetByID(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (h *{{.Name}}Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req entity.Create{{.Name}}Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	item, err := h.svc.Create(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *{{.Name}}Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		return
	}
	var req entity.Update{{.Name}}Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	item, err := h.svc.Update(id, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (h *{{.Name}}Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		return
	}
	if err := h.svc.Delete(id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
`

// -------------------------------------------------------------------------
// Route injection
// -------------------------------------------------------------------------

// injectRoutes appends the new resource routes into internal/routes/routes.go.
// It looks for a specific injection comment anchor:
//
//	// gostack:routes
func (c *CRUDConfig) injectRoutes() error {
	routesFile := filepath.Join(c.Root, "internal", "routes", "routes.go")
	if _, err := os.Stat(routesFile); os.IsNotExist(err) {
		fmt.Printf("  ⚠️  routes.go not found, skipping route injection\n")
		return nil
	}

	if err := c.injectImports(routesFile); err != nil {
		return fmt.Errorf("inject imports: %w", err)
	}

	content, err := os.ReadFile(routesFile)
	if err != nil {
		return err
	}

	src := string(content)
	anchor := "// gostack:routes"

	if !strings.Contains(src, anchor) {
		fmt.Printf("  ⚠️  anchor '%s' not found in routes.go — add it manually:\n", anchor)
		return nil
	}

	snippet := routeSnippet(c)
	updated := strings.Replace(src, anchor, anchor+"\n\t"+snippet, 1)
	if err := os.WriteFile(routesFile, []byte(updated), 0644); err != nil {
		return err
	}
	fmt.Printf("  ✅ Injected routes: /api/v1/%s\n", c.Plural)
	return nil
}

// injectImports adds missing handler, service, repository imports to routes.go.
func (c *CRUDConfig) injectImports(routesFile string) error {
	content, err := os.ReadFile(routesFile)
	if err != nil {
		return err
	}

	src := string(content)

	needed := []string{
		fmt.Sprintf("%s/internal/handler", c.Meta.ModuleName),
		fmt.Sprintf("%s/internal/repository", c.Meta.ModuleName),
		fmt.Sprintf("%s/internal/service", c.Meta.ModuleName),
	}

	var toAdd []string
	for _, imp := range needed {
		if !strings.Contains(src, imp) {
			toAdd = append(toAdd, imp)
		}
	}

	if len(toAdd) == 0 {
		return nil
	}

	// Find the import block's closing paren
	importStart := strings.Index(src, "import (")
	if importStart == -1 {
		return nil
	}

	rest := src[importStart:]
	closeIdx := strings.Index(rest, "\n)")
	if closeIdx == -1 {
		return nil
	}

	insertAt := importStart + closeIdx
	var sb strings.Builder
	sb.WriteString(src[:insertAt])
	for _, imp := range toAdd {
		sb.WriteString(fmt.Sprintf("\n\t\"%s\"", imp))
	}
	sb.WriteString(src[insertAt:])

	return os.WriteFile(routesFile, []byte(sb.String()), 0644)
}

func routeSnippet(c *CRUDConfig) string {
	switch c.Meta.Framework {
	case "fiber":
		return fiberRouteSnippet(c)
	case "echo":
		return echoRouteSnippet(c)
	case "chi":
		return chiRouteSnippet(c)
	default:
		return ginRouteSnippet(c)
	}
}

func ginRouteSnippet(c *CRUDConfig) string {
	return fmt.Sprintf(
		`// %s routes
	{
		h := handler.New%sHandler(service.New%sService(repository.New%sRepository(db)))
		%s := api.Group("/%s")
		%s.GET("", h.List)
		%s.GET("/:id", h.Get)
		%s.POST("", h.Create)
		%s.PUT("/:id", h.Update)
		%s.DELETE("/:id", h.Delete)
	}`,
		c.Name,
		c.Name, c.Name, c.Name,
		c.Lower, c.Plural,
		c.Lower, c.Lower, c.Lower, c.Lower, c.Lower,
	)
}

func fiberRouteSnippet(c *CRUDConfig) string {
	return fmt.Sprintf(
		`// %s routes
	{
		h := handler.New%sHandler(service.New%sService(repository.New%sRepository(db)))
		%s := api.Group("/%s")
		%s.Get("", h.List)
		%s.Get("/:id", h.Get)
		%s.Post("", h.Create)
		%s.Put("/:id", h.Update)
		%s.Delete("/:id", h.Delete)
	}`,
		c.Name,
		c.Name, c.Name, c.Name,
		c.Lower, c.Plural,
		c.Lower, c.Lower, c.Lower, c.Lower, c.Lower,
	)
}

func echoRouteSnippet(c *CRUDConfig) string {
	return fmt.Sprintf(
		`// %s routes
	{
		h := handler.New%sHandler(service.New%sService(repository.New%sRepository(db)))
		%s := api.Group("/%s")
		%s.GET("", h.List)
		%s.GET("/:id", h.Get)
		%s.POST("", h.Create)
		%s.PUT("/:id", h.Update)
		%s.DELETE("/:id", h.Delete)
	}`,
		c.Name,
		c.Name, c.Name, c.Name,
		c.Lower, c.Plural,
		c.Lower, c.Lower, c.Lower, c.Lower, c.Lower,
	)
}

func chiRouteSnippet(c *CRUDConfig) string {
	return fmt.Sprintf(
		`// %s routes
	{
		h := handler.New%sHandler(service.New%sService(repository.New%sRepository(db)))
		r.Route("/api/v1/%s", func(sub chi.Router) {
			sub.Get("/", h.List)
			sub.Get("/{id}", h.Get)
			sub.Post("/", h.Create)
			sub.Put("/{id}", h.Update)
			sub.Delete("/{id}", h.Delete)
		})
	}`,
		c.Name,
		c.Name, c.Name, c.Name,
		c.Plural,
	)
}

// -------------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------------

func writeCRUDTemplate(tmplStr, dir, fileName string, data interface{}) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}

	destPath := filepath.Join(dir, fileName)
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("file already exists: %s", destPath)
	}

	t, err := template.New("").Parse(tmplStr)
	if err != nil {
		return err
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		return err
	}

	fmt.Printf("  ✅ Created: %s\n", destPath)
	return nil
}

// moduleName reads module name from go.mod in root.
func moduleName(root string) string {
	f, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return "myapp"
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module ")
		}
	}
	return "myapp"
}
