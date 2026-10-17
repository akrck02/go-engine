package project

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
)

// AssetEntry is a file or a folder of the project's assets, as the editor's tree shows them.
type AssetEntry struct {
	Name     string
	Path     string // in the project, as "assets/models/house.glb"
	Folder   bool
	Children []AssetEntry
}

// AssetKind is what a file of the assets is, by its extension.
type AssetKind int

const (
	OtherAsset AssetKind = iota
	ModelAsset           // .glb
	ImageAsset           // .png, .jpg: a sprite
	SoundAsset           // .wav, .ogg, .mp3
)

// KindOf tells what an asset file is.
func KindOf(file string) AssetKind {
	switch strings.ToLower(path.Ext(file)) {
	case ".glb":
		return ModelAsset
	case ".png", ".jpg", ".jpeg":
		return ImageAsset
	case ".wav", ".ogg", ".mp3":
		return SoundAsset
	}
	return OtherAsset
}

// AssetTree reads the project's assets folder as a tree: folders first, then files, by name;
// hidden files (as .gitkeep) left out. Assets are whatever is in the folder: copy files there,
// from the file manager or the code editor, and the editor shows them.
func (project *Project) AssetTree() AssetEntry {
	return project.assetEntry(AssetsFolder, "assets")
}

func (project *Project) assetEntry(file, name string) AssetEntry {
	entry := AssetEntry{Name: name, Path: file, Folder: true}
	children, err := os.ReadDir(project.Path(file))
	if err != nil {
		return entry
	}
	for _, child := range children {
		if strings.HasPrefix(child.Name(), ".") {
			continue
		}
		childPath := path.Join(file, child.Name())
		if child.IsDir() {
			entry.Children = append(entry.Children, project.assetEntry(childPath, child.Name()))
			continue
		}
		entry.Children = append(entry.Children, AssetEntry{Name: child.Name(), Path: childPath})
	}
	sort.SliceStable(entry.Children, func(first, second int) bool {
		a, b := entry.Children[first], entry.Children[second]
		if a.Folder != b.Folder {
			return a.Folder
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
	return entry
}

// OpenInZed opens the project's folder in Zed.
func (project *Project) OpenInZed() error {
	for _, program := range []string{"zed", "zeditor", "zed-editor"} {
		if found, err := exec.LookPath(program); err == nil {
			command := exec.Command(found, project.Folder)
			command.Dir = project.Folder
			if err := command.Start(); err != nil {
				return err
			}
			go command.Wait()
			return nil
		}
	}
	return errors.New("Zed was not found (zed, zeditor): install it from zed.dev, or open " + project.Folder + " yourself")
}
