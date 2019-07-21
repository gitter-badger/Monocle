package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/apsdehal/go-logger"
	"github.com/ddouglas/eveindex"
	"github.com/ddouglas/eveindex/esi"
	"github.com/ddouglas/eveindex/mysql"
	"github.com/kelseyhightower/envconfig"
)

var err error
var wg sync.WaitGroup
var logit *logger.Logger
var db *mysql.DB
var esiClient *esi.Client
var e420d bool // Error 420d

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
	logit.SetFormat("#%{id} %{time} %{file}:%{line} => %{lvl}\t%{message}")
	logit.SetLogLevel(logger.InfoLevel)

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

	var where = map[string]interface{}{
		"etag": "",
	}

	var perPage = 250

	for {
		time.Sleep(time.Millisecond * 500)
		if e420d {
			logit.Error("We got 420d")
			time.Sleep(time.Second * 60)
		}
		for x := 1; x <= 10; x++ {
			characters, err := db.SelectCharacters(x, perPage, where)
			if err != nil {
				if err != sql.ErrNoRows {
					logit.Fatalf("Unabel to query for characters: %s", err)
				}
				continue
			}

			logit.Infof("Queried %d characters", len(characters))
			if len(characters) == 0 {
				continue
			}
			wg.Add(1)
			go processCharacters(characters)

		}
		logit.Info("Routines Launched. Waiting for completion")
		wg.Wait()
		logit.Info("All Routines done")

	}

	logit.Info("Done")

}

func processCharacters(characters []eveindex.Character) {
	defer wg.Done()
	for _, character := range characters {
		processCharacter(character.ID)
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
	if !character.IsExpired() {
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
	case 420:
		logit.Error("420 Response Code received from ESI")
		e420d = true
		return
	default:
		logit.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	logit.Debugf("\tCharacter: %d:%s\tNew Character: %t", character.ID, character.Name, newCharacter)

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

	if !corporation.IsExpired() {
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

	logit.Debugf("\tCorporation: %d:%s\tNew Corporation: %t", corporation.ID, corporation.Name, newCorporation)

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

	if !alliance.IsExpired() {
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

	logit.Debugf("\tAlliance: %d:%s\tNew Alliance: %t", alliance.ID, alliance.Name, newAlliance)

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
