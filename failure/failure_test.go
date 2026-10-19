package failure

import (
	"errors"
	"io/fs"
	"testing"
)

var (
	missing = Define("example.missing", "the thing is not there")
	broken  = Define("example.broken", "the thing is broken")
)

func TestACodeIsAnErrorOnItsOwn(test *testing.T) {
	var err error = missing
	if !Is(err, missing) || Is(err, broken) || CodeOf(err) != missing {
		test.Fatal("a bare code is not its own code")
	}
	if err.Error() != "example.missing: the thing is not there" {
		test.Fatalf("message: %q", err.Error())
	}
}

func TestAWrappedErrorKeepsItsCodeAndItsCause(test *testing.T) {
	err := Wrap(missing, "scenes/main.scene.json", fs.ErrNotExist)
	if !Is(err, missing) || !errors.Is(err, missing) || !errors.Is(err, fs.ErrNotExist) || Is(err, broken) {
		test.Fatal("the wrapped error lost its code or its cause")
	}
	if CodeOf(err) != missing {
		test.Fatal("CodeOf missed the code")
	}
	outer := Wrap(broken, "while loading", err)
	if CodeOf(outer) != broken || !Is(outer, missing) {
		test.Fatal("the outer code must win, and the inner still be found")
	}
	if CodeOf(errors.New("plain")) != nil {
		test.Fatal("a plain error has a code")
	}
}

func TestReturningACodeAllocatesNothing(test *testing.T) {
	find := func() error { return missing }
	allocations := testing.AllocsPerRun(1000, func() {
		if !Is(find(), missing) {
			test.Fatal("not found")
		}
	})
	if allocations != 0 {
		test.Fatalf("returning and checking a code allocates %v times", allocations)
	}
}
