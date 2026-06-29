package updater_test

import (
	"testing"

	"github.com/alifkhasan01/gostack-cli/internal/updater"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		current string
		latest  string
		want    bool
	}{
		{"0.1.0", "v0.2.0", true},
		{"0.2.0", "v0.2.0", false},
		{"0.2.0", "v0.1.0", false},
		{"dev", "v0.2.0", false},   // dev build never updates
		{"", "v0.2.0", false},      // empty version never updates
		{"0.1.0", "v0.1.1", true},
		{"0.9.0", "v0.10.0", true},
		{"1.0.0", "v1.0.0", false},
	}

	for _, tc := range cases {
		got := updater.IsNewer(tc.current, tc.latest)
		if got != tc.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
		}
	}
}

func TestPlatformAssetName(t *testing.T) {
	name := updater.PlatformAssetName()
	if name == "" {
		t.Error("PlatformAssetName() returned empty string")
	}
	// Must contain OS and arch
	if len(name) < 8 {
		t.Errorf("PlatformAssetName() too short: %q", name)
	}
}
