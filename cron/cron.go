package cron

import (
	"log"

	"github.com/ddouglas/monocle/core"
	"github.com/jasonlvhit/gocron"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Handler struct {
	*core.App
}

func Action(c *cli.Context) error {

	core, err := core.New()
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
		return cli.NewExitError(err, 1)
	}

	h := Handler{
		core,
	}

	gocron.Every(5).Seconds().Do(func() {
		h.Deltas()
	})

	<-gocron.Start()

	return nil

}

func (h *Handler) Deltas() {

	query := `
		INSERT INTO corporation_deltas (
			corporation_id, 
			member_count, 
			created_at
		) SELECT 
			id, 
			member_count, 
			NOW() 
		FROM corporations 
		WHERE 
			closed = 0;
	`

	_, err := h.DB.Exec(query)
	if err != nil {
		h.Logger.Error(err.Error())
	}
	return
}
