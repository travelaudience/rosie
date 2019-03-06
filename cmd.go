package rosie

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/travelaudience/rosie/pkg/dag"
)

var (
	_ Executor = &CmdTask{}
)

// CmdTask is a type of task that executes locally available programs.
type CmdTask struct {
	*task
	closure func(context.Context, Resulter) *exec.Cmd
	wraps   *CmdTask
}

// Cmd instantiate new CmdTask object.
// It requires at least one command to be passed, otherwise, it panics.
// Each and every element in the slice is threatened as a template.
// It understands annotations surrounded by `[[...]]` for example [[.Result.Value]].
func Cmd(name string, commands ...string) *CmdTask {
	if len(commands) == 0 {
		panic(&InitError{
			msg: "command is mandatory",
		})
	}

	t := &CmdTask{
		task: &task{name: name},
	}
	t.closure = func(ctx context.Context, res Resulter) *exec.Cmd {
		buf := bytes.NewBuffer(nil)
		for i, command := range commands {
			tmpl, err := template.New(fmt.Sprintf("%s-%d", name, i)).
				Delims("[[", "]]").
				Parse(command)
			if err != nil {
				panic(&InitError{
					msg: fmt.Sprintf("command template (%s) initialization failure", command),
					err: err,
				})
			}
			if err := tmpl.Execute(buf, struct {
				Result interface{}
			}{
				Result: res.Result(),
			}); err != nil {
				panic(&InitError{
					msg: "command template execution failure",
					err: err,
				})
			}
			commands[i] = buf.String()
			buf.Reset()
		}

		/* #nosec */
		cmd := exec.CommandContext(ctx, commands[0], commands[1:]...)
		cmd.Env = os.Environ()

		t.description = strings.Join(commands, " ")

		return cmd
	}
	t.setAnchor(&dag.Node{}, t)
	return t
}

// Dir is a CmdTask wrapper that changes the directory where a program will be executed.
func Dir(wrapped *CmdTask, dir string) *CmdTask {
	t := &CmdTask{
		task:  &task{name: fmt.Sprintf("dir(%s)", wrapped.name)},
		wraps: wrapped,
	}
	t.closure = func(ctx context.Context, res Resulter) *exec.Cmd {
		cmd := wrapped.closure(ctx, res)
		cmd.Dir = dir
		t.description = wrapped.description + " [" + dir + "]"

		return cmd
	}
	t.setAnchor(&dag.Node{}, t)
	return t
}

// Env is a CmdTask wrapper that extends list environment variables.
func Env(wrapped *CmdTask, env ...string) *CmdTask {
	t := &CmdTask{
		task: &task{name: fmt.Sprintf("env(%s)", wrapped.name)},

		wraps: wrapped,
	}
	t.closure = func(ctx context.Context, res Resulter) *exec.Cmd {
		cmd := wrapped.closure(ctx, res)
		cmd.Env = append(cmd.Env, env...)
		t.description = wrapped.description

		return cmd
	}
	t.setAnchor(&dag.Node{}, t)
	return t
}

// MakeDir creates a directory.
func MakeDir(dir string) *CmdTask {
	return Cmd("mkdir", "mkdir", "-p", dir)
}

// RemoveDir removes a directory and all files/directories inside.
func RemoveDir(dir string) *CmdTask {
	return Cmd("rmdir", "rm", "-rf", dir)
}

// Exec implements Executor interface.
func (t *CmdTask) Exec(ctx context.Context) (<-chan Piece, error) {
	previousResulter := t.gatherParentResults()
	previousResult := previousResulter.Result()

	cmd := t.closure(ctx, previousResulter)
	out := make(chan Piece)

	stdres := bytes.NewBuffer(nil)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		t.setErr(err)
		return nil, err
	}

	go func() {
		sc := bufio.NewScanner(io.TeeReader(stdout, stdres))
		for sc.Scan() {
			out <- Piece{Text: sc.Text()}
		}
		sc = bufio.NewScanner(stderr)
		for sc.Scan() {
			out <- Piece{Text: sc.Text()}
		}

		if err := cmd.Wait(); err != nil {
			out <- Piece{Err: err}
			t.setErr(err)
			close(out)
			return
		}

		var res []string
		sc = bufio.NewScanner(stdres)
		for sc.Scan() {
			res = append(res, sc.Text())
		}

		t.setResult(Result{key: previousResult.key, value: res})
		if err := t.task.run(); err != nil {
			out <- Piece{Err: err}
			t.setErr(err)
			close(out)
			return
		}
		close(out)
	}()

	return out, nil
}
