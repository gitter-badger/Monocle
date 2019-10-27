package core

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/kelseyhightower/envconfig"

	"github.com/apsdehal/go-logger"
	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
	"github.com/ddouglas/monocle/mysql"
)

type (
	App struct {
		ESI    *esi.Client
		DB     *mysql.DB
		DGO    *discordgo.Session
		Logger *logger.Logger
	}
)

func New() (*App, error) {
	var config monocle.Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal("unable initialize environment variables")
	}

	logging, err := logger.New("monocle-core", 1, os.Stdout)
	if err != nil {
		log.Fatal("Unable to create app lication logger")
		os.Exit(1)
	}

	logging.SetFormat("#%{id} %{time} %{file}:%{line} => %{lvl} %{message}")

	switch config.LogLevel {
	case 1:
		logging.SetLogLevel(logger.CriticalLevel)
	case 2:
		logging.SetLogLevel(logger.ErrorLevel)
	case 3:
		logging.SetLogLevel(logger.WarningLevel)
	case 4:
		logging.SetLogLevel(logger.NoticeLevel)
	case 5:
		logging.SetLogLevel(logger.InfoLevel)
	case 6:
		logging.SetLogLevel(logger.DebugLevel)
	default:
		log.Println("Logging Level Not Set. Defaulting to Info")
		logging.SetLogLevel(logger.InfoLevel)
	}

	connection, err := mysql.Connect()
	if err != nil {
		logging.Fatalf("Encoutered Error Attempting to setup DB Connection: %s", err)
	}

	esiClient, err := esi.New()
	if err != nil {
		logging.Fatalf("Encoutered Error Attempting to setup ESI Client: %s", err)
	}

	token := fmt.Sprintf("Bot %s", config.DiscordToken)
	discord, err := discordgo.New(token)
	if err != nil {
		logging.Fatalf("Encoutered Error Attempting to setup Discord Go: %s", err)
	}

	discord.LogLevel = discordgo.LogDebug

	return &App{
		ESI:    esiClient,
		DB:     connection,
		DGO:    discord,
		Logger: logging,
	}, nil

}
