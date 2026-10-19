package project

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/internal/build"
)

// A game is a Go module that requires go-engine at one version. Updating it requires go-engine
// at the newest commit of a branch (or a tag, or a commit) instead, from its repository: a local
// copy that go.mod replaced it with is let go of. As a module's replaces apply only to itself,
// the game repeats go-engine's (go-renderer's branch, today); they are copied from the new
// version's go.mod. Go reads go.mod only: a go.work beside the game is left out (GOWORK=off).

// LatestEngine is the branch go-engine's latest version is taken from.
const LatestEngine = "main"

// onlyModule leaves a go.work out of what Go reads: updating is go.mod's.
var onlyModule = []string{"GOWORK=off"}

// EngineVersion is the version of go-engine the project's go.mod requires, and the folder that
// stands in for it when go.mod replaces it with a local copy (empty when none does).
func (project *Project) EngineVersion() (version, localCopy string, err error) {
	output, err := build.Output(project.Folder, onlyModule, build.GoTool(), "list", "-m", "-f",
		"{{.Version}} {{with .Replace}}{{if not .Version}}{{.Path}}{{end}}{{end}}", engineModule)
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
	release, err := project.engineRelease(version)
	if err != nil {
		return "", err
	}
	replaces, err := engineReplaces(release.GoMod)
	if err != nil {
		return "", err
	}
	arguments := append([]string{"mod", "edit", "-dropreplace=" + engineModule, "-require=" + engineModule + "@" + release.Version}, replaces...)
	if err := build.Run(project.Folder, onlyModule, build.GoTool(), arguments...); err != nil {
		return "", failure.Wrap(EngineNotUpdated, release.Version, err)
	}
	if err := project.Tidy(); err != nil {
		return "", err
	}
	return release.Version, nil
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
	output, err := build.Output(project.Folder, onlyModule, build.GoTool(), "mod", "download", "-json", engineModule+"@"+version)
	if err == nil {
		err = json.Unmarshal([]byte(output), &release)
	}
	if err != nil || release.Version == "" || release.GoMod == "" {
		return engineRelease{}, failure.Wrap(EngineNotUpdated, version, err)
	}
	return release, nil
}

// engineRepository is where go-engine's branches are.
const engineRepository = "https://" + engineModule

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

// engineReplaces are go-engine's own replaces, in a go.mod, as go mod edit's -replace arguments.
func engineReplaces(goMod string) ([]string, error) {
	data, err := os.ReadFile(goMod)
	if err != nil {
		return nil, failure.Wrap(EngineNotUpdated, goMod, err)
	}
	var arguments []string
	for _, match := range replaceDirective.FindAllStringSubmatch(string(data), -1) {
		if match[3] == "" {
			continue // a folder of go-engine's own computer: nothing a game can repeat
		}
		old := strings.Join(strings.Fields(match[1]), "@")
		arguments = append(arguments, "-replace="+old+"="+match[2]+"@"+match[3])
	}
	return arguments, nil
}
