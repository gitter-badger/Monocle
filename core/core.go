package core

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"

	"github.com/apsdehal/go-logger"
	"github.com/ddouglas/monocle/esi"
	"github.com/ddouglas/monocle/evewho"
	"github.com/ddouglas/monocle/mysql"
	"github.com/kelseyhightower/envconfig"
)

var err error

type (
	App struct {
		Config Config
		ESI    *esi.Client
		Who    *evewho.Client
		DB     *mysql.DB
		DGO    *discordgo.Session
		Logger *logger.Logger
	}

	Config struct {
		DBDriver string `envconfig:"DB_DRIVER" required:"true"`
		DBHost   string `envconfig:"DB_HOST" required:"true"`
		DBPort   string `envconfig:"DB_PORT" required:"true"`
		DBName   string `envconfig:"DB_NAME" required:"true"`
		DBUser   string `envconfig:"DB_USER" required:"true"`
		DBPass   string `envconfig:"DB_PASS" required:"true"`

		LogLevel uint `envconfig:"LOG_LEVEL" required:"true"`

		DiscordToken string `envconfig:"DISCORD_TOKEN" required:"true"`

		PopulateBegin int `envconfig:"POPULATE_BEGIN"`

		// HttpServerPort string `envconfig:"HTTP_SERVER_PORT" required:"true"`
	}
)

func New() (*App, error) {
	var config Config
	err = envconfig.Process("MONOCLE", &config)
	if err != nil {
		log.Fatalf("Unable to scan environment variables into the application: %s", err)
		os.Exit(1)
	}

	logging, err := logger.New("monocle-core", 1, os.Stdout)
	if err != nil {
		log.Fatal("Unable to create application logger")
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

	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName)
	db, err := mysql.Connect(mysqlDSN)
	if err != nil {
		logging.Fatalf("Encoutered Error Attempting to setup DB Connection: %s", err)
	}

	esiClient, err := esi.New("monocle")
	if err != nil {
		logging.Fatalf("Encoutered Error Attempting to setup ESI Client: %s", err)
	}

	evewhoClient, err := evewho.New("monocle")
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
		Config: config,
		ESI:    esiClient,
		Who:    evewhoClient,
		DB:     db,
		DGO:    discord,
		Logger: logging,
	}, nil
}
