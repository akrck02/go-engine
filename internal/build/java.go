package build

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Android's builds compile go-renderer's Java and sign the package with a JDK's tools, and
// Android's sdkmanager runs on one: a JDK, not only Java's runtime, as it is javac that is needed.

// FindJava finds a JDK's folder (what $JAVA_HOME names): the one chosen, $JAVA_HOME, the one
// whose javac is on the path, or Android Studio's.
func (tools Tools) FindJava() (string, error) {
	if place, isChosen, err := chosenFolder(tools.Java); isChosen {
		return place, err
	}
	if home := os.Getenv("JAVA_HOME"); home != "" && FileExists(home) {
		return home, nil
	}
	if compiler, err := exec.LookPath("javac"); err == nil {
		return javaHomeOf(compiler), nil
	}
	if home := firstExisting(javaPlaces()); home != "" {
		return home, nil
	}
	return "", JavaNotFound
}

// javaHomeOf is the JDK's folder of its javac: the one on the path is often a link to the JDK's
// own bin, whose parent is the folder.
func javaHomeOf(compiler string) string {
	if real, err := filepath.EvalSymlinks(compiler); err == nil {
		compiler = real
	}
	return filepath.Dir(filepath.Dir(compiler))
}
