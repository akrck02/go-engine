package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akrck02/go-engine/grid/game"
	"github.com/akrck02/go-engine/internal/build"
)

func TestTheProgramIsWrittenAgainstThisSource(t *testing.T) {
	source, err := build.SourceFolder()
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	if err := writeProgram(game.Sample(), source, work); err != nil {
		t.Fatal(err)
	}
	module, _ := os.ReadFile(filepath.Join(work, "go.mod"))
	for _, want := range []string{"replace github.com/akrck02/go-engine => " + filepath.ToSlash(source), "replace github.com/akrck02/go-renderer =>"} {
		if !strings.Contains(string(module), want) {
			t.Errorf("go.mod lacks %q:\n%s", want, module)
		}
	}
	data, _ := os.ReadFile(filepath.Join(work, "game.json"))
	if _, err := game.Read(data); err != nil {
		t.Fatal(err)
	}
}
