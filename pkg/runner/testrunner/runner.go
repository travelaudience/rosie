package testrunner

import (
	"context"
	"testing"
	"time"

	"github.com/travelaudience/rosie"
)

type iteratorProvider interface {
	Iter() (*rosie.Iterator, error)
}

func Run(t *testing.T, prov iteratorProvider, assert func(*testing.T, error)) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	iter, err := prov.Iter()
	if err != nil {
		t.Fatal(err)
	}

	for {
		tsk, ok := iter.Next()
		if !ok {
			break
		}

		if rnr, ok := tsk.(rosie.Executor); ok {
			out, err := rnr.Exec(ctx)
			if err != nil {
				assert(t, err)
			}

			drain(t, out, assert)
		}
	}
}

func drain(t *testing.T, in <-chan rosie.Piece, assert func(*testing.T, error)) {
	for piece := range in {
		if piece.Err != nil {
			assert(t, piece.Err)
		}

		t.Log(piece.Text)
	}
}
