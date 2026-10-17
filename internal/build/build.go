// Package build runs the Go tools for go-engine: it finds go-engine's own source, writes a go.mod
// that builds a program against it, and runs go and TinyGo. The grid exporter and the projects
// both use it.
package build

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// SourceFolder is the folder of go-engine's source, which programs made by the editor build
// against: where this file was when the editor was built.
func SourceFolder() (string, error) {
	_, file, _, found := runtime.Caller(0)
	if !found {
		return "", errors.New("go-engine's source cannot be found")
	}
	folder := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	if _, err := os.Stat(filepath.Join(folder, "go.mod")); err != nil {
		return "", fmt.Errorf("go-engine's source was at %s when the editor was built, and is not there now", folder)
	}
	return folder, nil
}

// replaceLine finds the replace directives in go-engine's own go.mod, which a program built
// against it must repeat: a module's replacements apply only to itself.
var replaceLine = regexp.MustCompile(`(?m)^replace .*$`)

// ModuleFile is the go.mod of a program named module that builds against go-engine's source.
func ModuleFile(module, source string) (string, error) {
	own, err := os.ReadFile(filepath.Join(source, "go.mod"))
	if err != nil {
		return "", err
	}
	var file strings.Builder
	fmt.Fprintf(&file, "module %s\n\ngo 1.25\n\nrequire github.com/akrck02/go-engine v0.0.0\n\n", module)
	fmt.Fprintf(&file, "replace github.com/akrck02/go-engine => %s\n", filepath.ToSlash(source))
	for _, line := range replaceLine.FindAllString(string(own), -1) {
		file.WriteString(line + "\n")
	}
	return file.String(), nil
}

// Sums are go-engine's go.sum, so a program built against it need not fetch what is known.
func Sums(source string) []byte {
	sums, _ := os.ReadFile(filepath.Join(source, "go.sum"))
	return sums
}

// Run runs a tool in a folder, and returns what it printed when it fails.
func Run(folder string, environment []string, tool string, arguments ...string) error {
	_, err := Output(folder, environment, tool, arguments...)
	return err
}

// Output runs a tool in a folder and returns what it printed.
func Output(folder string, environment []string, tool string, arguments ...string) (string, error) {
	command := exec.Command(tool, arguments...)
	command.Dir = folder
	command.Env = append(os.Environ(), environment...)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	if err := command.Run(); err != nil {
		return output.String(), fmt.Errorf("%s %s: %w\n%s", filepath.Base(tool), arguments[0], err, LastLines(output.String(), 12))
	}
	return output.String(), nil
}

// LastLines keeps the end of a tool's output, where its errors are.
func LastLines(text string, count int) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	return strings.Join(lines[max(len(lines)-count, 0):], "\n")
}

// GoTool is the go command: $GO, go on the path, or the one the editor was built with.
func GoTool() string {
	if tool := os.Getenv("GO"); tool != "" {
		return tool
	}
	if tool, err := exec.LookPath("go"); err == nil {
		return tool
	}
	if tool := filepath.Join(runtime.GOROOT(), "bin", "go"); FileExists(tool) {
		return tool
	}
	return "go"
}

// GoOnPath is a PATH that has GoTool's folder first: TinyGo, and the programs it starts, call go.
func GoOnPath() string {
	return "PATH=" + filepath.Dir(GoTool()) + string(os.PathListSeparator) + os.Getenv("PATH")
}

// TinyGoTool is the TinyGo with the nintendo3ds target: $TINYGO, tinygo on the path, or in
// ~/go/bin.
func TinyGoTool() (string, error) {
	if tool := os.Getenv("TINYGO"); tool != "" {
		return tool, nil
	}
	if tool, err := exec.LookPath("tinygo"); err == nil {
		return tool, nil
	}
	if home, err := os.UserHomeDir(); err == nil {
		if tool := filepath.Join(home, "go", "bin", "tinygo"); FileExists(tool) {
			return tool, nil
		}
	}
	return "", errors.New("building for the 3DS needs TinyGo with the nintendo3ds target (github.com/akrck02/tinygo-3ds): set $TINYGO")
}

// FileExists tells whether a path exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ExecutableSuffix is what executables end with on this system.
func ExecutableSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// FileName turns a title into a file name: lower case, words joined by dashes.
func FileName(title string) string {
	var name strings.Builder
	dash := false
	for _, letter := range strings.ToLower(title) {
		switch {
		case letter >= 'a' && letter <= 'z', letter >= '0' && letter <= '9':
			if dash && name.Len() > 0 {
				name.WriteByte('-')
			}
			name.WriteRune(letter)
			dash = false
		default:
			dash = true
		}
	}
	if name.Len() == 0 {
		return "game"
	}
	return name.String()
}
