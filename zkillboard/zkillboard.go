package zkillboard

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

/**
Websocket client using a basic supervision pattern, so that in the event that the websocket connection breaks, the
system schedules another connection.
*/

type Session struct {
	Address            url.URL
	Reconnect          chan bool
	Done               chan bool
	Close              chan bool
	Term               chan os.Signal
	ReconnectOnFailure bool
	WaitGroup          sync.WaitGroup
	WsConn             *websocket.Conn
	Main               chan []byte
}

type LittleKill struct {
	ID   uint32 `json:"killid"`
	Hash string `json:"hash"`
}

func New() *Session {
	return &Session{
		Address: url.URL{
			Scheme: "wss",
			Host:   "zkillboard.com:2096",
		},
		Done:               make(chan bool),
		Close:              make(chan bool),
		Term:               make(chan os.Signal),
		ReconnectOnFailure: true,
	}
}

func (s *Session) Start() {
	defer func() {
		s.WaitGroup.Done()
	}()
	s.WaitGroup.Add(1)
	logit.Info("Before Listen")
	go s.listen()
	logit.Info("After Listen")

	//wait for events to occur.

	for {
		select {
		case <-s.Done:
			return
		case <-s.Close:

			logit.Infof("Supervisor: Connection Closed \n")

			logit.Infof("Supervisor: Reconnecting \n")
			time.Sleep(2 * time.Second)
			s.WaitGroup.Add(1)
			go s.listen()
		}
	}
}

func (s *Session) listen() {
	done := make(chan bool)

	defer func() {
		if r := recover(); r != nil {
			fmt.Info("Recovered in defer function of Listen()", r)
		}

		select {
		case <-s.Reconnect:
			s.WaitGroup.Add(1)
			go s.listen()
		}

		return
	}()

	signal.Notify(s.Term, os.Interrupt)

	logit.Infof("Connecting to %s", s.Address.String())

	c, _, err := websocket.DefaultDialer.Dial(s.Address.String(), nil)
	if err != nil {
		logit.Panicf("dial: %s", err)
	}
	logit.Infof("Connected to %s", s.Address.String())
	s.WsConn = c

	s.WaitGroup.Add(1)
	go func() {
		defer close(done)
		defer s.WaitGroup.Done()
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				s.Reconnect <- true
				logit.Info("read: ", err)
				return
			}
			logit.Infof("ReadRoutine: Following Message Received. Push to Main Chan: %s\n\n", msg)
			s.Main <- msg
		}
	}()

	for {
		select {
		case <-done:
			logit.Info("ListenRoutine: Done, exiting ")
			return
		case <-s.Term:
			logit.Info("ListenRoutine: SIGTERM, exiting ")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logit.Info("write close:", err)
			}

			s.WaitGroup.Done()
			return
		}
	}
}
