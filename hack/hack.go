package hack

import (
	"context"
	"fmt"
	"strings"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/core"
	"github.com/urfave/cli"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func Action(c *cli.Context) error {

	core, err := core.New()
	if err != nil {
		return err
	}

	page := 1
	limit := 10000

	for {
		var characters []*monocle.Character
		offset := (page * limit) - limit

		core.Logger.Infof("Starting Page: %d Offset: %d", page, offset)

		err := boiler.Characters(
			qm.Where(boiler.CharacterColumns.CorporationID+"=?", 0),
			qm.Limit(limit),
			qm.Offset(offset),
		).Bind(context.Background(), core.DB, &characters)
		if err != nil {
			return err
		}

		if len(characters) == 0 {
			break
		}

		var ids []interface{}
		var q []string
		for _, v := range characters {
			q = append(q, "?")
			ids = append(ids, v.ID)
		}

		qStr := strings.Join(q, ", ")

		query := `
			UPDATE 
				characters
			SET
				etag = "",
				ignored = 0,
				expires = NOW()
			WHERE id IN (%s)
		`

		query = fmt.Sprintf(query, qStr)

		_, err = core.DB.Exec(query, ids...)
		if err != nil {
			return err
		}

		core.Logger.Infof("Finished with Page: %d Offset: %d", page, offset)
	}

	core.Logger.Infof("Script Done")
	return nil

}
