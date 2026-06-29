package runner

import (
	"fmt"
	"os/exec"
)

// GoModTidy runs "go mod tidy" in the given directory.
func GoModTidy(dir string) error {
	fmt.Println("  Running go mod tidy ...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod tidy failed: %w\n%s", err, string(out))
	}
	return nil
}

// GitInit runs "git init" in the given directory.
func GitInit(dir string) error {
	fmt.Println("  Initializing git repository ...")
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init failed: %w\n%s", err, string(out))
	}
	return nil
}
