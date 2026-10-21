package project

import (
	"errors"
	"testing"
)

func TestATargetWithoutAnEmulatorSaysSoByItsCode(test *testing.T) {
	var project Project
	for _, target := range []Target{Desktop, Android} {
		if _, err := project.PlayInEmulator(target); !errors.Is(err, TargetCannot) {
			test.Errorf("an emulator of %v gives %v", target, err)
		}
	}
}

func TestStoppingIsOnlyNeededWhereTheGameOutlivesItsCommand(test *testing.T) {
	var project Project
	for _, target := range []Target{Desktop, Nintendo3DS} {
		if err := project.StopOn(target); err != nil {
			test.Errorf("stopping on %v gives %v", target, err)
		}
	}
}

func TestEveryTargetHasAName(test *testing.T) {
	for _, target := range Targets {
		if target.String() == "" {
			test.Errorf("target %d has no name", target)
		}
	}
}
