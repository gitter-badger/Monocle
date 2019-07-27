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
	var charLast = 0
	var corpLast = 0
	var alliLast = 0
	var charDiff, corpDiff, alliDiff string
	var charCount, corpCount, alliCount monocle.Counter

	sleep := c.Int("sleep")
	expired := c.Bool("expired")
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

		if charCount.Count > charLast {
			charDiff = fmt.Sprintf("↑%d", charCount.Count-charLast)
		} else if charCount.Count < charLast {
			charDiff = fmt.Sprintf("↓%d", charLast-charCount.Count)
		} else {
			charDiff = fmt.Sprintf("↔%d", 0)
		}

		if corpCount.Count > corpLast {
			corpDiff = fmt.Sprintf("↑%d", corpCount.Count-corpLast)
		} else if corpCount.Count < corpLast {
			corpDiff = fmt.Sprintf("↓%d", corpLast-corpCount.Count)
		} else {
			corpDiff = fmt.Sprintf("↔%d", 0)
		}

		if alliCount.Count > alliLast {
			alliDiff = fmt.Sprintf("↑%d", alliCount.Count-alliLast)
		} else if alliCount.Count < alliLast {
			alliDiff = fmt.Sprintf("↓%d", alliLast-alliCount.Count)
		} else {
			alliDiff = fmt.Sprintf("↔%d", 0)
		}

		core.Logger.Infof("Char: %d (%s)\tCorp: %d (%s)\tAlli: %d (%s)", charCount.Count, charDiff, corpCount.Count, corpDiff, alliCount.Count, alliDiff)

		charLast = charCount.Count
		corpLast = corpCount.Count
		alliLast = alliCount.Count

		time.Sleep(time.Second * time.Duration(sleep))

	}

}
