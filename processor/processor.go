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
	workers    uint64
	threshold  uint64
	errorCount uint64
	records    uint64
	sleep      uint64
	begin      uint64
	done       uint64
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
	workers = c.Uint64("workers")
	records = c.Uint64("records")
	begin = c.Uint64("begin")

	done = c.Uint64("done")
	sleep = c.Uint64("sleep")

	p.Logger.Infof("Starting process with %d workers", workers)

	switch scope {
	case "charHunter":
		p.charHunter()
	case "charUpdater":
		p.charUpdater()
	case "corpHunter":
		p.corpHunter()
	case "corpUpdater":
		p.corpUpdater()
	default:
		return cli.NewExitError(errors.New("scope not specified"), 1)
	}

	return nil
}
