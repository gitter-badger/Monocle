package updater

import (
	"log"
	"sync"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/core"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type (
	Updater struct {
		*core.App
	}
)

var (
	workers,
	sleep, threshold,
	records int
	scope string
	wg    sync.WaitGroup
	err   error
)

func Process(c *cli.Context) error {

	core, err := core.New()
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
		return cli.NewExitError(err, 1)
	}

	updater := Updater{
		core,
	}

	workers = c.Int("workers")
	records = c.Int("records")
	sleep = c.Int("sleep")
	threshold = c.Int("threshold")
	scope = c.String("scope")

	if records <= threshold {
		err = errors.New("Records cannot be less than or equal to threshold")
		return cli.NewExitError(err, 1)
	}

	core.Logger.Infof("Starting Updates with the Scope of %s utilizing %d workers and assigning %d records to each worker", scope, workers, records)

	switch scope {
	case "characters":
		_ = updater.evaluateCharacters(sleep, threshold)
	case "corporations":
		_ = updater.evaluateCorporations(sleep, threshold)
	case "alliances":
		_ = updater.evaluateAlliances(sleep, threshold)
	default:
		return cli.NewExitError(errors.New("scope not specified"), 1)
	}

	return nil

}

func chunkCharacterSlice(size int, slice []monocle.Character) [][]monocle.Character {

	var chunk [][]monocle.Character
	chunk = make([][]monocle.Character, 0)

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
