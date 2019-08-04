package populate

import (
	"log"
	"sync"

	"github.com/ddouglas/monocle/core"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Populator struct {
	*core.App
}

var (
	workers,
	// threshold,
	errorCount,
	records int
	sleep       int
	begin, done int
	scope       string
	wg          sync.WaitGroup
)

func Action(c *cli.Context) error {
	core, err := core.New()
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
		return cli.NewExitError(err, 1)
	}

	populator := Populator{
		core,
	}

	scope = c.String("scope")
	workers = c.Int("workers")
	records = c.Int("records")
	begin = c.Int("begin")

	done = c.Int("done")
	sleep = c.Int("sleep")

	populator.Logger.Infof("Starting process with %d workers", workers)

	switch scope {
	case "getAlliancelList":
		_ = populator.getAlliancelList()
	case "getAllianceCorpMemberList":
		_ = populator.getAllianceCorpList()
	case "getAllianceCharMemberList":
		_ = populator.getAllianceCharList()
	case "charHunter":
		_ = populator.charHunter()
		// case "corpHunter":
		// 	_ = populator.corpHunter()
		// case "alliHunter":
		// 	_ = populator.alliHunter()

	}

	return nil
}
