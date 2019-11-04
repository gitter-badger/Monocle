package core

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
	"github.com/ddouglas/monocle/mysql"
)

type (
	App struct {
		ESI    *esi.Client
		DB     *mysql.DB
		DGO    *discordgo.Session
		Logger *logrus.Logger
	}
)

func New(name string) (*App, error) {
	var config monocle.Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal("unable initialize environment variables")
	}
	// logDir := "core/logs"
	// if _, err := os.Stat(logDir); os.IsNotExist(err) {
	// 	err = os.Mkdir(logDir, os.ModeDir)
	// 	if err != nil {
	// 		log.Fatal("unable make log directory")
	// 	}
	// }

	var logger = logrus.New()
	// if name == "" {
	// 	name = "unknown"
	// }
	// name = fmt.Sprintf("%s/%s", logDir, name)
	// logFileName := fmt.Sprintf("%s_%s.log", name, time.Now().Format("2006-01-02-15-04"))
	// logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// mw := io.MultiWriter(os.Stdout, logFile)
	logger.SetOutput(os.Stdout)

	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		logger.WithField("err", err).Fatal("failed to configure log level for logrus")
	}

	logger.SetLevel(level)

	logger.SetFormatter(&logrus.JSONFormatter{})
	// logger.SetFormatter(&logrus.TextFormatter{
	// 	DisableColors: false,
	// 	FullTimestamp: true,
	// })

	connection, err := mysql.Connect()
	if err != nil {
		logger.WithField("err", err).Fatal("unable to setup db connection")
	}

	esiClient, err := esi.New()
	if err != nil {
		logger.WithField("err", err).Fatal("unable to setup esi client")
	}

	token := fmt.Sprintf("Bot %s", config.DiscordToken)
	discord, err := discordgo.New(token)
	if err != nil {
		logger.WithField("err", err).Fatal("unable to setup discord client")
	}

	discord.LogLevel = discordgo.LogDebug

	return &App{
		ESI:    esiClient,
		DB:     connection,
		DGO:    discord,
		Logger: logger,
	}, nil

}
