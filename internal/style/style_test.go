package style

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// allowedShortNames are the short names the guide allows: what everyone calls them, and Go's own
// init, which no other name can replace.
var allowedShortNames = map[string]bool{"err": true, "x": true, "y": true, "z": true, "ui": true, "init": true, "_": true}

// shortWords are English words of one or two letters: words, not abbreviations.
var shortWords = map[string]bool{"at": true, "by": true, "to": true, "up": true, "is": true, "of": true, "in": true,
	"on": true, "or": true, "as": true, "be": true, "do": true, "go": true, "if": true, "it": true, "no": true, "an": true}

// abbreviations are shortenings the guide forbids, as names or as the start of one.
var abbreviations = []string{
	"pos", "ctx", "msg", "btn", "idx", "cfg", "conf", "tmp", "temp", "val", "num", "str", "ptr", "img",
	"buf", "src", "dst", "req", "res", "resp", "cnt", "len", "arr", "obj", "elem", "attr", "param",
	"args", "info", "init", "calc", "util", "mgr", "impl", "prev", "curr", "cur", "dir", "fn", "cb",
}

// longestFunction is the most lines a function may take, its signature and braces counted: the
// guide's "about 40", with some room.
const longestFunction = 50

// sourceFile is a parsed Go file of the repository.
type sourceFile struct {
	path   string
	syntax *ast.File
	files  *token.FileSet
}

// repositoryFiles parses every Go file of the repository.
func repositoryFiles(test *testing.T) []sourceFile {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		test.Fatal(err)
	}
	var parsed []sourceFile
	files := token.NewFileSet()
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		syntax, err := parser.ParseFile(files, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		parsed = append(parsed, sourceFile{path: relative, syntax: syntax, files: files})
		return nil
	})
	if walkErr != nil {
		test.Fatal(walkErr)
	}
	return parsed
}

func (file sourceFile) where(node ast.Node) string {
	position := file.files.Position(node.Pos())
	return file.path + ":" + itoa(position.Line)
}

func itoa(number int) string {
	if number == 0 {
		return "0"
	}
	digits := ""
	for number > 0 {
		digits = string(rune('0'+number%10)) + digits
		number /= 10
	}
	return digits
}

func TestNamesAreWholeWords(test *testing.T) {
	for _, file := range repositoryFiles(test) {
		for _, name := range declaredNames(file.syntax) {
			if problem := nameProblem(name.Name); problem != "" {
				test.Errorf("%s: %q %s", file.where(name), name.Name, problem)
			}
		}
	}
}

// declaredNames are the names a file declares: variables, constants, types, functions, their
// parameters and results, fields, and import names.
func declaredNames(syntax *ast.File) []*ast.Ident {
	var names []*ast.Ident
	ast.Inspect(syntax, func(node ast.Node) bool {
		switch declared := node.(type) {
		case *ast.ValueSpec:
			names = append(names, declared.Names...)
		case *ast.TypeSpec:
			names = append(names, declared.Name)
		case *ast.FuncDecl:
			names = append(names, declared.Name)
		case *ast.Field:
			names = append(names, declared.Names...)
		case *ast.ImportSpec:
			if declared.Name != nil {
				names = append(names, declared.Name)
			}
		case *ast.AssignStmt:
			if declared.Tok == token.DEFINE {
				names = append(names, assignedNames(declared)...)
			}
		case *ast.RangeStmt:
			names = append(names, rangeNames(declared)...)
		case *ast.LabeledStmt:
			names = append(names, declared.Label)
		}
		return true
	})
	return names
}

func assignedNames(assignment *ast.AssignStmt) []*ast.Ident {
	var names []*ast.Ident
	for _, left := range assignment.Lhs {
		if name, isName := left.(*ast.Ident); isName {
			names = append(names, name)
		}
	}
	return names
}

func rangeNames(loop *ast.RangeStmt) []*ast.Ident {
	if loop.Tok != token.DEFINE {
		return nil
	}
	var names []*ast.Ident
	for _, expression := range []ast.Expr{loop.Key, loop.Value} {
		if name, isName := expression.(*ast.Ident); isName {
			names = append(names, name)
		}
	}
	return names
}

// nameProblem says what is wrong with a name, or "" when nothing is.
func nameProblem(name string) string {
	if allowedShortNames[name] || allowedShortNames[strings.ToLower(name)] || shortWords[strings.ToLower(name)] {
		return ""
	}
	if len(name) < 3 {
		return "is too short to say what it is"
	}
	for _, word := range camelCaseWords(name) {
		if isAbbreviation(word) {
			return "holds the abbreviation " + word
		}
	}
	return ""
}

func isAbbreviation(word string) bool {
	for _, abbreviation := range abbreviations {
		if word == abbreviation {
			return true
		}
	}
	return false
}

// camelCaseWords are the words of a camelCase name, in lower case: "behaviourInfo" is
// "behaviour" and "info".
func camelCaseWords(name string) []string {
	var words []string
	start := 0
	for index, letter := range name {
		if index > start && letter >= 'A' && letter <= 'Z' {
			words = append(words, strings.ToLower(name[start:index]))
			start = index
		}
	}
	return append(words, strings.ToLower(name[start:]))
}

func TestFunctionsAreShort(test *testing.T) {
	for _, file := range repositoryFiles(test) {
		ast.Inspect(file.syntax, func(node ast.Node) bool {
			function, isFunction := node.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				return true
			}
			lines := file.files.Position(function.End()).Line - file.files.Position(function.Pos()).Line + 1
			if lines > longestFunction && !strings.HasSuffix(file.path, "_test.go") {
				test.Errorf("%s: %s takes %d lines; split it (at most %d)", file.where(function), function.Name.Name, lines, longestFunction)
			}
			return true
		})
	}
}

func TestNoElseAfterAReturn(test *testing.T) {
	for _, file := range repositoryFiles(test) {
		ast.Inspect(file.syntax, func(node ast.Node) bool {
			condition, isIf := node.(*ast.IfStmt)
			if !isIf || condition.Else == nil || len(condition.Body.List) == 0 {
				return true
			}
			if _, returns := condition.Body.List[len(condition.Body.List)-1].(*ast.ReturnStmt); returns {
				test.Errorf("%s: an else after a return: return early instead", file.where(condition.Else))
			}
			return true
		})
	}
}

func TestErrorsComeWithCodes(test *testing.T) {
	for _, file := range repositoryFiles(test) {
		if strings.HasPrefix(file.path, "failure"+string(os.PathSeparator)) || strings.HasSuffix(file.path, "_test.go") {
			continue
		}
		ast.Inspect(file.syntax, func(node ast.Node) bool {
			call, isCall := node.(*ast.CallExpr)
			if !isCall {
				return true
			}
			selector, isSelector := call.Fun.(*ast.SelectorExpr)
			if !isSelector {
				return true
			}
			pkg, isName := selector.X.(*ast.Ident)
			if !isName {
				return true
			}
			if (pkg.Name == "errors" && selector.Sel.Name == "New") || (pkg.Name == "fmt" && selector.Sel.Name == "Errorf") {
				test.Errorf("%s: %s.%s makes an error without a code: use failure", file.where(call), pkg.Name, selector.Sel.Name)
			}
			return true
		})
	}
}
