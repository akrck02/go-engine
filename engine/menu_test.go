package engine

import (
	"testing"
	"testing/fstest"
)

func TestTheMenuListsTheScenesByTheirNames(test *testing.T) {
	files := fstest.MapFS{
		"scenes/level-1.scene.json": {Data: []byte(`{"name": "The harbour", "mode": "2d", "nodes": []}`)},
		"scenes/level-2.scene.json": {Data: []byte(`{"mode": "2d", "nodes": []}`)},
		"scenes/notes.txt":          {Data: []byte("not a scene")},
	}
	scenes := scenesOf(files)
	if len(scenes) != 2 || scenes[0] != (MenuScene{"scenes/level-1.scene.json", "The harbour"}) ||
		scenes[1] != (MenuScene{"scenes/level-2.scene.json", "level-2"}) {
		test.Fatalf("the menu offers %+v", scenes)
	}
}

func TestChoosingASceneOfTheMenuPlaysItNext(test *testing.T) {
	game := &runner{played: platform(test)}
	game.choose("scenes/level-2.scene.json")
	if game.played.changeRequested != "scenes/level-2.scene.json" {
		test.Fatalf("choosing a scene asks for %q", game.played.changeRequested)
	}
}

func TestAGamesOwnInterfaceIsGivenItsScenes(test *testing.T) {
	game := &runner{played: platform(test), menu: []MenuScene{{"scenes/a.scene.json", "A"}}}
	menu := Menu{game}
	if len(menu.Scenes()) != 1 || menu.Playing() != game.played {
		test.Fatalf("the menu gives %+v, playing %v", menu.Scenes(), menu.Playing())
	}
	menu.Play("scenes/a.scene.json")
	if game.played.changeRequested != "scenes/a.scene.json" {
		test.Fatal("playing a scene of the menu does not ask for it")
	}
}
