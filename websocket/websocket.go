package websocket

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/core"
)

var err error
var wg sync.WaitGroup

type (
	Listener struct {
		*core.App
	}
	LittleKill struct {
		Action        string `json:"action"`
		KillID        uint   `json:"killID"`
		CharacterID   uint64 `json:"character_id"`
		CorporationID uint   `json:"corporation_id"`
		AllianceID    uint   `json:"alliance_id"`
		ShipTypeID    uint   `json:"ship_type_id"`
		URL           string `json:"url"`
		Hash          string `json:"hash"`
	}
)

func Start(c *cli.Context) error {
	core, err := core.New()
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
		return cli.NewExitError(err, 1)
	}

	listener := Listener{
		core,
	}

	wg.Add(1)
	go listener.supervisor()
	core.Logger.Info("Waiting for supervisor to die")

	wg.Wait()
	core.Logger.Info("Bye")
	return nil
}

func (l *Listener) supervisor() {
	connected := make(chan bool, 10)
	disconnected := make(chan bool, 10)
	done := make(chan bool)
	stream := make(chan []byte)

	defer func() {
		wg.Done()
	}()

	wg.Add(1)
	go l.listen(stream, connected, disconnected, done)

	for {
		select {
		case kill := <-stream:
			wg.Add(1)
			go l.processStream(kill)
		case <-done:
			l.Logger.Info("Done in Supervisor")
			l.Logger.Infof("Number of Go Routines Remaining: %d", runtime.NumGoroutine())
			return
		case <-disconnected:
			msg := fmt.Sprint("Supervisor: Disconnected from Websocket. Attempting to reconnect")
			l.Logger.Error(msg)
			msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
			go func(msg string) {
				_, _ = l.DGO.ChannelMessageSend("394991263344230411", msg)
				return
			}(msg)
			time.Sleep(2 * time.Second)
			wg.Add(1)
			go l.listen(stream, connected, disconnected, done)
		case <-connected:
			l.Logger.Info("Supervisor: Connected to Websocket")
		}
	}
}

func (l *Listener) listen(stream chan []byte, connected, disconnected, done chan bool) {

	defer func() {
		if r := recover(); r != nil {
			l.Logger.Infof("Recovered in f %s", r)
			disconnected <- true
		}

		return
	}()
	defer wg.Done()

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
		l.Logger.Infof("Encoutered Error Attempting marshal sub message: %s", err)
		return
	}

	l.Logger.Infof("Connecting to %s", address.String())

	c, _, err := gorilla.DefaultDialer.Dial(address.String(), nil)
	if err != nil {
		l.Logger.Panicf("dial: %s", err)
	}
	l.Logger.Infof("Connected to %s", address.String())

	l.Logger.Infof("Sending Sub Message: %s", msg)

	err = c.WriteMessage(gorilla.TextMessage, msg)
	if err != nil {
		l.Logger.Infof("Encoutered Error Attempting to scan environvent variables: %s", err)
		return
	}

	connected <- true

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				err, ok := err.(*gorilla.CloseError)
				if ok {
					code := err.Code
					l.Logger.Infof("Error Code: %d", code)
					if code == 1000 {
						return
					}
					disconnected <- true
					l.Logger.Info("Pushed True boolean on to Disconnected Chan")
				}
				return
			}

			stream <- message

		}
	}()

	ticker := time.NewTicker(time.Second * 10)

	for {
		select {
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			l.Logger.Info("Interrupted")
			err := c.WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, ""))
			if err != nil {
				l.Logger.Errorf("Failed to write close message: %s", err)
			}

			done <- true
			return
		}
	}
}

func (l *Listener) processStream(kill []byte) {
	defer wg.Done()

	var killmail LittleKill
	err = json.Unmarshal(kill, &killmail)
	if err != nil {
		l.Logger.ErrorF("Unable to unmarshel kill into struct: %s", kill)
		return
	}
	l.Logger.Infof("\tReceived: %d:%s", killmail.KillID, killmail.Hash)

	if killmail.CharacterID > 0 {
		l.processCharacter(killmail.CharacterID)
	}

	if killmail.CorporationID > 0 {
		l.processCorporation(uint64(killmail.CorporationID))
	}

	if killmail.AllianceID > 0 {
		l.processAlliance(uint64(killmail.AllianceID))
	}

	return

}

func (l *Listener) processCharacter(id uint64) {

	var newCharacter bool

	character, err := l.DB.SelectCharacterByCharacterID(id)
	if err != nil {
		if err != sql.ErrNoRows {
			l.Logger.Errorf("DB Query for Character ID %d Failed with Error %s", id, err)
			return
		}
		character.ID = id
		newCharacter = true
	}
	if !character.IsExpired() {
		return
	}

	response, err := l.ESI.GetCharactersCharacterID(character)
	if err != nil {
		l.Logger.Errorf(err.Error())
		return
	}

	character = response.Data.(monocle.Character)

	l.Logger.Debugf("\tCharacter: %d:%s\tNew Character: %t", character.ID, character.Name, newCharacter)

	switch newCharacter {
	case true:
		_, err := l.DB.InsertCharacter(character)
		if err != nil {
			l.Logger.Errorf("Error Encountered attempting to insert new character into database: %s", err)
			return
		}
	case false:
		_, err := l.DB.UpdateCharacterByID(character)
		if err != nil {
			l.Logger.Errorf("Error Encountered attempting to update character in database: %s", err)
			return
		}
	}
}

func (l *Listener) processCorporation(id uint64) {

	var newCorporation bool

	corporation, err := l.DB.SelectCorporationByCorporationID(id)
	if err != nil {
		if err != sql.ErrNoRows {
			l.Logger.Errorf("DB Query for Corporation ID %d Failed with Error %s", id, err)
			return
		}
		corporation.ID = id
		newCorporation = true
	}

	if !corporation.IsExpired() {
		return
	}

	response, err := l.ESI.GetCorporationsCorporationID(corporation)
	if err != nil {
		l.Logger.Errorf("Error completing request to ESI for Character information: %s", err)
		return
	}

	corporation = response.Data.(monocle.Corporation)

	l.Logger.Debugf("\tCorporation: %d:%s\tNew Corporation: %t", corporation.ID, corporation.Name, newCorporation)

	switch newCorporation {
	case true:
		_, err := l.DB.InsertCorporation(corporation)
		if err != nil {
			l.Logger.Errorf("Error Encountered attempting to insert new corporation into database: %s", err)
			return
		}
	case false:
		_, err := l.DB.UpdateCorporationByID(corporation)
		if err != nil {
			l.Logger.Errorf("Error Encountered attempting to update corporation in database: %s", err)
			return
		}
	}
}

func (l *Listener) processAlliance(id uint64) {
	var newAlliance bool

	alliance, err := l.DB.SelectAllianceByAllianceID(id)
	if err != nil {
		if err != sql.ErrNoRows {
			l.Logger.Errorf("DB Query for Alliance ID %d Failed with Error %s", id, err)
			return
		}
		alliance.ID = id
		newAlliance = true
	}

	if !alliance.IsExpired() {
		return
	}

	response, err := l.ESI.GetAlliancesAllianceID(alliance)
	if err != nil {
		l.Logger.Errorf("Error completing request to ESI for Alliance information: %s", err)
		return
	}

	alliance = response.Data.(monocle.Alliance)

	l.Logger.Debugf("\tAlliance: %d:%s\tNew Alliance: %t", alliance.ID, alliance.Name, newAlliance)

	switch newAlliance {
	case true:
		_, err := l.DB.InsertAlliance(alliance)
		if err != nil {
			l.Logger.Errorf("Error Encountered attempting to insert new alliance into database: %s", err)
			return
		}
	case false:
		_, err := l.DB.UpdateAllianceByID(alliance)
		if err != nil {
			l.Logger.Errorf("Error Encountered attempting to update alliance in database: %s", err)
			return
		}
	}
}
