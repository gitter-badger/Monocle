package cron

import (
	"fmt"
	"log"
	"sync"

	"github.com/robfig/cron/v3"

	"github.com/ddouglas/monocle/core"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Handler struct {
	*core.App
}

var wg sync.WaitGroup

func Action(c *cli.Context) error {

	core, err := core.New("cron")
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
		return cli.NewExitError(err, 1)
	}

	h := Handler{
		core,
	}
	crn := cron.New()
	wg.Add(1)

	countCronSig := "0 */2 * * *"
	msg := fmt.Sprintf("Registering Count func with Cron Sig of \n\t`%s`\n", countCronSig)
	crn.AddFunc(countCronSig, h.Counts)

	deltaCronSig := "0 11 * * *"
	msg = fmt.Sprintf("%s\nRegistering Deltas func with Cron Sig of \n\t`%s`\n", msg, deltaCronSig)
	crn.AddFunc(deltaCronSig, h.Deltas)

	msg = fmt.Sprintf("%s\nStarting Cron", msg)
	h.Logger.Info(msg)
	h.SendDicoMsg(msg)
	crn.Start()
	wg.Wait()
	return nil

}

func (h *Handler) SendDicoMsg(s string) {

	// s := fmt.Sprintf("<@!277968564827324416> %s", s)
	h.DGO.ChannelMessageSend("394991263344230411", s)
}

func (h *Handler) Deltas() {

	msg := "Starting Deltas Logger"
	h.Logger.Info(msg)
	h.SendDicoMsg(msg)
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

	msg = "Finish Deltas Logger"
	h.Logger.Info(msg)
	h.SendDicoMsg(msg)
}

func (h *Handler) Counts() {

	msg := "Starting Count Logger"
	h.Logger.Info(msg)
	h.SendDicoMsg(msg)

	query := `
		INSERT INTO totals (
			characters,
			corporations,
			alliances,
			created_at
		) SELECT (
			SELECT COUNT(*) FROM characters WHERE ignored = 0
		) AS characters, (
			SELECT COUNT(*) FROM corporations WHERE closed = 0 AND ignored = 0
		) AS corporations, (
			SELECT COUNT(*) FROM alliances WHERE closed = 0 AND ignored = 0
		) AS alliances,
		NOW()
	`

	_, err := h.DB.Exec(query)
	if err != nil {
		h.Logger.Error(err.Error())
		return
	}

	msg = "Finishing Count Logger"
	h.Logger.Info(msg)
	h.SendDicoMsg(msg)

}
