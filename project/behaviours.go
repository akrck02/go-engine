package project

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// Behaviour is a behaviour the project's code defines: a struct type with a Start or an
// Update method, as the editor offers it.
type Behaviour struct {
	Name   string
	Doc    string // the type's comment, first sentence
	File   string // where it is defined, in the project
	Fields []Field
}

// Field is a field the scene can set: an exported field, by its JSON name.
type Field struct {
	Name    string // as the scene names it: the JSON name, or the field's own
	Type    string // as written in the code: float64, string, bool, space.Vector…
	Comment string
}

// Behaviours reads the project's Go files (not compiling them) and finds its behaviours, so the
// editor can offer them while the code is being written in another editor.
func (project *Project) Behaviours() []Behaviour {
	files := token.NewFileSet()
	packages, err := parser.ParseDir(files, project.Folder, func(behaviour fs.FileInfo) bool {
		return !strings.HasSuffix(behaviour.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil && len(packages) == 0 {
		return nil
	}
	structs := map[string]*Behaviour{}
	withMethods := map[string]bool{}
	for _, parsed := range packages {
		for path, file := range parsed.Files {
			relative, _ := filepath.Rel(project.Folder, path)
			for _, declaration := range file.Decls {
				switch declared := declaration.(type) {
				case *ast.GenDecl:
					collectStructs(declared, relative, structs)
				case *ast.FuncDecl:
					if receiver := receiverType(declared); receiver != "" && (declared.Name.Name == "Update" || declared.Name.Name == "Start") {
						withMethods[receiver] = true
					}
				}
			}
		}
	}
	var found []Behaviour
	for name, behaviour := range structs {
		if withMethods[name] {
			found = append(found, *behaviour)
		}
	}
	sort.Slice(found, func(first, second int) bool { return found[first].Name < found[second].Name })
	return found
}

func collectStructs(declaration *ast.GenDecl, file string, into map[string]*Behaviour) {
	for _, spec := range declaration.Specs {
		typeSpec, isType := spec.(*ast.TypeSpec)
		if !isType {
			continue
		}
		structType, isStruct := typeSpec.Type.(*ast.StructType)
		if !isStruct {
			continue
		}
		doc := typeSpec.Doc
		if doc == nil {
			doc = declaration.Doc
		}
		behaviour := &Behaviour{Name: typeSpec.Name.Name, Doc: firstSentence(doc.Text()), File: file}
		for _, field := range structType.Fields.List {
			for _, name := range field.Names {
				if !name.IsExported() {
					continue
				}
				fieldName := name.Name
				if field.Tag != nil {
					tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("json")
					if jsonName, _, _ := strings.Cut(tag, ","); jsonName == "-" {
						continue
					} else if jsonName != "" {
						fieldName = jsonName
					}
				}
				behaviour.Fields = append(behaviour.Fields, Field{Name: fieldName, Type: typeString(field.Type), Comment: strings.TrimSpace(field.Comment.Text())})
			}
		}
		into[behaviour.Name] = behaviour
	}
}

// receiverType is the type a method is declared on, without its pointer.
func receiverType(function *ast.FuncDecl) string {
	if function.Recv == nil || len(function.Recv.List) == 0 {
		return ""
	}
	expression := function.Recv.List[0].Type
	if star, isPointer := expression.(*ast.StarExpr); isPointer {
		expression = star.X
	}
	if identifier, isIdentifier := expression.(*ast.Ident); isIdentifier {
		return identifier.Name
	}
	return ""
}

func typeString(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typeString(typed.X) + "." + typed.Sel.Name
	case *ast.StarExpr:
		return "*" + typeString(typed.X)
	case *ast.ArrayType:
		return "[]" + typeString(typed.Elt)
	case *ast.MapType:
		return "map[" + typeString(typed.Key) + "]" + typeString(typed.Value)
	}
	return "?"
}

func firstSentence(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if end := strings.Index(text, ". "); end >= 0 {
		return text[:end+1]
	}
	return text
}
