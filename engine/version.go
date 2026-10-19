package engine

import (
	"runtime/debug"
	"strings"
)

// Module is Go Engine's module path.
const Module = "github.com/akrck02/go-engine"

// Release is the version of Go Engine being made.
const Release = "v0.0.2"

// localRevision is the commit of a local copy of Go Engine a game was built with (a go.work), set
// when go-engine builds it: a local copy's build information does not say.
var localRevision string

// shortRevision is how many characters of a commit say which it is.
const shortRevision = 7

// Version says which Go Engine this program was built with: its release and commit, as
// "v0.0.1 (ba08523)", or "v0.0.1 (77aa001, local)" for a local copy, as the editor's window and
// the logs show it.
func Version() string {
	built, _ := debug.ReadBuildInfo()
	return versionFrom(built, localRevision)
}

// versionFrom reads the version from a program's build information, and the local commit when
// there is one.
func versionFrom(built *debug.BuildInfo, local string) string {
	if built == nil {
		return Release
	}
	if built.Main.Path == Module {
		if revision := ownRevision(built); revision != "" {
			return withCommit(revision)
		}
		return withCommit("local") // go run leaves the commit out
	}
	for _, dependency := range built.Deps {
		if dependency.Path != Module {
			continue
		}
		if replaced := dependency.Replace; replaced != nil && (replaced.Version == "" || replaced.Version == "(devel)") {
			return dependencyVersion("(devel)", local) // replaced by a folder: a local copy
		}
		return dependencyVersion(dependency.Version, local)
	}
	return Release
}

// dependencyVersion reads a game's go-engine requirement: a tag, a commit, or a local copy.
func dependencyVersion(version, local string) string {
	switch {
	case version == "(devel)" && local != "":
		return withCommit(local + ", local")
	case version == "(devel)":
		return withCommit("local")
	case strings.Count(version, "-") >= 2: // a pseudo-version: its last part is the commit
		revision := version[strings.LastIndex(version, "-")+1:]
		return withCommit(revision[:min(shortRevision, len(revision))])
	}
	return version
}

// ownRevision is the commit the program itself was built from, when it is Go Engine's own.
func ownRevision(built *debug.BuildInfo) string {
	revision, modified := "", false
	for _, setting := range built.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value[:min(shortRevision, len(setting.Value))]
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	if modified {
		revision += ", modified"
	}
	return revision
}

func withCommit(commit string) string {
	if commit == "" {
		return Release
	}
	return Release + " (" + commit + ")"
}
