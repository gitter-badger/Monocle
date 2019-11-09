package main

import (
	"log"
	"os"

	"github.com/ddouglas/monocle/audit"
	"github.com/ddouglas/monocle/cron"
	"github.com/ddouglas/monocle/hack"
	"github.com/ddouglas/monocle/processor"
	"github.com/ddouglas/monocle/server"
	"github.com/urfave/cli"
)

// go:generate sqlboiler  --wipe --struct-tag-casing camel --no-tests mysql

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

var page = cli.IntFlag{
	Name: "page",
}

var end = cli.IntFlag{
	Name: "end",
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
	Value: 5,
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
		cli.Command{
			Name:      "auditor",
			Category:  "Population",
			Usage:     "auditor",
			UsageText: "processes records within the database and populates the database with new records from the API",
			Action:    audit.Action,
			Flags: []cli.Flag{
				scope, workers, records, sleep, begin, done, page, end,
			},
		},
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
