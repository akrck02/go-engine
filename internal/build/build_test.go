package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileNamesAreLowerCaseWordsJoinedByDashes(test *testing.T) {
	for title, want := range map[string]string{"The keeper's night": "the-keeper-s-night", "  ¡Hola!  ": "hola", "": "game", "Level 2": "level-2"} {
		if got := FileName(title); got != want {
			test.Errorf("FileName(%q) = %q, want %q", title, got, want)
		}
	}
}

func TestAModuleBuildsAgainstThisSource(test *testing.T) {
	source, err := SourceFolder()
	if err != nil {
		test.Fatal(err)
	}
	module, err := ModuleFile("example/game", source)
	if err != nil {
		test.Fatal(err)
	}
	for _, want := range []string{"module example/game", "replace github.com/akrck02/go-engine => " + source} {
		if !strings.Contains(module, want) {
			test.Errorf("go.mod lacks %q:\n%s", want, module)
		}
	}
}

func TestAModuleRepeatsGoEnginesOwnReplacements(test *testing.T) {
	source := test.TempDir()
	own := "module github.com/akrck02/go-engine\n\nreplace example.com/lamp => example.com/lamp v1.2.0\n"
	if err := os.WriteFile(filepath.Join(source, "go.mod"), []byte(own), 0o644); err != nil {
		test.Fatal(err)
	}
	module, err := ModuleFile("example/game", source)
	if err != nil {
		test.Fatal(err)
	}
	if !strings.Contains(module, "replace example.com/lamp => example.com/lamp v1.2.0") {
		test.Errorf("a replacement of go-engine's own is not repeated:\n%s", module)
	}
}
