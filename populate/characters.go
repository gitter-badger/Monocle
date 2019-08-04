package populate

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

func (p *Populator) processCharacterList(ids []uint64, next chan bool) {
	for _, id := range ids {
		p.processCharacter(id)
	}
	wg.Done()
	next <- true
	return
}

func (p *Populator) charHunter() error {
	var response esi.Response
	var value struct {
		Value int `json:"value"`
	}
	kv, err := p.DB.SelectValueByKey("last_good_character_id")
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Criticalf("Unable to query for ID: %s", err)
			return err
		}
	}

	err = json.Unmarshal(kv.Value, &value)
	if err != nil {
		p.Logger.Criticalf("Unable to unmarshal value: %s", err)
		return err
	}

	begin = value.Value

	p.Logger.Infof("Starting at ID %d", begin)

	for x := begin; x <= 2147483647; x += workers * records {
		end := x + (workers * records)
		msg := fmt.Sprintf("Errors: %d Remaining: %d Loop: %d - %d", p.ESI.Remain, p.ESI.Reset, x, x+(workers*records))
		p.Logger.CriticalF("\t%s", msg)

		for {
			p.Logger.InfoF("Checking for valid end of %d", end)
			response, err = p.ESI.HeadCharactersCharacterID(uint64(end))
			if err != nil {
				p.Logger.ErrorF(err.Error())
				if response.Code >= 500 {
					time.Sleep(time.Second * 5)
					continue
				}
				time.Sleep(time.Minute * 30)
				continue
			}
			break
		}

		for y := 1; y <= workers; y++ {
			if p.ESI.Remain < 20 {
				msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
				p.Logger.Errorf("\t%s", msg)
				msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
				p.DGO.ChannelMessageSend("394991263344230411", msg)
				time.Sleep(time.Second * time.Duration(p.ESI.Reset))
			}
			ystart := (y * records) - records + x
			yend := (y * records) + x

			wg.Add(1)
			go func(start, end int) {

				for z := start; z <= end; z++ {
					p.processCharacter(uint64(z))
					p.processCharacterCorpHistory(uint64(z))
				}
				// next <- true
				wg.Done()
			}(ystart, yend)
		}

		p.Logger.Info("Done Dispatching. Waiting for Completion")
		wg.Wait()
		p.Logger.Infof("Completed, sleep for %d seconds", sleep)
		time.Sleep(time.Second * time.Duration(sleep))

		value.Value = end

		kv.Value, err = json.Marshal(value)
		if err != nil {
			p.Logger.Criticalf("Unable to unmarshal value: %s", err)
			return err
		}

		_, err = p.DB.UpdateValueByKey(kv)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.Criticalf("Unable to query for ID: %s", err)
				return err
			}
		}

	}

	return nil
}

func (p *Populator) processCharacter(id uint64) {

	var newCharacter bool

	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("\t%s", msg)
		msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		p.DGO.ChannelMessageSend("394991263344230411", msg)
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

	response, err := p.ESI.GetCharactersCharacterID(character)
	if err != nil {
		p.Logger.Errorf("Error completing request to ESI for Character information: %s", err)
		return
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

func (p *Populator) processCharacterCorpHistory(id uint64) {
	var newEtag bool
	var history []monocle.CharacterCorporationHistory
	var response esi.Response

	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("\t%s", msg)
		msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		p.DGO.ChannelMessageSend("394991263344230411", msg)
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
