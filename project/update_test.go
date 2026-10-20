package project

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestAProjectSaysWhichEngineItRequires(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
	}
	version, _, err := created.EngineVersion()
	if err != nil || version == "" {
		test.Fatalf("the engine required is %q, %v", version, err)
	}
}

func TestTheEnginesVersionedReplacesAreRepeated(test *testing.T) {
	goMod := filepath.Join(test.TempDir(), "go.mod")
	content := "module example.com/engine\n\n" +
		"// a library's branch\nreplace example.com/renderer => example.com/renderer v0.0.0-20260101000000-abcdef123456\n" +
		"replace example.com/sound v1.2.0 => example.com/fork/sound v1.2.1\n" +
		"replace example.com/local => ../local\n"
	if err := os.WriteFile(goMod, []byte(content), 0o644); err != nil {
		test.Fatal(err)
	}
	replaces, err := engineReplaces(goMod)
	wanted := []replacement{
		{"example.com/renderer", "example.com/renderer@v0.0.0-20260101000000-abcdef123456"},
		{"example.com/sound@v1.2.0", "example.com/fork/sound@v1.2.1"},
	}
	if err != nil || !slices.Equal(replaces, wanted) {
		test.Fatalf("the replaces repeated are %q, %v", replaces, err)
	}
}

func TestAnEngineVersionThatCannotBeFetchedIsSaidByItsCode(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
	}
	test.Setenv("GOPROXY", "off") // nothing is fetched: the version cannot be found
	if _, err := created.UpdateEngine("no-such-version"); !errors.Is(err, EngineNotUpdated) {
		test.Fatalf("an engine version that cannot be fetched gives %v", err)
	}
}

func TestAVersionWithoutASlashIsGivenToGoAsItIs(test *testing.T) {
	for _, version := range []string{"main", "v0.0.1", "efec863"} {
		if commit, err := branchCommit(version); err != nil || commit != version {
			test.Errorf("%q becomes %q, %v", version, commit, err)
		}
	}
}
