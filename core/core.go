package core

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

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
		Logger *logrus.Entry
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
		logger.WithError(err).Fatal("failed to configure log level for logrus")
	}

	logger.SetLevel(level)

	logger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyMsg: "message",
		},
	})

	logger.AddHook(NewGraylogLogrusHook(config.GraylogURL, config.GraylogPort))

	connection, err := mysql.Connect()
	if err != nil {
		logger.WithError(err).Fatal("unable to setup db connection")
	}

	esiClient, err := esi.New()
	if err != nil {
		logger.WithError(err).Fatal("unable to setup esi client")
	}

	token := fmt.Sprintf("Bot %s", config.DiscordToken)
	discord, err := discordgo.New(token)
	if err != nil {
		logger.WithError(err).Fatal("unable to setup discord client")
	}

	discord.LogLevel = discordgo.LogDebug

	host, err := os.Hostname()
	if err != nil {
		logger.WithError(err).Fatal("unable to determine hostname of host")

	}

	entry := logger.WithFields(logrus.Fields{
		"host": host,
	})

	return &App{
		ESI:    esiClient,
		DB:     connection,
		DGO:    discord,
		Logger: entry,
	}, nil

}

type GraylogLogrusHook struct {
	glef GlefClient
}

type GlefClient struct {
	client *http.Client
	host   url.URL
}

func NewGraylogLogrusHook(host string, port uint) *GraylogLogrusHook {
	return &GraylogLogrusHook{
		glef: GlefClient{
			client: &http.Client{
				Timeout: 30 * time.Second,
			},
			host: url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("%s:%d", host, port),
				Path:   "/gelf",
			},
		},
	}
}

func (h *GraylogLogrusHook) Fire(e *logrus.Entry) error {

	l, err := e.String()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, h.glef.host.String(), bytes.NewBufferString(l))
	if err != nil {
		return err
	}

	rep, err := h.glef.client.Do(req)
	if err != nil {
		return err
	}

	if rep.StatusCode > 202 {
		defer rep.Body.Close()
		bbody, err := ioutil.ReadAll(rep.Body)
		if err != nil {
			return errors.New("unable to read error response from response")
		}

		return fmt.Errorf("Unexcpexted Request Code received: %d Body: %s", rep.StatusCode, string(bbody))
	}

	return nil
}

func (h *GraylogLogrusHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
