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
	scope string
	wg    sync.WaitGroup
)

func Action(c *cli.Context) error {
	core, err := core.New("auditor")
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
	}

	scope = c.String("scope")

	a := &Auditor{
		core,
	}

	switch scope {
	case "charUpdater":
		a.charUpdater(c)
	// case "corpUpdater":
	// 	a.corpUpdater(c)
	default:
		return cli.NewExitError(errors.New("scope not specified"), 1)
	}

	return nil
}
