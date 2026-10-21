package style

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The samples are modules of their own, which take go-engine from the repository and the author's
// libraries by version. When go-engine moves to newer libraries the samples must move with it, or
// they no longer build: this test says which sample was left behind.

// sharedModules are the libraries go-engine and its samples must take at the same version.
var sharedModules = []string{"github.com/akrck02/go-renderer", "github.com/akrck02/go-ui"}

// requiredVersions are the versions a go.mod requires, by module.
func requiredVersions(test *testing.T, file string) map[string]string {
	data, err := os.ReadFile(file)
	if err != nil {
		test.Fatal(err)
	}
	versions := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "require "))
		if len(fields) >= 2 && strings.HasPrefix(fields[1], "v") {
			versions[fields[0]] = fields[1]
		}
	}
	return versions
}

func TestTheSamplesTakeTheLibrariesAtGoEnginesVersions(test *testing.T) {
	engine := requiredVersions(test, filepath.Join("..", "..", "go.mod"))
	samples, err := filepath.Glob(filepath.Join("..", "..", "examples", "*", "go.mod"))
	if err != nil || len(samples) == 0 {
		test.Fatalf("no sample found: %v", err)
	}
	for _, sample := range samples {
		versions := requiredVersions(test, sample)
		for _, module := range sharedModules {
			if found, wanted := versions[module], engine[module]; found != "" && found != wanted {
				test.Errorf("%s takes %s at %s, go-engine at %s: run go mod tidy there", sample, module, found, wanted)
			}
		}
	}
}
