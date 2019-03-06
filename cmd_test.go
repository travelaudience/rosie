package rosie_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/travelaudience/rosie"

	"github.com/travelaudience/rosie/pkg/runner/testrunner"
)

func TestCmd_optimistic(t *testing.T) {
	g := rosie.Group("test-group")
	g.Beginning().
		Then(rosie.Cmd("command", "ls", "-lha")).
		Then(rosie.Fn("assert", rosie.StringSliceClosure(func(_ context.Context, _ io.Writer, res []string) (interface{}, error) {
			for _, val := range res {
				if strings.Contains(val, "cmd.go") {
					return val, nil
				}
			}

			return nil, errors.New("empty cmd.go to be present")
		}))).
		Then(rosie.Cmd("command", "echo", "[[.Result.Value]]")).
		Then(rosie.Fn("assert", rosie.StringSliceClosure(func(_ context.Context, _ io.Writer, res []string) (interface{}, error) {
			if len(res) != 1 {
				return nil, errors.New("wrong slice length, expected 1 element")
			}
			if strings.Contains(res[0], "cmd.go") {
				return res[0], nil
			}

			return nil, errors.New("expected cmd.go to be present")
		}))).
		Then(rosie.Dir(rosie.Cmd("command", "ls", "-lha"), "pkg/runner/testrunner")).
		Then(rosie.Fn("assert", rosie.StringSliceClosure(func(_ context.Context, _ io.Writer, res []string) (interface{}, error) {
			for _, v := range res {
				if strings.Contains(v, "runner.go") {
					return v, nil
				}
			}

			return nil, errors.New("empty cmd.go to be present")
		}))).
		Then(rosie.Env(rosie.Cmd("command", "printenv", "TEST_VAR"), "TEST_VAR=TACTL_OK")).
		Then(rosie.Fn("assert", rosie.StringSliceClosure(func(_ context.Context, _ io.Writer, res []string) (interface{}, error) {
			for _, v := range res {
				if strings.Contains(v, "TACTL_OK") {
					return v, nil
				}
			}

			return nil, errors.New("expected OK to be present")
		})))

	testrunner.Run(t, g, noError)
}

func TestCmd_pessimisticLackOfCommands(t *testing.T) {
	defer assertPanicInitError(t)

	g := rosie.Group("test-group")
	g.Beginning().
		Then(rosie.Cmd("command"))

	testrunner.Run(t, g, noError)
}

func TestCmd_pessimisticMalformedTemplate(t *testing.T) {
	defer assertPanicInitError(t)

	g := rosie.Group("test-group")
	g.Beginning().
		Then(rosie.Cmd("command", "echo", "[[.Result"))

	testrunner.Run(t, g, noError)
}

func TestCmd_pessimisticMissingTemplateArguments(t *testing.T) {
	defer assertPanicInitError(t)

	g := rosie.Group("test-group")
	g.Beginning().
		Then(rosie.Cmd("command", "echo", "[[.NotAResult]]"))

	testrunner.Run(t, g, noError)
}

func assertPanicInitError(t *testing.T) {
	if err := recover(); err != nil {
		if _, ok := err.(*rosie.InitError); ok {
			return
		}
		t.Errorf("expected panic to carry InitError, got %T", err)
	}
	t.Error("expected panic")
}
