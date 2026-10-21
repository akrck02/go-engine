package project

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTheReleaseWorkflowBuildsForEachSystem(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Harbour Lights")
	if err != nil {
		test.Fatal(err)
	}
	workflow := created.releaseWorkflow()
	for _, wanted := range []string{"harbour-lights-linux-amd64", "harbour-lights-windows-amd64.exe", "harbour-lights-macos-arm64",
		"ubuntu-latest", "windows-latest", "macos-latest", "go-version-file: go.mod", "-tags release", `tags: ["v*"]`} {
		if !strings.Contains(workflow, wanted) {
			test.Errorf("the workflow does not have %q", wanted)
		}
	}
	if strings.Contains(workflow, "{program}") || strings.Contains(workflow, "\t") {
		test.Error("the workflow keeps a placeholder, or a tab YAML refuses")
	}
}

func TestAReleaseOfAGameOnALocalEngineIsRefused(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
	}
	if _, err := created.WriteReleaseWorkflow(); !errors.Is(err, ReleaseNeedsPublishedEngine) {
		test.Fatalf("a game on go-engine's folder gives %v", err)
	}
	if _, err := os.Stat(created.Path(ReleaseWorkflowFile)); err == nil {
		test.Fatal("the workflow was written anyway")
	}
}
