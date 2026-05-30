package integration

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestVersion(t *testing.T) {
	outDir := t.TempDir()
	binPath := outDir + "/dotlock"
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	build := exec.Command("go", "build", "-o", binPath, "github.com/ahmadraza100/dotlock/cmd/dotlock")
	if b, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, string(b))
	}
	cmd := exec.Command(binPath, "version")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("dotlock version run failed: %v\n%s", err, string(b))
	}
}