package updater

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/core"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func Counter(c *cli.Context) error {

	core, err := core.New()
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
		return cli.NewExitError(err, 1)
	}

	var charDiff, corpDiff, alliDiff string
	var charCount, charLast uint
	var corpLast, corpCount uint
	var alliLast, alliCount uint

	sleep := c.Int("sleep")
	expired := c.Bool("expired")
	save := c.Bool("save")
	for {
		if expired {
			charCount, err = core.DB.SelectCountOfExpiredCharacterEtags()
		} else {
			charCount, err = core.DB.SelectCountOfCharacterEtags()
		}
		if err != nil {
			if err != sql.ErrNoRows {
				core.Logger.Fatalf("Unable to query for characters: %s", err)
			}
		}

		if expired {
			corpCount, err = core.DB.SelectCountOfExpiredCorporationEtags()
		} else {
			corpCount, err = core.DB.SelectCountOfCorporationEtags()
		}
		if err != nil {
			if err != sql.ErrNoRows {
				core.Logger.Fatalf("Unable to query for characters: %s", err)
			}
		}

		if expired {
			alliCount, err = core.DB.SelectCountOfExpiredAllianceEtags()
		} else {
			alliCount, err = core.DB.SelectCountOfAllianceEtags()
		}
		if err != nil {
			if err != sql.ErrNoRows {
				core.Logger.Fatalf("Unable to query for characters: %s", err)
			}
		}

		if charCount > charLast {
			charDiff = fmt.Sprintf("+%d", charCount-charLast)
		} else if charCount < charLast {
			charDiff = fmt.Sprintf("-%d", charLast-charCount)
		} else {
			charDiff = fmt.Sprintf("=%d", 0)
		}

		if corpCount > corpLast {
			corpDiff = fmt.Sprintf("+%d", corpCount-corpLast)
		} else if corpCount < corpLast {
			corpDiff = fmt.Sprintf("-%d", corpLast-corpCount)
		} else {
			corpDiff = fmt.Sprintf("=%d", 0)
		}

		if alliCount > alliLast {
			alliDiff = fmt.Sprintf("+%d", alliCount-alliLast)
		} else if alliCount < alliLast {
			alliDiff = fmt.Sprintf("-%d", alliLast-alliCount)
		} else {
			alliDiff = fmt.Sprintf("=%d", 0)
		}

		core.Logger.Infof("Char: %d (%s)\tCorp: %d (%s)\tAlli: %d (%s)", charCount, charDiff, corpCount, corpDiff, alliCount, alliDiff)

		if save {
			err := core.DB.InsertCounter(monocle.Counter{
				CharCount: charCount,
				CorpCount: corpCount,
				AlliCount: alliCount,
			})
			if err != nil {
				core.Logger.Error(err.Error())
			}
		}

		charLast = charCount
		corpLast = corpCount
		alliLast = alliCount

		time.Sleep(time.Second * time.Duration(sleep))

	}

}
