package hack

import (
	"sync"
	"time"

	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/core"
	"github.com/urfave/cli"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

var wg sync.WaitGroup

func Action(c *cli.Context) error {

	app, err := core.New()
	if err != nil {
		return err
	}
	page := 1
	limit := 10000

	for {
		var ids []uint64
		offset := (page * limit) - limit

		app.Logger.Infof("Starting Page: %d Offset: %d", page, offset)

		charQuery := boiler.NewQuery(
			qm.Select("id"),
			qm.From("characters"),
			qm.Where(boiler.CharacterColumns.CorporationID+"=?", 0),
			qm.Limit(limit),
			qm.Offset(offset),
		)

		queryStr, args := queries.BuildQuery(charQuery)
		app.Logger.Infof("Executing Query: %s", queryStr)

		err := app.DB.Select(&ids, queryStr, args...)
		if err != nil {
			return err
		}

		length := len(ids)
		if length == 0 {
			break
		}

		app.Logger.Infof("Successfully Queried %d Characters", length)

		chunks := chunkUint64Slice(1000, ids)

		for _, chunk := range chunks {
			wg.Add(1)
			go func(core *core.App, chunk []uint64) {
				for _, id := range chunk {
					// core.Logger.Infof("Processing ID %d", id)
					_, err := core.DB.Exec(`
					UPDATE characters SET etag = "", ignored = 0, expires = NOW() WHERE id = ?
					`, id)
					if err != nil {
						core.Logger.Error(err.Error())
					}
				}
				wg.Done()
			}(app, chunk)
		}
		app.Logger.Info("Waiting")
		wg.Wait()
		app.Logger.Info("Done. Sleeping for 1 second")
		time.Sleep(time.Second * 1)
	}

	return nil

}

func chunkUint64Slice(size int, slice []uint64) [][]uint64 {
	chunk := make([][]uint64, 0)

	if len(slice) <= size {
		chunk = append(chunk, slice)
		return chunk
	}

	for x := 0; x <= len(slice); x += size {

		end := x + size

		if end > len(slice) {
			end = len(slice)
		}

		chunk = append(chunk, slice[x:end])

	}

	return chunk
}
