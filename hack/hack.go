package hack

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/core"
	"github.com/urfave/cli"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type Hack struct {
	*core.App
}

func Action(c *cli.Context) error {

	core, err := core.New("hack")
	if err != nil {
		log.Fatal(err)
		return nil
	}

	ctx := context.Background()

	go func() {
		for {
			count, err := boiler.Corporations(
				qm.Where("expires > NOW()"),
			).Count(ctx, core.DB)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("[cor][%s] - Count: %d\n", time.Now().Format("2006-01-02 15:04:05"), count)
			time.Sleep(time.Second * 30)
		}

	}()

	go func() {
		for {

			t := time.Now().UTC()

			spew.Dump(t)
			// h := t.Hour()
			// m := t.Minute()

			// fmt.Printf("[time][%s] Hour: %d\tMinute: %d\n", t.Format("2006-01-02 15:04:05"), h, m)
			time.Sleep(time.Second * 1)
		}

	}()

	var fake chan bool

	<-fake

	return nil

}
