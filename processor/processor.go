package processor

import (
	"log"
	"sync"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Processor struct {
	*core.App
}

type EtagResource struct {
	model  *monocle.EtagResource
	exists bool
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
	core, err := core.New("processor")
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
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

	p.Logger.WithField("workers", workers).Info("starting processor")

	switch scope {
	case "charHunter":
		p.charHunter()
	case "charUpdater":
		p.charUpdater()
	case "corpHunter":
		p.corpHunter()
	case "corpUpdater":
		p.corpUpdater()
	case "alliHunter":
		p.alliHunter()
	case "alliUpdater":
		p.alliUpdater()
	default:
		return cli.NewExitError(errors.New("scope not specified"), 1)
	}

	return nil
}

func (p *Processor) EvaluateESIArtifacts() {

	if p.ESI.Remain < 20 {
		p.Logger.WithFields(logrus.Fields{
			"errors":    p.ESI.Remain,
			"remaining": p.ESI.Reset,
		}).Error("error count is low. sleeping...")
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

}

func (p *Processor) SleepDuringDowntime(t time.Time) {

	const lower = 39300 // 10:55 UTC 10h * 3600s + 55m * 60s
	const upper = 42900 // 11:25 UTC 11h * 3600s + 25m * 60s

	hm := (t.Hour() * 3600) + (t.Minute() * 60)

	if hm >= lower && hm <= upper {
		p.Logger.Info("Entering Sleep Phase for Downtime")
		time.Sleep(time.Second * time.Duration(upper-hm))
	}
}
