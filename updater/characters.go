package updater

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
	"github.com/pkg/errors"
)

func (u *Updater) evaluateCharacters(sleep, threshold int) error {

	var errorCount int

	for {
		if errorCount >= 10 {
			return errors.New("Error Count High. Exiting Programs")
		}
		for x := 1; x <= workers; x++ {
			characters, err := u.DB.SelectExpiredCharacterEtags(x, records)
			if err != nil {
				errorCount++
				if err != sql.ErrNoRows {
					u.Logger.Fatalf("Unable to query for characters: %s", err)
				}
				continue
			}

			if len(characters) <= threshold {
				u.Logger.Infof("Minimum threshold of %d for job not met. Sleeping for %d seconds", threshold, sleep)
				time.Sleep(time.Second * time.Duration(sleep))
				continue
			}

			u.Logger.Infof("Successfully Queried %d Characters", len(characters))
			wg.Add(1)
			go u.updateCharacters(characters)
		}
		u.Logger.Info("Waiting")
		wg.Wait()
		u.Logger.Info("Done")
	}

}

func (u *Updater) updateCharacters(characters []monocle.Character) {
	defer wg.Done()
	for _, character := range characters {
		if !character.IsExpired() {
			continue
		}
		u.updateCharacter(character)
	}
	return
}

func (u *Updater) updateCharacter(character monocle.Character) {
	u.Logger.DebugF("Updating Character %s", character.ID)

	response, err := u.ESI.GetCharactersCharacterID(character.ID, character.Etag)
	if err != nil {
		u.Logger.Errorf("Error completing request to ESI for Character %d information: %s", character.ID, err)
		return
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &character)
		if err != nil {
			u.Logger.Errorf("unable to unmarshel response body for %d: %s", character.ID, err)
			return
		}
		expires, err := esi.RetreiveExpiresHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to parse expires header for url %s: %s", response.Path, err)
		}

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to retrieve etag header for url %s: %s", response.Path, err)
		}
		character.Etag = etag

		character.Expires = expires
		break
	case 304:
		expires, err := esi.RetreiveExpiresHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to parse expires header for url %s: %s", response.Path, err)
		}

		character.Expires = expires

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to retrieve etag header for url %s: %s", response.Path, err)
		}
		character.Etag = etag
		break
	default:
		u.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	_, err = u.DB.UpdateCharacterByID(character)
	if err != nil {
		u.Logger.Errorf("Error Encountered attempting to update character in database: %s", err)
		return
	}

	return
}
