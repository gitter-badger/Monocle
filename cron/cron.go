package cron

import (
	"log"

	"github.com/ddouglas/monocle/core"
	"github.com/jasonlvhit/gocron"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func Action(c *cli.Context) error {

	core, err := core.New()
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
		return cli.NewExitError(err, 1)
	}

	gocron.Every(5).Seconds().Do(func() {
		core.DGO.ChannelMessageSend("394991263344230411", "Hello There from GoCron")
	})

	<-gocron.Start()

	return nil

}
