package game

// Sample is a small game to start from: a lighthouse keeper who pushes crates out of the way,
// gathers the lamp oil and climbs to the lantern, in two scenes.
func Sample() *Game {
	sample := &Game{
		Title: "The keeper's night",
		Kinds: []Kind{
			{Name: "Keeper", Colour: "#2f4858", Shape: Rounded, Icon: "user", Behaviour: Behaviour{Player: true}},
			{Name: "Wall", Colour: "#6b705c", Shape: Square, Behaviour: Behaviour{Solid: true}},
			{Name: "Crate", Colour: "#b08968", Shape: Rounded, Icon: "package", Behaviour: Behaviour{Pushable: true}},
			{Name: "Oil", Colour: "#f6bd60", Shape: Small, Behaviour: Behaviour{Collectable: true}},
			{Name: "Lantern", Colour: "#e76f51", Shape: Circle, Icon: "lamp", Behaviour: Behaviour{Goal: true}},
			{Name: "Rocks", Colour: "#3d405b", Shape: Rounded, Icon: "mountain", Behaviour: Behaviour{Hazard: true}},
		},
	}
	sample.Scenes = append(sample.Scenes,
		sceneFrom("The yard", "#cfe1b9", []string{
			"############",
			"#K.........#",
			"#..C...O...#",
			"#..C.......#",
			"#.....###..#",
			"#..O..#L...#",
			"#.....###..#",
			"############",
		}),
		sceneFrom("The cliff", "#a9d6e5", []string{
			"##############",
			"#K...R....O..#",
			"#..C.R..C....#",
			"#....R.......#",
			"#.O..R..RRR..#",
			"#.......R.L..#",
			"##############",
		}),
	)
	return sample
}

// sceneFrom lays a scene out from rows of letters: # a wall, K the keeper, C a crate, O oil, L the
// lantern, R rocks, and a dot an empty cell.
func sceneFrom(name, background string, rows []string) Scene {
	kinds := map[rune]string{'#': "Wall", 'K': "Keeper", 'C': "Crate", 'O': "Oil", 'L': "Lantern", 'R': "Rocks"}
	scene := Scene{Name: name, Width: len(rows[0]), Height: len(rows), Background: background}
	for y, row := range rows {
		for x, letter := range row {
			if kind, found := kinds[letter]; found {
				scene.Things = append(scene.Things, Thing{Kind: kind, X: x, Y: y})
			}
		}
	}
	return scene
}
