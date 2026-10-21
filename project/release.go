package project

import (
	"strings"

	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/internal/build"
	"github.com/akrck02/go-engine/internal/notices"
)

// A game's desktop programs for Windows, macOS and Linux are built where each system is, on
// GitHub's machines, when a version tag is pushed: Go builds for another system only with cgo's C
// compiler and that system's libraries, which GLFW needs. The editor writes the workflow into the
// game's repository; GitHub attaches each program, and the notices of the code they carry, to the
// tag's release.

// ReleaseWorkflowFile is where the workflow goes in the game's folder.
const ReleaseWorkflowFile = ".github/workflows/release.yml"

// releaseNoticesFile is the notices of the code a desktop program carries, kept beside the
// workflow for it to attach.
const releaseNoticesFile = "THIRD_PARTY_NOTICES.md"

// WriteReleaseWorkflow writes the workflow that builds the game for Windows, macOS and Linux on
// each version tag (v1.0.0), and the notices it attaches. It returns the workflow's path.
// A game whose go.mod replaces go-engine with a folder of this computer cannot be built elsewhere:
// it is refused, to be updated first (go-engine -update).
func (project *Project) WriteReleaseWorkflow() (string, error) {
	if _, localCopy, err := project.EngineVersion(); err != nil || localCopy != "" {
		return "", failure.Wrap(ReleaseNeedsPublishedEngine, localCopy, err)
	}
	workflow := project.Path(ReleaseWorkflowFile)
	if err := writeFile(workflow, []byte(project.releaseWorkflow())); err != nil {
		return "", err
	}
	return workflow, writeFile(project.Path(releaseNoticesFile), notices.Desktop)
}

// releaseWorkflow is the workflow's text, with the game's program name in it.
func (project *Project) releaseWorkflow() string {
	return strings.ReplaceAll(releaseWorkflowTemplate, "{program}", build.FileName(project.Settings.Name))
}

// releaseWorkflowTemplate builds on each system's machine with the Go of the game's go.mod, and
// uploads with GitHub's own command-line tool: no action beyond GitHub's own.
var releaseWorkflowTemplate = strings.TrimLeft(`
# Written by Go Engine: builds the game for Windows, macOS and Linux on each version tag, and
# attaches the programs to the tag's release.
name: Release

on:
  push:
    tags: ["v*"]

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Make the release
        env:
          GH_TOKEN: ${{ github.token }}
        run: gh release view "$GITHUB_REF_NAME" || gh release create "$GITHUB_REF_NAME" --generate-notes THIRD_PARTY_NOTICES.md

  build:
    needs: release
    strategy:
      matrix:
        include:
          - machine: ubuntu-latest
            program: {program}-linux-amd64
          - machine: windows-latest
            program: {program}-windows-amd64.exe
          - machine: macos-latest
            program: {program}-macos-arm64
    runs-on: ${{ matrix.machine }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: Install GLFW's libraries
        if: runner.os == 'Linux'
        run: sudo apt-get update && sudo apt-get install -y libgl1-mesa-dev xorg-dev
      - name: Build
        run: go build -tags release -o ${{ matrix.program }} .
      - name: Attach to the release
        env:
          GH_TOKEN: ${{ github.token }}
        shell: bash
        run: gh release upload "$GITHUB_REF_NAME" "${{ matrix.program }}" --clobber
`, "\n")
