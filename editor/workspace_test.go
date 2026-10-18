package editor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSwitchingWorkspacesKeepsTheSceneAndTheChoice(test *testing.T) {
	editor := open(test)
	if editor.ShownWorkspace() != LevelWorkspace {
		test.Fatalf("the editor opens in %v, not Level", editor.ShownWorkspace().Name())
	}
	cube, _ := editor.Scene.Find("Cube")
	editor.Select(cube)
	scene := editor.Scene
	for _, workspace := range Workspaces {
		editor.ShowWorkspace(workspace)
		if editor.Scene != scene || editor.Selected != cube {
			test.Fatalf("in %s the scene or the choice changed", workspace.Name())
		}
	}
}

func TestThePlayWorkspacePlaysAndLeavingItStops(test *testing.T) {
	editor := open(test)
	editor.ShowWorkspace(ArtWorkspace)
	editor.ShowWorkspace(PlayWorkspace)
	if _, playing := editor.PlayingHere(); !playing {
		test.Fatal("the Play workspace does not play")
	}
	editor.LeavePlay()
	if _, playing := editor.PlayingHere(); playing || editor.ShownWorkspace() != ArtWorkspace {
		test.Fatalf("leaving Play went to %s, playing %v", editor.ShownWorkspace().Name(), playing)
	}
}

// iconsFile is the editor's file of icons.
const iconsFile = "icons.go"

func TestNoIconStandsForTwoThings(test *testing.T) {
	files := token.NewFileSet()
	parsed, err := parser.ParseFile(files, iconsFile, nil, 0)
	if err != nil {
		test.Fatal(err)
	}
	meaningOf := map[string]string{}
	ast.Inspect(parsed, func(node ast.Node) bool {
		spec, isValue := node.(*ast.ValueSpec)
		if !isValue {
			return true
		}
		for index, name := range spec.Names {
			literal, isLiteral := spec.Values[index].(*ast.BasicLit)
			if !isLiteral {
				continue
			}
			icon, _ := strconv.Unquote(literal.Value)
			if other, taken := meaningOf[icon]; taken {
				test.Errorf("%q stands for %s and %s", icon, other, name.Name)
			}
			meaningOf[icon] = name.Name
		}
		return true
	})
}

// iconCalls are go-ui's calls whose first text is an icon.
var iconCalls = []string{"IconButton(", "FlatIconButton(", "PrimaryIconButton(", "MenuButton(", "MenuItem(", "IconSized(", "Selectable(", "treeButton(context, "}

func TestEveryIconComesFromTheIconsFile(test *testing.T) {
	paths, _ := filepath.Glob("*.go")
	for _, path := range paths {
		if path == iconsFile || strings.HasSuffix(path, "_test.go") {
			continue
		}
		data, _ := os.ReadFile(path)
		for number, line := range strings.Split(string(data), "\n") {
			for _, call := range iconCalls {
				if strings.Contains(line, call+`"`) {
					test.Errorf("%s:%d writes an icon by hand: %s", path, number+1, strings.TrimSpace(line))
				}
			}
		}
	}
}

func TestAWorkspaceIsFoundByItsWord(test *testing.T) {
	if workspace, found := WorkspaceNamed("art"); !found || workspace != ArtWorkspace {
		test.Fatal("art is not the Art workspace")
	}
	if _, found := WorkspaceNamed("kitchen"); found {
		test.Fatal("a workspace that does not exist was found")
	}
}
