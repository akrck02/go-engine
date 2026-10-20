# failure

```go
import "github.com/akrck02/go-engine/failure"
```

Package failure gives every error go-engine returns a code: a stable name, as
"engine.scene-not-found", that a program can check and a person can search for.

A code is declared once, as a package-level variable, and is itself an error:

```go
var SceneNotFound = failure.Define("engine.scene-not-found", "the scene file does not exist")
```

Returning the code alone allocates nothing and comparing it is comparing two
pointers, as cheap as Go's own sentinel errors. When the error needs its details
(which file, what went wrong underneath), Wrap makes one error with them;
that happens while loading or building, never in a game's frame.

Callers check an error by its code, wrapped or not:

```go
if failure.Is(err, engine.SceneNotFound) { … }
```

errors.Is works the same way, since an error made by Wrap answers to its code.

### Functions

| Function | |
|---|---|
| `func Is(err error, code *Code) bool` | Is tells whether an error, or any error it wraps, has a code. |
| `func Wrap(code *Code, details string, cause error) error` | Wrap makes an error of a code with its details and the error underneath (nil if none). |

### Code

Code is a kind of error: its stable name and what it means. Declare codes
with Define.

```go
type Code struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func CodeOf(err error) *Code` | CodeOf is the code of an error, the outermost one that has one; nil for an error without a code (one from outside go-engine). |
| `func Define(name, meaning string) *Code` | Define declares a code. |
| `func (code *Code) Error() string` | Error is the code's name and meaning: a code is an error on its own. |
| `func (code *Code) Name() string` | Name is the code's stable name, as "engine.scene-not-found". |

### Error

Error is a coded error with its details: what the code means here, and the
error underneath.

```go
type Error struct {
	Code    *Code
	Details string // what is particular to this error: a file, a name
	Cause   error  // the error underneath, if any
}
```

| Function or method | |
|---|---|
| `func (failed *Error) Error() string` | Error is the code's name, the details, and the error underneath. |
| `func (failed *Error) Is(target error) bool` | Is tells whether the error has a code: errors.Is(err, code) asks it. |
| `func (failed *Error) Unwrap() error` | Unwrap is the error underneath, so errors.Is and errors.As reach it. |

---

[← Reference](0-packages.md) · generated from the doc comments by `generate.py`
