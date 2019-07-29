package main

import (
	"log"
	"os"

	"github.com/ddouglas/monocle/populate"
	"github.com/ddouglas/monocle/updater"
	"github.com/ddouglas/monocle/websocket"
	"github.com/urfave/cli"
)

// var defPort = cli.UintFlag{
// 	Name:  "port",
// 	Usage: "Custom Port numer for the API to listen on",
// 	Value: 8000,
// }

var app *cli.App
var scope = cli.StringFlag{
	Name:  "scope",
	Usage: "Defines the Scope of this execution (i.e. Characters, Corporations, etc)",
}

var workers = cli.IntFlag{
	Name:  "workers",
	Usage: "Defines the number of GoRoutines that can be inflight at a single time",
	Value: 10,
}
var records = cli.IntFlag{
	Name:  "records",
	Usage: "Defines the number of records that we attempt to pull from the database at one time. Must be a multiple of Workers",
	Value: 250,
}
var begin = cli.IntFlag{
	Name: "begin",
}
var done = cli.IntFlag{
	Name:  "done",
	Value: 98000000,
}
var sleep = cli.IntFlag{
	Name:  "sleep",
	Usage: "The number of seconds the loop should sleep for",
	Value: 10,
}
var expired = cli.BoolFlag{
	Name:  "expired",
	Usage: "Limit to query to just expired records or count all records in the database",
}
var threshold = cli.IntFlag{
	Name:  "threshold",
	Usage: "The minimum number of records that should be returned from the query in order for the job to run",
	Value: 100,
}

func init() {
	app = cli.NewApp()
	app.Name = "monocle Core"
	app.Usage = "Core Application of monocle Backend. Responsibile for managing Websocket Connections and running background tasks"
	app.Version = "0.0.2"
	app.Commands = []cli.Command{
		cli.Command{
			Name:      "listen",
			Category:  "Websocket",
			Usage:     "listen",
			UsageText: "Opens a WSS connection to the zKillboard Websocket and Listens to the connection for new killmails. Unknown Killmail Victims are checked against the database and saved if unknown",
			Action:    websocket.Start,
		},
		cli.Command{
			Name:      "update",
			Category:  "Something IDK",
			Usage:     "update",
			UsageText: "Monitors Database for Expired Etags and makes an HTTP Request to ESI to check to see if there are any updates",
			Flags: []cli.Flag{
				scope, workers, records, sleep, threshold,
			},
			Action: updater.Process,
		},
		cli.Command{
			Name:      "populate",
			Category:  "Population",
			Usage:     "populate",
			UsageText: "Checks in with ESI looking for new entities starting with Alliances",
			Action:    populate.Action,
			Flags: []cli.Flag{
				scope, workers, records, sleep, begin, done,
			},
		},
		cli.Command{
			Name:      "counter",
			Category:  "Monitoring",
			Usage:     "counter",
			UsageText: "Continuous Loop that runs a query every few seconds and returns a count of expired etags",
			Flags: []cli.Flag{
				sleep, expired,
			},
			Action: updater.Counter,
		},
	}

}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
