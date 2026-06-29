package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Kind represents the type of file to generate.
type Kind string

const (
	KindHandler    Kind = "handler"
	KindService    Kind = "service"
	KindRepository Kind = "repository"
	KindMigration  Kind = "migration"
	KindModule     Kind = "module"
	KindCRUD       Kind = "crud"
)

// Generate creates a new file of the given kind with the given name.
// It detects the project root by looking for go.mod.
func Generate(kind Kind, name string) error {
	// CRUD is handled separately — it needs project.Meta
	if kind == KindCRUD {
		return GenerateCRUD(name)
	}

	root, err := findProjectRoot()
	if err != nil {
		return err
	}

	switch kind {
	case KindHandler:
		return generateHandler(root, name)
	case KindService:
		return generateService(root, name)
	case KindRepository:
		return generateRepository(root, name)
	case KindMigration:
		return generateMigration(root, name)
	case KindModule:
		return generateModule(root, name)
	default:
		return fmt.Errorf("unknown kind: %s", kind)
	}
}

// -------------------------------------------------------------------------
// Handler
// -------------------------------------------------------------------------

var handlerTmpl = `package handler

import "net/http"

// {{.Name}}Handler handles HTTP requests for {{.Lower}} resources.
type {{.Name}}Handler struct {
	// TODO: inject service
}

// New{{.Name}}Handler creates a new {{.Name}}Handler.
func New{{.Name}}Handler() *{{.Name}}Handler {
	return &{{.Name}}Handler{}
}

// List handles GET /{{.Lower}}s
func (h *{{.Name}}Handler) List(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.WriteHeader(http.StatusOK)
}

// Get handles GET /{{.Lower}}s/:id
func (h *{{.Name}}Handler) Get(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.WriteHeader(http.StatusOK)
}

// Create handles POST /{{.Lower}}s
func (h *{{.Name}}Handler) Create(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.WriteHeader(http.StatusCreated)
}

// Update handles PUT /{{.Lower}}s/:id
func (h *{{.Name}}Handler) Update(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.WriteHeader(http.StatusOK)
}

// Delete handles DELETE /{{.Lower}}s/:id
func (h *{{.Name}}Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.WriteHeader(http.StatusNoContent)
}
`

func generateHandler(root, name string) error {
	dir := filepath.Join(root, "internal", "handler")
	return writeTemplate(handlerTmpl, dir, strings.ToLower(name)+"_handler.go", name)
}

// -------------------------------------------------------------------------
// Service
// -------------------------------------------------------------------------

var serviceTmpl = `package service

// {{.Name}}Service defines business logic for {{.Lower}}.
type {{.Name}}Service interface {
	GetAll() ([]interface{}, error)
	GetByID(id int) (interface{}, error)
	Create(data interface{}) error
	Update(id int, data interface{}) error
	Delete(id int) error
}

type {{.Lower}}Service struct {
	// TODO: inject repository
}

// New{{.Name}}Service creates a new {{.Name}}Service.
func New{{.Name}}Service() {{.Name}}Service {
	return &{{.Lower}}Service{}
}

func (s *{{.Lower}}Service) GetAll() ([]interface{}, error) {
	// TODO: implement
	return nil, nil
}

func (s *{{.Lower}}Service) GetByID(id int) (interface{}, error) {
	// TODO: implement
	return nil, nil
}

func (s *{{.Lower}}Service) Create(data interface{}) error {
	// TODO: implement
	return nil
}

func (s *{{.Lower}}Service) Update(id int, data interface{}) error {
	// TODO: implement
	return nil
}

func (s *{{.Lower}}Service) Delete(id int) error {
	// TODO: implement
	return nil
}
`

func generateService(root, name string) error {
	dir := filepath.Join(root, "internal", "service")
	return writeTemplate(serviceTmpl, dir, strings.ToLower(name)+"_service.go", name)
}

// -------------------------------------------------------------------------
// Repository
// -------------------------------------------------------------------------

var repositoryTmpl = `package repository

// {{.Name}}Repository defines data-access operations for {{.Lower}}.
type {{.Name}}Repository interface {
	FindAll() ([]interface{}, error)
	FindByID(id int) (interface{}, error)
	Save(data interface{}) error
	Update(id int, data interface{}) error
	Delete(id int) error
}

type {{.Lower}}Repository struct {
	// TODO: inject DB
}

// New{{.Name}}Repository creates a new {{.Name}}Repository.
func New{{.Name}}Repository() {{.Name}}Repository {
	return &{{.Lower}}Repository{}
}

func (r *{{.Lower}}Repository) FindAll() ([]interface{}, error) {
	return nil, nil
}

func (r *{{.Lower}}Repository) FindByID(id int) (interface{}, error) {
	return nil, nil
}

func (r *{{.Lower}}Repository) Save(data interface{}) error {
	return nil
}

func (r *{{.Lower}}Repository) Update(id int, data interface{}) error {
	return nil
}

func (r *{{.Lower}}Repository) Delete(id int) error {
	return nil
}
`

func generateRepository(root, name string) error {
	dir := filepath.Join(root, "internal", "repository")
	return writeTemplate(repositoryTmpl, dir, strings.ToLower(name)+"_repository.go", name)
}

// -------------------------------------------------------------------------
// Migration
// -------------------------------------------------------------------------

var migrationTmpl = `-- Migration: {{.Name}}
-- Created by GoStack CLI

-- +migrate Up
CREATE TABLE IF NOT EXISTS {{.Lower}}s (
    id         SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS {{.Lower}}s;
`

func generateMigration(root, name string) error {
	dir := filepath.Join(root, "migrations")
	// Use timestamp prefix for ordering
	fileName := nextMigrationFileName(dir, name)
	return writeTemplate(migrationTmpl, dir, fileName, name)
}

func nextMigrationFileName(dir, name string) string {
	entries, _ := os.ReadDir(dir)
	seq := len(entries) + 1
	return fmt.Sprintf("%04d_%s.sql", seq, strings.ToLower(name))
}

// -------------------------------------------------------------------------
// Module (handler + service + repository in one shot)
// -------------------------------------------------------------------------

func generateModule(root, name string) error {
	fmt.Printf("  Generating module: %s\n", name)
	if err := generateHandler(root, name); err != nil {
		return err
	}
	if err := generateService(root, name); err != nil {
		return err
	}
	if err := generateRepository(root, name); err != nil {
		return err
	}
	fmt.Printf("  ✅ Module '%s' generated (handler + service + repository)\n", name)
	return nil
}

// -------------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------------

type tmplData struct {
	Name  string
	Lower string
}

func writeTemplate(tmplStr, dir, fileName, name string) error {
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

	data := tmplData{
		Name:  capitalize(name),
		Lower: strings.ToLower(name),
	}

	if err := t.Execute(f, data); err != nil {
		return err
	}

	fmt.Printf("  ✅ Created: %s\n", destPath)
	return nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// findProjectRoot walks up from cwd looking for go.mod.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found — run this command inside a Go project")
}
