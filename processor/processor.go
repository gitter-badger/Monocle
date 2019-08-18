package processor

import (
	"log"
	"sync"

	"github.com/ddouglas/monocle/core"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Processor struct {
	*core.App
}

var (
	workers    int
	threshold  int
	errorCount int
	records    int
	sleep      int
	begin      int
	done       int
	scope      string
	wg         sync.WaitGroup
)

func Action(c *cli.Context) error {
	core, err := core.New()
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
		return cli.NewExitError(err, 1)
	}

	p := Processor{
		core,
	}

	scope = c.String("scope")
	workers = c.Int("workers")
	records = c.Int("records")
	begin = c.Int("begin")

	done = c.Int("done")
	sleep = c.Int("sleep")

	p.Logger.Infof("Starting process with %d workers", workers)

	switch scope {

	}

	return nil
}
