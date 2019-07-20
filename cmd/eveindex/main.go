package main

import (
	"encoding/json"

	"net/url"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/apsdehal/go-logger"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
)

var err error
var wg sync.WaitGroup
var logit *logger.Logger

type Config struct {
	DBDriver string `envconfig:"DB_DRIVER" required:"true"`
	DBHost   string `envconfig:"DB_HOST" required:"true"`
	DBPort   string `envconfig:"DB_PORT" required:"true"`
	DBName   string `envconfig:"DB_NAME" required:"true"`
	DBUser   string `envconfig:"DB_USER" required:"true"`
	DBPass   string `envconfig:"DB_PASS" required:"true"`
}

func main() {
	var config Config
	err = envconfig.Process("INDEX", &config)
	if err != nil {
		logit.Fatalf("Encoutered Error Attempting to scan environvent variables: %s", err)
	}

	logit, err = logger.New("eveindex-ws", 1, os.Stdout)
	logit.SetFormat("#%{id} %{time} %{file}:%{line} => %{lvl} %{message}")
	logit.SetLogLevel(logger.InfoLevel)

	wg.Add(1)

	logit.Infof("Number of Go Routines: %d", runtime.NumGoroutine())

	go supervisor()
	logit.Info("Waiting for supervisor to die")

	wg.Wait()
	logit.Info("Bye")

}

func supervisor() {
	connected := make(chan bool, 10)
	disconnected := make(chan bool, 10)
	done := make(chan bool)
	defer func() {
		wg.Done()
	}()

	logit.Infof("Number of Go Routines: %d", runtime.NumGoroutine())
	wg.Add(1)
	go listen(connected, disconnected, done)

	logit.Infof("Number of Go Routines: %d", runtime.NumGoroutine())

	for {
		select {
		case <-done:
			logit.Info("Done in Supervisor")
			logit.Infof("Number of Go Routines Remaining: %d", runtime.NumGoroutine())
			return
		case <-disconnected:
			logit.Infof("Supervisor: Disconnected for Websocket. Attempting to reconnect")
			time.Sleep(2 * time.Second)
			wg.Add(1)
			go listen(connected, disconnected, done)
		case <-connected:
			logit.Info("Supervisor: Connected to Websocket")
		}
	}
}

func listen(connected, disconnected, done chan bool) {

	logit.Infof("Number of Go Routines: %d", runtime.NumGoroutine())

	defer func() {
		if r := recover(); r != nil {
			logit.Infof("Recovered in f %s", r)
		}
		wg.Done()
		return
	}()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	address := url.URL{
		Scheme: "wss",
		Host:   "zkillboard.com:2096",
	}

	subMsg := struct {
		Action  string `json:"action"`
		Channel string `json:"channel"`
	}{
		Action:  "sub",
		Channel: "all:*",
	}

	msg, err := json.Marshal(subMsg)
	if err != nil {
		logit.Infof("Encoutered Error Attempting marshal sub message: %s", err)
		return
	}

	logit.Infof("Connecting to %s", address.String())

	c, _, err := websocket.DefaultDialer.Dial(address.String(), nil)
	if err != nil {
		logit.Panicf("dial: %s", err)
	}
	logit.Infof("Connected to %s", address.String())

	logit.Infof("Sending Sub Message: %s", msg)

	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		logit.Infof("Encoutered Error Attempting to scan environvent variables: %s", err)
		return
	}

	connected <- true

	defer func() {
		c.Close()
	}()
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		logit.Infof("Number of Go Routines: %d", runtime.NumGoroutine())

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logit.Infof("Read: %s", err)
				return
			}

			logit.Infof("Received: %s", message)
		}
	}()

	// ticker := time.NewTicker(time.Second)

	for {
		select {
		// case t := <-ticker.C:
		// 	err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
		// 	if err != nil {
		// 		logit.Info("write:", err)
		// 		return
		// 	}
		case <-interrupt:
			logit.Info("Interrupted")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logit.Infof("Failed to write close message: %s", err)
				return
			}
			done <- true
			return
		}
	}
}
