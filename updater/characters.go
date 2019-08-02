package updater

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

func (u *Updater) evaluateCharacters(sleep, threshold int) error {

	for {
		u.Logger.DebugF("Current Error Count: %d Remain: %d", u.count, u.reset)
		if u.count < 20 {
			u.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", u.reset)
			time.Sleep(time.Second * time.Duration(u.reset))
		}

		characters, err := u.DB.SelectExpiredCharacterEtags(records * workers)
		if err != nil {
			if err != sql.ErrNoRows {
				u.Logger.Fatalf("Unable to query for characters: %s", err)
			}
			continue
		}

		if len(characters) < threshold {
			u.Logger.Infof("Minimum threshold of %d for job not met. Sleeping for %d seconds", threshold, sleep)
			time.Sleep(time.Second * time.Duration(sleep))
			continue
		}

		charChunk := chunkCharacterSlice(records, characters)

		for _, characters := range charChunk {
			wg.Add(1)
			go func(characters []monocle.Character) {
				for _, character := range characters {
					u.updateCharacter(character)
					u.updateCharacterCorpHistory(character)
				}
				wg.Done()
			}(characters)
		}

		u.Logger.Info("Waiting")
		wg.Wait()
		u.Logger.Info("Done")
	}

}

func (u *Updater) updateCharacter(character monocle.Character) {

	var response esi.Response

	u.Logger.DebugF("Updating Character %d", character.ID)
	if u.count < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", u.reset)
		u.Logger.Errorf("\t%s", msg)
		msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		u.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(u.reset))
	}
	attempts := 0
	for {
		if attempts >= 3 {
			break
		}
		response, err = u.ESI.GetCharactersCharacterID(character.ID, character.Etag)
		if err != nil {
			u.Logger.Errorf("Error completing request to ESI for Character %d information: %s", character.ID, err)
			return
		}

		mx.Lock()
		u.reset = esi.RetrieveErrorResetFromResponse(response)
		u.count = esi.RetrieveErrorCountFromResponse(response)
		mx.Unlock()

		if response.Code < 500 {
			break
		}
		u.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting request again in 1 second", response.Code, response.Path)

		time.Sleep(1 * time.Second)
		attempts++
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &character)
		if err != nil {
			u.Logger.Errorf("unable to unmarshel response body for %d: %s", character.ID, err)
			return
		}
		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
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
	case 304:
		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
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

func (u *Updater) updateCharacterCorpHistory(character monocle.Character) {
	var newEtag bool
	var history []monocle.CharacterCorporationHistory
	var response esi.Response

	u.Logger.Debugf("Updating Character %d Corporation History", character.ID)

	if u.count < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", u.reset)
		u.Logger.Errorf("\t%s", msg)
		msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		u.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(u.reset))
	}

	historyEtag, err := u.DB.SelectEtagByIDAndResource(character.ID, "character_corporation_history")
	if err != nil {
		if err != sql.ErrNoRows {
			u.Logger.Errorf("Unable to query character_corporation_history etag resource for Character %d due to SQL Error: %s", character.ID, err)
			return
		}

		newEtag = true
		historyEtag.ID = character.ID
		historyEtag.Resource = "character_corporation_history"
	}

	attempts := 0
	for {
		if attempts >= 3 {
			break
		}
		response, err = u.ESI.GetCharactersCharacterIDCorporationHistory(historyEtag.ID, historyEtag.Etag)
		if err != nil {
			u.Logger.Errorf("Error completing request to ESI for Character %d Corporation History: %s", historyEtag.ID, err)
			return
		}

		mx.Lock()
		u.reset = esi.RetrieveErrorResetFromResponse(response)
		u.count = esi.RetrieveErrorCountFromResponse(response)
		mx.Unlock()

		if response.Code < 500 {
			break
		}

		u.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting request again in 1 second", response.Code, response.Path)

		time.Sleep(1 * time.Second)
		attempts++
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &history)
		if err != nil {
			u.Logger.Errorf("unable to unmarshel response body for %d corporation history: %s", historyEtag.ID, err)
			return
		}

		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to parse expires header for url %s: %s", response.Path, err)
		}
		historyEtag.Expires = expires

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to retrieve etag header for url %s: %s", response.Path, err)
		}
		historyEtag.Etag = etag

		break
	case 304:
		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to parse expires header for url %s: %s", response.Path, err)
		}
		historyEtag.Expires = expires

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to retrieve etag header for url %s: %s", response.Path, err)
		}
		historyEtag.Etag = etag

		break
	default:
		u.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	existing, err := u.DB.SelectCharacterCorporationHistoryByID(historyEtag.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			u.Logger.Errorf("Unable to query character_corporation_history etag resource for Character %d due to SQL Error: %s", character.ID, err)
			return
		}
	}

	diff := diffExistingHistory(existing, history)
	switch newEtag {
	case true:
		_, err := u.DB.InsertEtag(historyEtag)
		if err != nil {
			u.Logger.Errorf("error encountered attempting to insert new etag for history into database: %s", err)
			return
		}
	case false:
		_, err := u.DB.UpdateEtagByIDAndResource(historyEtag)
		if err != nil {
			u.Logger.Errorf("error encountered attempting to insert new etag for history into database: %s", err)
			return
		}
	}

	if len(diff) > 0 {
		_, err := u.DB.InsertCharacterCorporationHistory(historyEtag.ID, diff)
		if err != nil {
			u.Logger.Errorf("error encountered attempting to insert new character corporation history records into database: %s", err)
			return
		}
	}
	return

}

func diffExistingHistory(a []monocle.CharacterCorporationHistory, b []monocle.CharacterCorporationHistory) []monocle.CharacterCorporationHistory {
	c := convertToMap(a)
	d := convertToMap(b)
	result := make([]monocle.CharacterCorporationHistory, 0)
	for recordID, history := range d {
		if _, ok := c[recordID]; !ok {
			result = append(result, history)
		}
	}

	return result
}

func convertToMap(a []monocle.CharacterCorporationHistory) map[uint]monocle.CharacterCorporationHistory {
	result := make(map[uint]monocle.CharacterCorporationHistory, 0)
	for _, history := range a {
		result[history.RecordID] = history
	}
	return result
}
