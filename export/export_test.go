package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akrck02/go-engine/game"
)

func TestFileNamesAreLowerCaseWordsJoinedByDashes(t *testing.T) {
	for title, want := range map[string]string{"The keeper's night": "the-keeper-s-night", "  ¡Hola!  ": "hola", "": "game", "Level 2": "level-2"} {
		if got := FileName(title); got != want {
			t.Errorf("FileName(%q) = %q, want %q", title, got, want)
		}
	}
}

func TestTheProgramIsWrittenAgainstThisSource(t *testing.T) {
	source, err := sourceFolder()
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
