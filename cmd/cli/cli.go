package main

import (
	"log"
	"os"

	"github.com/ddouglas/monocle/cron"
	"github.com/ddouglas/monocle/hack"
	"github.com/ddouglas/monocle/processor"
	"github.com/ddouglas/monocle/server"
	"github.com/urfave/cli"
)

// go:generate sqlboiler mysql --wipe --struct-tag-casing camel --no-tests

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

var port = cli.UintFlag{
	Name:  "port",
	Value: 8080,
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

var save = cli.BoolFlag{
	Name: "save",
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
			Name:      "processor",
			Category:  "Population",
			Usage:     "processor",
			UsageText: "processes records within the database and populates the database with new records from the API",
			Action:    processor.Action,
			Flags: []cli.Flag{
				scope, workers, records, sleep, begin, done,
			},
		},
		// cli.Command{
		// 	Name:      "counter",
		// 	Category:  "Monitoring",
		// 	Usage:     "counter",
		// 	UsageText: "Continuous Loop that runs a query every few seconds and returns a count of expired etags",
		// 	Flags: []cli.Flag{
		// 		sleep, expired, save,
		// 	},
		// 	Action: updater.Counter,
		// },
		cli.Command{
			Name:      "api",
			Category:  "HTTP",
			Usage:     "api",
			UsageText: "Starts the API to serve HTTPS requests",
			Flags: []cli.Flag{
				port,
			},
			Action: server.Serve,
		},
		cli.Command{
			Name:      "cron",
			Category:  "Scheduled",
			Usage:     "cron",
			UsageText: "Run GoCron Implmentation",
			Action:    cron.Action,
		},
		cli.Command{
			Name:     "hack",
			Category: "Hacking",
			Action:   hack.Action,
		},
	}

}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
