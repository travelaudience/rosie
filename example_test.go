package rosie_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/travelaudience/rosie/pkg/runner/clirunner"

	. "github.com/travelaudience/rosie"
)

func Example() {
	var count int

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	group := Group("test")
	group.Beginning().
		Then(MakeDir("tmp")).
		Then(Cmd("test", "go", "test",
			"-race",
			"-count", "2",
			"-coverprofile", "./tmp/cover.out",
			"-covermode", "atomic",
			"-run", "TestCmd_optimistic",
			"./...",
		)).
		Then(Fn("count", StringSliceClosure(func(_ context.Context, _ io.Writer, res []string) (interface{}, error) {
			count = len(res)
			return nil, nil
		})))

	if err := clirunner.Run(ctx, os.Stdout, group, clirunner.VerbosityOpts{
		// Enable for debugging purposes.
		//Output: testing.Verbose(),
		//Task:   testing.Verbose(),
	}); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(count)

	// Output: 6
}
