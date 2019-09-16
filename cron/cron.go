package cron

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/volatiletech/sqlboiler/boil"

	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/core"
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
	msg := "Registering Count func with GoCron"
	h.Logger.Info(msg)
	h.SendDicoMsg(msg)
	gocron.Every(1).Hour().Do(h.Counts)

	msg = "Registering Deltas func with GoCron"
	h.Logger.Info(msg)
	h.SendDicoMsg(msg)
	gocron.Every(1).Day().At("11:00").Do(h.Deltas)

	msg = "Start GoCron"
	h.Logger.Info(msg)
	h.SendDicoMsg(msg)
	<-gocron.Start()

	return nil

}

func (h *Handler) SendDicoMsg(s string) {

	// msg := fmt.Sprintf("<@!277968564827324416> %s", s)
	h.DGO.ChannelMessageSend("394991263344230411", s)
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

func (h *Handler) Counts() {

	msg := "Starting Count Logger"
	h.Logger.Info(msg)
	h.SendDicoMsg(msg)
	query := `
		SELECT (
			SELECT COUNT(*) FROM characters WHERE ignored = 0
		) AS characters, (
			SELECT COUNT(*) FROM corporations WHERE closed = 0 AND ignored = 0
		) AS corporations
	`

	var counts struct {
		Character    uint64 `db:"characters" json:"characters"`
		Corporations uint64 `db:"corporations" json:"corporations"`
	}

	err := h.DB.GetContext(context.Background(), &counts, query)

	if err != nil {
		h.Logger.Error(err.Error())
		return
	}

	var kv boiler.KV
	data, err := json.Marshal(counts)
	if err != nil {
		h.Logger.Error(err.Error())
		return
	}

	kv.K = "current_table_counts"
	kv.V = data
	kv.CreatedAt = time.Now()

	err = kv.Upsert(
		context.Background(),
		h.DB,
		boil.Whitelist(
			boiler.KVColumns.V,
			boiler.KVColumns.UpdatedAt,
		), boil.Infer())

	if err != nil {
		h.Logger.Error(err.Error())
		return
	}

	msg = "Finishing Count Logger"
	h.Logger.Info(msg)
	h.SendDicoMsg(msg)

}
