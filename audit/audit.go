package audit

import (
	"log"
	"sync"

	"github.com/ddouglas/monocle/core"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Auditor struct {
	*core.App
}

var (
	workers uint64
	records uint64
	sleep   uint64
	begin   uint64
	done    uint64
	scope   string
	wg      sync.WaitGroup
)

func Action(c *cli.Context) error {
	core, err := core.New("auditor")
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
	}

	a := Auditor{
		core,
	}

	scope = c.String("scope")
	workers = c.Uint64("workers")
	records = c.Uint64("records")
	begin = c.Uint64("begin")
	done = c.Uint64("done")
	sleep = c.Uint64("sleep")

	a.Logger.WithField("workers", workers).Info("starting auditor")

	switch scope {
	case "charUpdater":
		a.charUpdater()
	default:
		return cli.NewExitError(errors.New("scope not specified"), 1)
	}

	return nil
}
