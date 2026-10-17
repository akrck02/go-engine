package build

import (
	"strings"
	"testing"
)

func TestFileNamesAreLowerCaseWordsJoinedByDashes(t *testing.T) {
	for title, want := range map[string]string{"The keeper's night": "the-keeper-s-night", "  ¡Hola!  ": "hola", "": "game", "Level 2": "level-2"} {
		if got := FileName(title); got != want {
			t.Errorf("FileName(%q) = %q, want %q", title, got, want)
		}
	}
}

func TestAModuleBuildsAgainstThisSource(t *testing.T) {
	source, err := SourceFolder()
	if err != nil {
		t.Fatal(err)
	}
	module, err := ModuleFile("example/game", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"module example/game", "replace github.com/akrck02/go-engine => " + source, "replace github.com/akrck02/go-renderer =>"} {
		if !strings.Contains(module, want) {
			t.Errorf("go.mod lacks %q:\n%s", want, module)
		}
	}
}
