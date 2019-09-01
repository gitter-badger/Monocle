package processor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

func (p *Processor) corpHunter() {
	var value struct {
		Value uint64 `json:"value"`
	}

	kv, err := p.DB.SelectValueByKey("last_good_corporation_id")
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

	for x := begin; x <= 98999999; x++ {
		msg := fmt.Sprintf("Errors: %d Remaining: %d ID: %d", p.ESI.Remain, p.ESI.Reset, x)
		p.Logger.CriticalF("%s", msg)
		attempts := 0
		for {
			if attempts >= 2 {
				msg := fmt.Sprintf("Head Requests to %d failed. Sleep for %d minutes before trying again", x, sleep)
				p.Logger.Errorf("%s", msg)
				// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
				// p.DGO.ChannelMessageSend("394991263344230411", msg)
				time.Sleep(time.Minute * time.Duration(sleep))
				attempts = 0
			}
			p.Logger.DebugF("Checking for validity of %d", x)
			response, err := p.ESI.HeadCorporationsCorporationID(uint(x))
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

			p.processCorporation(x)
			p.processCorporationAllianceHistory(x)
			break
		}

		value.Value = x

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

		p.Logger.InfoF("Completed, sleep for %d seconds", 2)
		time.Sleep(time.Second * 1)

	}

	return
}

// func (p *Processor) corpUpdater() {
// 	for {
// 		p.Logger.DebugF("Current Error Count: %d Remain: %d", p.ESI.Remain, p.ESI.Reset)
// 		if p.ESI.Remain < 20 {
// 			p.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
// 			time.Sleep(time.Second * time.Duration(p.ESI.Reset))
// 		}

// 		characters, err := p.DB.SelectExpiredCharacterEtags(records * workers)
// 		if err != nil {
// 			if err != sql.ErrNoRows {
// 				p.Logger.Fatalf("Unable to query for characters: %s", err)
// 			}
// 			continue
// 		}

// 		if len(characters) < threshold {
// 			p.Logger.Infof("Minimum threshold of %d for job not met. Sleeping for %d seconds", threshold, sleep)
// 			time.Sleep(time.Second * time.Duration(sleep))
// 			continue
// 		}

// 		p.Logger.Infof("Successfully Queried %d Characters", len(characters))

// 		charChunk := chunkCharacterSlice(records, characters)

// 		for _, characters := range charChunk {
// 			wg.Add(1)
// 			go func(characters []monocle.Character) {
// 				for _, character := range characters {
// 					p.processCharacter(character.ID)
// 					p.processCharacterCorpHistory(character.ID)
// 				}
// 				wg.Done()
// 			}(characters)
// 		}

// 		p.Logger.Info("Waiting")
// 		wg.Wait()
// 		p.Logger.Info("Done")
// 	}
// }

func (p *Processor) processCorporation(id uint64) {

	var response esi.Response
	var new bool
	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

	corporation, err := p.DB.SelectCorporationByCorporationID(id)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("DB Query for Corporation ID %d Failed with Error %s", id, err)
			return
		}
		corporation.ID = id
		new = true
	}

	if !corporation.IsExpired() {
		return
	}

	p.Logger.Debugf("Processing Corp %d", corporation.ID)

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.Errorf("All Attempts exhuasted for Corporation %d", corporation.ID)
			return
		}
		response, err = p.ESI.GetCorporationsCorporationID(corporation.ID, corporation.Etag)
		if err != nil {
			p.Logger.Errorf("Error completing request to ESI for Corporation %d information: %s", corporation.ID, err)
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		p.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting %d request again in 1 second", response.Code, response.Path, attempts)
		time.Sleep(1 * time.Second)
	}

	corporation = response.Data.(monocle.Corporation)

	p.Logger.Debugf("Corporation: %d:%s\tNew Corporation: %t", corporation.ID, corporation.Name, new)

	switch new {
	case true:
		_, err := p.DB.InsertCorporation(corporation)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to insert new corporation into database: %s", err)
			return
		}
	case false:
		_, err := p.DB.UpdateCorporationByID(corporation)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to update corporation in database: %s", err)
			return
		}
	}
}

func (p *Processor) processCorporationAllianceHistory(id uint64) {
	var newEtag bool
	var history []monocle.CorporationAllianceHistory
	var response esi.Response

	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

	historyEtag, err := p.DB.SelectEtagByIDAndResource(id, "corporation_alliance_history")
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("Unable to query corporation_alliance_history etag resource for Character %d due to SQL Error: %s", id, err)
			return
		}

		newEtag = true
		historyEtag.ID = id
		historyEtag.Resource = "corporation_alliance_history"
	}

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.Errorf("All Attempts exhuasted for Character %d", historyEtag.ID)
			return
		}
		response, err = p.ESI.GetCorporationsCorporationIDAllianceHistory(historyEtag)
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

	data := response.Data.(map[string]interface{})
	history = data["history"].([]monocle.CorporationAllianceHistory)
	historyEtag = data["etag"].(monocle.EtagResource)

	existing, err := p.DB.SelectCorporationAllianceHistoryByID(historyEtag.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("Unable to query corporation_alliance_history etag resource for Character %d due to SQL Error: %s", historyEtag.ID, err)
			return
		}
	}

	diff := diffExistingCorpAlliHistory(existing, history)
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
		_, err := p.DB.InsertCorporationAllianceHistory(historyEtag.ID, diff)
		if err != nil {
			p.Logger.Errorf("error encountered attempting to insert new character corporation history records into database: %s", err)
			return
		}
	}
	return

}

func diffExistingCorpAlliHistory(a []monocle.CorporationAllianceHistory, b []monocle.CorporationAllianceHistory) []monocle.CorporationAllianceHistory {
	c := convertCorpAlliHistToMap(a)
	d := convertCorpAlliHistToMap(b)
	result := make([]monocle.CorporationAllianceHistory, 0)
	for recordID, history := range d {
		if _, ok := c[recordID]; !ok {
			result = append(result, history)
		}
	}
	return result
}

func convertCorpAlliHistToMap(a []monocle.CorporationAllianceHistory) map[uint]monocle.CorporationAllianceHistory {
	result := make(map[uint]monocle.CorporationAllianceHistory, 0)
	for _, history := range a {
		result[history.RecordID] = history
	}
	return result
}
