// Package project reads and writes the gostack.json file that lives at the
// root of every generated project. It lets generator commands know which
// framework / architecture / module the project uses without re-running the
// wizard.
package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const filename = "gostack.json"

// Meta holds project-level metadata written by `gostack create`.
type Meta struct {
	ProjectName  string `json:"project_name"`
	ModuleName   string `json:"module_name"`
	Type         string `json:"type"`
	Framework    string `json:"framework"`
	Architecture string `json:"architecture"`
	Database     string `json:"database"`
	ORM          string `json:"orm"`
	Auth         string `json:"auth"`
	GoStackVer   string `json:"gostack_version"`
}

// Write saves meta to <root>/gostack.json.
func Write(root string, m Meta) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, filename), data, 0644)
}

// Read loads gostack.json by walking up from cwd.
func Read() (*Meta, error) {
	root, err := findRoot()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(root, filename))
	if err != nil {
		return nil, fmt.Errorf("gostack.json not found — run this inside a GoStack project")
	}
	var m Meta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid gostack.json: %w", err)
	}
	return &m, nil
}

// ReadFromDir loads gostack.json from a specific directory.
func ReadFromDir(dir string) (*Meta, error) {
	data, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return nil, fmt.Errorf("gostack.json not found in %s", dir)
	}
	var m Meta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid gostack.json: %w", err)
	}
	return &m, nil
}

// findRoot walks up from cwd looking for gostack.json.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, filename)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("%s not found", filename)
}
