package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"net/url"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/apsdehal/go-logger"
	"github.com/ddouglas/eveindex/esi"
	"github.com/ddouglas/eveindex/mysql"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
)

var err error
var wg sync.WaitGroup
var logit *logger.Logger
var db *mysql.DB
var esiClient *esi.Client

type Config struct {
	DBDriver string `envconfig:"DB_DRIVER" required:"true"`
	DBHost   string `envconfig:"DB_HOST" required:"true"`
	DBPort   string `envconfig:"DB_PORT" required:"true"`
	DBName   string `envconfig:"DB_NAME" required:"true"`
	DBUser   string `envconfig:"DB_USER" required:"true"`
	DBPass   string `envconfig:"DB_PASS" required:"true"`
}

func main() {
	logit, err = logger.New("eveindex-ws", 1, os.Stdout)
	logit.SetFormat("#%{id} %{time} %{file}:%{line} => %{lvl} %{message}")
	logit.SetLogLevel(logger.DebugLevel)

	var config Config
	err = envconfig.Process("INDEX", &config)
	if err != nil {
		logit.Fatalf("Encoutered Error Attempting to scan environvent variables: %s", err)
	}
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName)
	db, err = mysql.Connect(mysqlDSN)
	if err != nil {
		logit.Fatalf("Encoutered Error Attempting to setup DB Connection: %s", err)
	}

	esiClient, err = esi.New("INDEX")
	if err != nil {
		logit.Fatalf("Encoutered Error Attempting to set ESI Client: %s", err)
	}

	wg.Add(1)
	go supervisor()
	logit.Info("Waiting for supervisor to die")

	wg.Wait()
	logit.Info("Bye")

}

func supervisor() {
	connected := make(chan bool, 10)
	disconnected := make(chan bool, 10)
	done := make(chan bool)
	stream := make(chan []byte)

	defer func() {
		wg.Done()
	}()

	wg.Add(1)
	go listen(stream, connected, disconnected, done)

	for {
		select {
		case kill := <-stream:
			wg.Add(1)
			go processStream(kill)
		case <-done:
			logit.Info("Done in Supervisor")
			logit.Infof("Number of Go Routines Remaining: %d", runtime.NumGoroutine())
			return
		case <-disconnected:
			logit.Infof("Supervisor: Disconnected from Websocket. Attempting to reconnect")
			time.Sleep(2 * time.Second)
			wg.Add(1)
			go listen(stream, connected, disconnected, done)
		case <-connected:
			logit.Info("Supervisor: Connected to Websocket")
		}
	}
}

func listen(stream chan []byte, connected, disconnected, done chan bool) {

	defer func() {
		if r := recover(); r != nil {
			logit.Infof("Recovered in f %s", r)
			disconnected <- true
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

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				err, ok := err.(*websocket.CloseError)
				if ok {
					code := err.Code
					logit.Infof("Error Code: %d", code)
					if code == 1000 {
						return
					}
					disconnected <- true
					logit.Info("Pushed True boolean on to Disconnected Chan")
				}
				return
			}

			stream <- message

		}
	}()

	for {
		select {

		case <-interrupt:
			logit.Info("Interrupted")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logit.Errorf("Failed to write close message: %s", err)
			}

			done <- true
			return
		}
	}
}

type LittleKill struct {
	Action        string `json:"action"`
	KillID        uint   `json:"killID"`
	CharacterID   uint64 `json:"character_id"`
	CorporationID uint   `json:"corporation_id"`
	AllianceID    uint   `json:"alliance_id"`
	ShipTypeID    uint   `json:"ship_type_id"`
	URL           string `json:"url"`
	Hash          string `json:"hash"`
}

func processStream(kill []byte) {
	defer wg.Done()

	var killmail LittleKill
	err = json.Unmarshal(kill, &killmail)
	if err != nil {
		logit.ErrorF("Unable to unmarshel kill into struct: %s", kill)
		return
	}
	logit.Debugf("\tReceived: %d:%s", killmail.KillID, killmail.Hash)

	if killmail.CharacterID > 0 {
		processCharacter(killmail.CharacterID)
	}

	if killmail.CorporationID > 0 {
		processCorporation(killmail.CorporationID)
	}

	if killmail.AllianceID > 0 {
		processAlliance(killmail.AllianceID)
	}

	return

}

func processCharacter(id uint64) {

	var newCharacter bool

	character, err := db.SelectCharacterByCharacterID(id)
	if err != nil {
		if err != sql.ErrNoRows {
			logit.Errorf("DB Query for Character ID %d Failed with Error %s", id, err)
			return
		}
		character.ID = id
		newCharacter = true
	}
	logit.Debugf("\tCharacter: %d:%s\tNew Character: %t\tCharacter Expiration: %s\tCharacter Expired: %t", character.ID, character.Name, newCharacter, character.Expires, character.IsExpired())
	if !character.IsExpired() {
		logit.Debugf("\tSkipping Character: %d", character.ID)
		return
	}

	response, err := esiClient.GetCharactersCharacterID(character.ID, character.Etag)
	if err != nil {
		logit.Errorf("Error completing request to ESI for Character information: %s", err)
		return
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &character)
		if err != nil {
			logit.Errorf("unable to unmarshel response body: %s", err)
			return
		}
		expires, err := retreiveExpiresHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}

		etag, err := retrieveEtagHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		character.Etag = etag

		character.Expires = expires
		break
	case 304:
		expires, err := retreiveExpiresHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}
		character.Expires = expires

		etag, err := retrieveEtagHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		character.Etag = etag

		break
	default:
		logit.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	switch newCharacter {
	case true:
		_, err := db.InsertCharacter(character)
		if err != nil {
			logit.Errorf("Error Encountered attempting to insert new character into database: %s", err)
			return
		}
	case false:
		_, err := db.UpdateCharacterByID(character)
		if err != nil {
			logit.Errorf("Error Encountered attempting to update character in database: %s", err)
			return
		}
	}
}

func processCorporation(id uint) {

	var newCorporation bool

	corporation, err := db.SelectCorporationByCorporationID(id)
	if err != nil {
		if err != sql.ErrNoRows {
			logit.Errorf("DB Query for Corporation ID %d Failed with Error %s", id, err)
			return
		}
		corporation.ID = id
		newCorporation = true
	}

	logit.Debugf("\tCorporation: %d:%s\tNew Corporation: %t\tCorporation Expiration: %s\tCorporation Expired: %t", corporation.ID, corporation.Name, newCorporation, corporation.Expires, corporation.IsExpired())

	if !corporation.IsExpired() {
		logit.Debugf("\tSkipping Corporation: %d", corporation.ID)
		return
	}

	response, err := esiClient.GetCorporationsCorporationID(corporation.ID, corporation.Etag)
	if err != nil {
		logit.Errorf("Error completing request to ESI for Character information: %s", err)
		return
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &corporation)
		if err != nil {
			logit.Errorf("unable to unmarshel response body: %s", err)
			return
		}
		expires, err := retreiveExpiresHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}
		corporation.Expires = expires

		etag, err := retrieveEtagHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		corporation.Etag = etag

		break
	case 304:
		expires, err := retreiveExpiresHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}

		etag, err := retrieveEtagHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		corporation.Etag = etag

		corporation.Expires = expires
		break
	default:
		logit.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	switch newCorporation {
	case true:
		_, err := db.InsertCorporation(corporation)
		if err != nil {
			logit.Errorf("Error Encountered attempting to insert new corporation into database: %s", err)
			return
		}
	case false:
		_, err := db.UpdateCorporationByID(corporation)
		if err != nil {
			logit.Errorf("Error Encountered attempting to update corporation in database: %s", err)
			return
		}
	}
}

func processAlliance(id uint) {
	var newAlliance bool

	alliance, err := db.SelectAllianceByAllianceID(id)
	if err != nil {
		if err != sql.ErrNoRows {
			logit.Errorf("DB Query for Alliance ID %d Failed with Error %s", id, err)
			return
		}
		alliance.ID = id
		newAlliance = true
	}

	logit.Debugf("\tAlliance: %d:%s\tNew Alliance: %t\tAlliance Expiration: %s\tAlliance Expired: %t", alliance.ID, alliance.Name, newAlliance, alliance.Expires, alliance.IsExpired())

	if !alliance.IsExpired() {
		logit.Debugf("\tSkipping Alliance: %d", alliance.ID)
		return
	}

	response, err := esiClient.GetAlliancesAllianceID(alliance.ID, alliance.Etag)
	if err != nil {
		logit.Errorf("Error completing request to ESI for Alliance information: %s", err)
		return
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &alliance)
		if err != nil {
			logit.Errorf("unable to unmarshel response body: %s", err)
			return
		}

		expires, err := retreiveExpiresHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}

		alliance.Expires = expires

		etag, err := retrieveEtagHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		alliance.Etag = etag
		break
	case 304:
		expires, err := retreiveExpiresHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}

		alliance.Expires = expires

		etag, err := retrieveEtagHeaderFromResponse(response)
		if err != nil {
			logit.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		alliance.Etag = etag
		break
	default:
		logit.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	switch newAlliance {
	case true:
		_, err := db.InsertAlliance(alliance)
		if err != nil {
			logit.Errorf("Error Encountered attempting to insert new alliance into database: %s", err)
			return
		}
	case false:
		_, err := db.UpdateAllianceByID(alliance)
		if err != nil {
			logit.Errorf("Error Encountered attempting to update alliance in database: %s", err)
			return
		}
	}
}

func retreiveExpiresHeaderFromResponse(response esi.Response) (time.Time, error) {
	if _, ok := response.Headers["Expires"]; !ok {
		err = fmt.Errorf("Expires Headers is missing for url %s", response.Path)
		return time.Time{}, err
	}
	return time.Parse(esi.LayoutESI, response.Headers["Expires"])
}

func retrieveEtagHeaderFromResponse(response esi.Response) (string, error) {
	if _, ok := response.Headers["Etag"]; !ok {
		err = fmt.Errorf("Etag Header is missing from url %s", response.Path)
		return "", err
	}
	return response.Headers["Etag"], nil
}
