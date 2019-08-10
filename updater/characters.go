package updater

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

func (u *Updater) evaluateCharacters(sleep, threshold int) error {

	for {
		u.Logger.DebugF("Current Error Count: %d Remain: %d", u.ESI.Remain, u.ESI.Reset)
		if u.ESI.Remain < 20 {
			u.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", u.ESI.Reset)
			time.Sleep(time.Second * time.Duration(u.ESI.Reset))
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

		u.Logger.Infof("Successfully Queried %d Characters", len(characters))

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
	if u.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", u.ESI.Reset)
		u.Logger.Errorf("\t%s", msg)
		msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		u.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(u.ESI.Reset))
	}
	attempts := 0
	for {
		if attempts >= 3 {
			u.Logger.Errorf("All Attempts exhuasted for Character %d", character.ID)
			return
		}
		response, err = u.ESI.GetCharactersCharacterID(character)
		if err != nil {
			u.Logger.Errorf("Error completing request to ESI for Character %d information: %s", character.ID, err)
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		u.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting %d request again in 1 second", response.Code, response.Path, attempts)
		time.Sleep(1 * time.Second)
	}

	character = response.Data.(monocle.Character)

	_, err = u.DB.UpdateCharacterByID(character)
	if err != nil {
		u.Logger.Errorf("Error Encountered attempting to update character %d in database: %s", character.ID, err)
		return
	}

	return
}

func (u *Updater) updateCharacterCorpHistory(character monocle.Character) {
	var newEtag bool
	var history []monocle.CharacterCorporationHistory
	var response esi.Response

	if u.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", u.ESI.Reset)
		u.Logger.Errorf("\t%s", msg)
		msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		u.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(u.ESI.Reset))
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
			u.Logger.Errorf("All Attempts exhuasted for Character %d", character.ID)
			return
		}
		response, historyEtag, err = u.ESI.GetCharactersCharacterIDCorporationHistory(historyEtag)
		if err != nil {
			u.Logger.Errorf("Error completing request to ESI for Character %d Corporation History: %s", historyEtag.ID, err)
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		u.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting %d request again in 1 second", response.Code, response.Path, attempts)
		time.Sleep(1 * time.Second)
	}

	history = response.Data.([]monocle.CharacterCorporationHistory)

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
