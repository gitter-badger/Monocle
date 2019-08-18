package processor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

func (p *Processor) charHunter() {
	var value struct {
		Value int `json:"value"`
	}

	kv, err := p.DB.SelectValueByKey("last_good_character_id")
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Criticalf("Unable to query for ID: %s", err)
			return
		}
	}

	err = json.Unmarshal(kv.Value, &value)
	if err != nil {
		p.Logger.Criticalf("Unable to unmarshal value: %s", err)
		return
	}

	begin = value.Value

	p.Logger.Infof("Starting at ID %d", begin)

	for x := begin; x <= 2147483647; x += records {

		end := x + records
		msg := fmt.Sprintf("Errors: %d Remaining: %d Loop: %d - %d", p.ESI.Remain, p.ESI.Reset, x, x+records)
		p.Logger.CriticalF("%s", msg)
		attempts := 0
		for {
			if attempts >= 2 {
				msg := fmt.Sprintf("Head Requests to %d failed. Sleep for %d minutes before trying again", end, sleep)
				p.Logger.Errorf("\t%s", msg)
				// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
				// p.DGO.ChannelMessageSend("394991263344230411", msg)
				time.Sleep(time.Minute * time.Duration(sleep))
				attempts = 0
			}
			p.Logger.DebugF("Checking for valid end of %d", end)
			response, err := p.ESI.HeadCharactersCharacterID(uint64(end))
			if err != nil {
				p.Logger.ErrorF(err.Error())
				time.Sleep(time.Second * 5)
				attempts = 3
				continue
			}

			if response.Code >= 500 {
				time.Sleep(time.Second * 5)
				attempts++
				continue
			}
			break
		}

		p.Logger.Infof("%d is valid, loop from %d to %d and call ESI Character and History API", end, x, end)

		for y := x; y <= end; y++ {
			if p.ESI.Remain < 20 {
				msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
				p.Logger.Errorf("\t%s", msg)
				// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
				// p.DGO.ChannelMessageSend("394991263344230411", msg)
				time.Sleep(time.Second * time.Duration(p.ESI.Reset))
			}
			wg.Add(1)
			go func(id uint64) {
				p.processCharacter(id)
				p.processCharacterCorpHistory(id)
				wg.Done()
				return
			}(uint64(y))
		}

		p.Logger.Info("Done Dispatching. Waiting for Completion")
		wg.Wait()
		p.Logger.InfoF("Completed, sleep for %d seconds", 2)

		value.Value = end

		kv.Value, err = json.Marshal(value)
		if err != nil {
			p.Logger.Criticalf("Unable to unmarshal value: %s", err)
			return
		}

		_, err = p.DB.UpdateValueByKey(kv)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.Criticalf("Unable to query for ID: %s", err)
				return
			}
		}

		time.Sleep(time.Second * 2)

	}

	return
}

func (p *Processor) charUpdater() {

}

func (p *Processor) processCharacter(id uint64) {

	var newCharacter bool
	var response esi.Response

	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("\t%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}
	character, err := p.DB.SelectCharacterByCharacterID(id)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("DB Query for Character ID %d Failed with Error %s", id, err)
			return
		}
		character.ID = id
		newCharacter = true
	}

	if !character.IsExpired() {
		return
	}

	p.Logger.Debugf("\tProcessing Char %d", character.ID)

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.Errorf("All Attempts exhuasted for Character %d", character.ID)
			return
		}
		response, err = p.ESI.GetCharactersCharacterID(character)
		if err != nil {
			p.Logger.Errorf("Error completing request to ESI for Character %d information: %s", character.ID, err)
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		p.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting %d request again in 1 second", response.Code, response.Path, attempts)
		time.Sleep(1 * time.Second)
	}

	character = response.Data.(monocle.Character)

	p.Logger.Debugf("\tCharacter: %d:%s\tNew Character: %t", character.ID, character.Name, newCharacter)

	switch newCharacter {
	case true:
		_, err := p.DB.InsertCharacter(character)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to insert new character into database: %s", err)
			return
		}
	case false:
		_, err := p.DB.UpdateCharacterByID(character)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to update character in database: %s", err)
			return
		}
	}
}

func (p *Processor) processCharacterCorpHistory(id uint64) {
	var newEtag bool
	var history []monocle.CharacterCorporationHistory
	var response esi.Response

	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("\t%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

	historyEtag, err := p.DB.SelectEtagByIDAndResource(id, "character_corporation_history")
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("Unable to query character_corporation_history etag resource for Character %d due to SQL Error: %s", id, err)
			return
		}

		newEtag = true
		historyEtag.ID = id
		historyEtag.Resource = "character_corporation_history"
	}

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.Errorf("All Attempts exhuasted for Character %d", historyEtag.ID)
			break
		}
		response, historyEtag, err = p.ESI.GetCharactersCharacterIDCorporationHistory(historyEtag)
		if err != nil {
			p.Logger.Errorf("Error completing request to ESI for Character %d Corporation History: %s", historyEtag.ID, err)
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		p.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting %d request again in 1 second", response.Code, response.Path, attempts)
		time.Sleep(1 * time.Second)
	}

	history = response.Data.([]monocle.CharacterCorporationHistory)

	existing, err := p.DB.SelectCharacterCorporationHistoryByID(historyEtag.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("Unable to query character_corporation_history etag resource for Character %d due to SQL Error: %s", historyEtag.ID, err)
			return
		}
	}

	diff := diffExistingHistory(existing, history)
	switch newEtag {
	case true:
		_, err := p.DB.InsertEtag(historyEtag)
		if err != nil {
			p.Logger.Errorf("error encountered attempting to insert new etag for history into database: %s", err)
			return
		}
	case false:
		_, err := p.DB.UpdateEtagByIDAndResource(historyEtag)
		if err != nil {
			p.Logger.Errorf("error encountered attempting to insert new etag for history into database: %s", err)
			return
		}
	}

	if len(diff) > 0 {
		_, err := p.DB.InsertCharacterCorporationHistory(historyEtag.ID, diff)
		if err != nil {
			p.Logger.Errorf("error encountered attempting to insert new character corporation history records into database: %s", err)
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
