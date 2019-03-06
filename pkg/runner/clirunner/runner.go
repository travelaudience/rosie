package clirunner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"
	"unicode/utf8"

	"github.com/travelaudience/rosie"
	"github.com/travelaudience/rosie/internal/draw"
	"github.com/travelaudience/rosie/pkg/dag"
)

type Iterator interface {
	Iter() (*rosie.Iterator, error)
}

type Drawer interface {
	NewEntry(length int, text string)
	EndEntry(length int)
	NewSection()
	NewColumn(length int, text string)
	NewLine(text string)
	EndLine()
}

type VerbosityOpts struct {
	Output bool
	Task   bool
}

type Runner struct {
	opts VerbosityOpts
	p    *printer
}

func New(draw Drawer, opts VerbosityOpts) *Runner {
	return &Runner{
		opts: opts,
		p: &printer{
			drawer:  draw,
			depth:   0,
			verbose: opts,
		},
	}
}

func (r *Runner) Run(ctx context.Context, prov Iterator) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	iter, err := prov.Iter()
	if err != nil {
		return err
	}

	for {
		tsk, ok := iter.Next()
		if !ok {
			break
		}

		r.p.next()

		if rnr, ok := tsk.(rosie.Executor); ok {
			out, err := rnr.Exec(ctx)
			r.p.logBefore(tsk)

			if err != nil {
				r.p.logAfter(tsk)
				return err
			}

			if err := r.p.drain(out); err != nil {
				r.p.logAfter(tsk)
				return err
			}
			r.p.logAfter(tsk)
		} else {
			r.p.logBefore(tsk)
		}
	}

	r.p.drawer.EndEntry(0)

	return nil
}
func Run(ctx context.Context, w io.Writer, prov Iterator, ver VerbosityOpts) error {
	r := New(&draw.Drawer{
		W: w,
	}, ver)
	return r.Run(ctx, prov)
}

type printer struct {
	drawer                  Drawer
	depth                   int
	verbose                 VerbosityOpts
	start                   time.Time
	combout                 *bytes.Buffer
	openSection, openHeader bool
	previous                rosie.Joint
}

func (p *printer) logBefore(tsk rosie.Joint) {
	footer := func() {
		p.drawer.NewSection()
		text := "\033[92m\u2713\033[0m ok"
		p.drawer.NewLine(text)
		p.drawer.NewColumn(draw.Width-2-utf8.RuneCountInString(text), fmt.Sprintf("\u23F1  %s", time.Since(p.start).String()))
		p.drawer.EndLine()
		p.start = time.Time{}
	}
	header := func() {
		p.drawer.NewEntry(p.depth, fmt.Sprintf("\u2022 %s", lightYellow(tsk.Name())))
		p.start = time.Now()
		p.openSection = true
		p.openHeader = false
	}

	switch tsk.Node().Type() {
	case dag.TypeMiddleBeginning, dag.TypeBeginning:
		if p.verbose.Task {
			if !p.start.IsZero() {
				footer()
			}
			header()
		}
		p.depth += 1
	case dag.TypeMiddleEnd:
		p.depth -= 1
		p.openHeader = true
	case dag.TypeHidden:
		return
	default:
		if p.openHeader {
			if !p.start.IsZero() {
				footer()
			}
			header()
		}
		if rnr, ok := tsk.(rosie.Executor); ok {
			if !p.verbose.Task {
				return
			}
			if p.openSection {
				p.drawer.NewSection()
				p.openSection = false
			}
			if rnr.Result().Err() == nil {
				p.drawer.NewLine(fmt.Sprintf("\u203A %s", tsk.Name()))
			} else {
				p.drawer.NewLine(fmt.Sprintf("\033[91m\u2717\033[0m %s", tsk.Name()))
			}
			if desc := tsk.(interface{ Desc() string }).Desc(); desc != "" {
				p.drawer.NewColumn(0, ": ")
				p.drawer.NewColumn(0, blue(desc))
			}

			p.drawer.EndLine()
		}
	}
}

func (p *printer) logAfter(tsk rosie.Joint) {
	defer func() {
		p.previous = tsk
	}()
	if rnr, ok := tsk.(rosie.Executor); ok {
		if rnr.Result().Err() != nil {
			p.drawer.NewSection()
			p.drawer.NewLine(fmt.Sprintf("\u23F1  %s", time.Since(p.start).String()))
			p.drawer.NewColumn(draw.Width, fmt.Sprintf("\033[91m\u2717\033[0m failure with error: %s", rnr.Result().Err()))
			p.drawer.EndLine()
			return
		}
	}
}

func (p *printer) drain(in <-chan rosie.Piece) error {
	start := false
	for piece := range in {
		if piece.Err != nil {
			return piece.Err
		}
		if p.verbose.Output {
			if !start {
				p.drawer.NewLine("  output:")
				p.drawer.EndLine()
				start = true
			}
			p.drawer.NewLine("  " + gray(piece.Text))
			p.drawer.EndLine()
		}
	}

	return nil
}

func (p *printer) next() {
	p.combout = &bytes.Buffer{}
}

func gray(s string) string {
	return fmt.Sprintf("\033[2m%s\033[0m", s)
}

func blue(s string) string {
	return fmt.Sprintf("\033[34m%s\033[0m", s)
}

func lightYellow(s string) string {
	return fmt.Sprintf("\033[93m%s\033[0m", s)
}
