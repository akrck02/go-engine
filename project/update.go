package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/internal/build"
)

// A game is a Go module that requires go-engine at one version. Updating it requires go-engine
// at the newest commit of a branch (or a tag, or a commit) instead, from its repository: a local
// copy that go.mod replaced it with is let go of. As a module's replaces apply only to itself,
// the game repeats go-engine's: the old version's are let go of, and the new version's copied. Go reads go.mod only: a go.work beside the game is left out (GOWORK=off).

// LatestEngine is the branch go-engine's latest version is taken from.
const LatestEngine = "main"

// onlyModule leaves a go.work out of what Go reads: updating is go.mod's.
var onlyModule = []string{"GOWORK=off"}

// EngineVersion is the version of go-engine the project's go.mod requires, and the folder that
// stands in for it when go.mod replaces it with a local copy (empty when none does).
func (project *Project) EngineVersion() (version, localCopy string, err error) {
	output, err := build.Output(project.Folder, onlyModule, build.GoTool(), "list", "-m", "-f",
		"{{.Version}} {{with .Replace}}{{if not .Version}}{{.Path}}{{end}}{{end}}", engine.Module)
	if err != nil {
		return "", "", failure.Wrap(EngineNotRead, project.Folder, err)
	}
	fields := strings.Fields(output)
	if len(fields) == 0 {
		return "", "", failure.Wrap(EngineNotRead, project.Folder, nil)
	}
	if len(fields) > 1 {
		localCopy = fields[1]
	}
	return fields[0], localCopy, nil
}

// UpdateEngine requires go-engine at a version (a branch as LatestEngine, a tag or a commit) with
// its replaces, and tidies go.mod; it returns the version required now.
func (project *Project) UpdateEngine(version string) (string, error) {
	oldModule, err := project.engineModuleFile()
	if err != nil {
		return "", err
	}
	release, err := project.engineRelease(version)
	if err != nil {
		return "", err
	}
	old, err := engineReplaces(oldModule)
	if err != nil {
		return "", err
	}
	replaces, err := engineReplaces(release.GoMod)
	if err != nil {
		return "", err
	}
	arguments := []string{"mod", "edit", "-dropreplace=" + engine.Module}
	for _, replace := range old {
		arguments = append(arguments, "-dropreplace="+replace.old)
	}
	arguments = append(arguments, "-require="+engine.Module+"@"+release.Version)
	for _, replace := range replaces {
		arguments = append(arguments, "-replace="+replace.old+"="+replace.replacement)
	}
	if err := build.Run(project.Folder, onlyModule, build.GoTool(), arguments...); err != nil {
		return "", failure.Wrap(EngineNotUpdated, release.Version, err)
	}
	if err := project.Tidy(); err != nil {
		return "", err
	}
	return release.Version, nil
}

// engineModuleFile is the go.mod of the go-engine the project requires now: its local copy's, or
// the version's as Go downloads it.
func (project *Project) engineModuleFile() (string, error) {
	version, localCopy, err := project.EngineVersion()
	if err != nil {
		return "", err
	}
	if localCopy == "" {
		release, err := project.engineRelease(version)
		return release.GoMod, err
	}
	if !filepath.IsAbs(localCopy) {
		localCopy = filepath.Join(project.Folder, localCopy)
	}
	return filepath.Join(localCopy, "go.mod"), nil
}

// engineRelease is a version of go-engine as Go downloads it: its exact version (a pseudo-version
// for a branch) and its go.mod on disk.
type engineRelease struct {
	Version string
	GoMod   string
}

// engineRelease asks Go for go-engine at a version.
func (project *Project) engineRelease(version string) (engineRelease, error) {
	version, err := branchCommit(version)
	if err != nil {
		return engineRelease{}, err
	}
	var release engineRelease
	output, err := build.Output(project.Folder, onlyModule, build.GoTool(), "mod", "download", "-json", engine.Module+"@"+version)
	if err == nil {
		err = json.Unmarshal([]byte(output), &release)
	}
	if err != nil || release.Version == "" || release.GoMod == "" {
		return engineRelease{}, failure.Wrap(EngineNotUpdated, version, err)
	}
	return release, nil
}

// engineRepository is where go-engine's branches are.
const engineRepository = "https://" + engine.Module

// branchCommit is the commit a branch whose name has a slash (feature/engine) is at, which Go
// takes as a version where it does not take the name; any other version is as it was given.
func branchCommit(version string) (string, error) {
	if !strings.Contains(version, "/") {
		return version, nil
	}
	output, err := build.Output("", nil, "git", "ls-remote", engineRepository, "refs/heads/"+version)
	fields := strings.Fields(output)
	if err != nil || len(fields) == 0 {
		return "", failure.Wrap(EngineNotUpdated, version+": no such branch", err)
	}
	return fields[0], nil
}

// replaceDirective is a replace of one line in a go.mod: what is replaced, and by what.
var replaceDirective = regexp.MustCompile(`(?m)^replace\s+(\S+(?:\s+v\S+)?)\s+=>\s+(\S+)(?:\s+(v\S+))?\s*$`)

// replacement is a replace directive: what is replaced (module@version, or a module in all its
// versions), and by what (module@version).
type replacement struct{ old, replacement string }

// engineReplaces are go-engine's own replaces with a version, in a go.mod.
func engineReplaces(goMod string) ([]replacement, error) {
	data, err := os.ReadFile(goMod)
	if err != nil {
		return nil, failure.Wrap(EngineNotUpdated, goMod, err)
	}
	var replaces []replacement
	for _, match := range replaceDirective.FindAllStringSubmatch(string(data), -1) {
		if match[3] == "" {
			continue // a folder of go-engine's own computer: nothing a game can repeat
		}
		replaces = append(replaces, replacement{old: strings.Join(strings.Fields(match[1]), "@"), replacement: match[2] + "@" + match[3]})
	}
	return replaces, nil
}
