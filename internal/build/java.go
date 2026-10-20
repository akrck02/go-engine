package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Android's builds compile go-renderer's Java and sign the package with a JDK's tools, and
// Android's sdkmanager runs on one: a JDK, not only Java's runtime, as it is javac that is needed.
// Version managers (jenv, sdkman, asdf) put scripts of their own on the path and may set
// $JAVA_HOME to their own folder, so a place is only taken when it has javac.

// FindJava finds a JDK's folder (what $JAVA_HOME names): the one chosen, else the first with
// javac of $JAVA_HOME, the folder of the javac on the path, the one java says it runs from,
// macOS's java_home and Android Studio's.
func (tools Tools) FindJava() (string, error) {
	if place, isChosen, err := chosenFolder(tools.Java); isChosen {
		return place, err
	}
	for _, home := range javaCandidates() {
		if home != "" && CheckJava(home) == nil {
			return home, nil
		}
	}
	return "", JavaNotFound
}

// javaCandidates are the folders a JDK may be in, the most telling first. Each is looked for only
// when the ones before it are not JDKs.
func javaCandidates() []string {
	candidates := []string{os.Getenv("JAVA_HOME")}
	if compiler, err := exec.LookPath("javac"); err == nil {
		candidates = append(candidates, javaHomeOf(compiler))
	}
	candidates = append(candidates, javaHomeSaidByJava(), macJavaHome())
	return append(candidates, javaPlaces()...)
}

// javaHomeOf is the JDK's folder of its javac: the one on the path is often a link to the JDK's
// own bin, whose parent is the folder.
func javaHomeOf(compiler string) string {
	if real, err := filepath.EvalSymlinks(compiler); err == nil {
		compiler = real
	}
	return filepath.Dir(filepath.Dir(compiler))
}

// javaHomeSaidByJava is the folder the java on the path says it runs from (its java.home), which
// a version manager's script hides; empty without one.
func javaHomeSaidByJava() string {
	output, err := Output("", nil, "java", "-XshowSettings:properties", "-version")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(output, "\n") {
		if name, value, found := strings.Cut(line, "="); found && strings.TrimSpace(name) == "java.home" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// macJavaHome is the JDK macOS's java_home chooses; empty on other systems or without one.
func macJavaHome() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	output, err := Output("", nil, "/usr/libexec/java_home")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}
