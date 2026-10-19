package engine

import (
	"runtime/debug"
	"testing"
)

func TestTheVersionSaysWhichGoEngineItIs(test *testing.T) {
	pinned := &debug.BuildInfo{Main: debug.Module{Path: "example.com/lights"},
		Deps: []*debug.Module{{Path: Module, Version: "v0.0.0-20261018185731-ba085231fc85"}}}
	tagged := &debug.BuildInfo{Main: debug.Module{Path: "example.com/lights"}, Deps: []*debug.Module{{Path: Module, Version: "v0.0.2"}}}
	local := &debug.BuildInfo{Main: debug.Module{Path: "example.com/lights"}, Deps: []*debug.Module{{Path: Module, Version: "(devel)"}}}
	itself := &debug.BuildInfo{Main: debug.Module{Path: Module, Version: "(devel)"}, Settings: []debug.BuildSetting{
		{Key: "vcs.revision", Value: "abc1234def5678"}, {Key: "vcs.modified", Value: "true"}}}
	replaced := &debug.BuildInfo{Main: debug.Module{Path: "example.com/lights"},
		Deps: []*debug.Module{{Path: Module, Version: "v0.0.0", Replace: &debug.Module{Path: "../go-engine", Version: "(devel)"}}}}
	cases := []struct {
		built    *debug.BuildInfo
		revision string
		version  string
	}{
		{pinned, "", Release + " (ba08523)"},
		{tagged, "", "v0.0.2"},
		{local, "", Release + " (local)"},
		{local, "77aa001", Release + " (77aa001, local)"},
		{replaced, "77aa001", Release + " (77aa001, local)"},
		{itself, "", Release + " (abc1234, modified)"},
		{nil, "", Release},
		{&debug.BuildInfo{Main: debug.Module{Path: Module, Version: "(devel)"}}, "", Release + " (local)"},
	}
	for _, each := range cases {
		if version := versionFrom(each.built, each.revision); version != each.version {
			test.Errorf("%q, not %q", version, each.version)
		}
	}
}
