package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/esi"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type Character struct {
	model  monocle.Character
	exists bool
}

func (p *Processor) charHunter() {
	var value struct {
		Value uint64 `json:"value"`
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

		p.Logger.Infof("%d is valid, loop from %d to %d", end, x, end)

		for y := x; y <= end; y++ {
			if p.ESI.Remain < 20 {
				msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
				p.Logger.Errorf("\t%s", msg)
				// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
				// p.DGO.ChannelMessageSend("394991263344230411", msg)
				time.Sleep(time.Second * time.Duration(p.ESI.Reset))
			}
			wg.Add(1)

			character := monocle.Character{ID: uint64(y)}
			go func(model monocle.Character) {
				character := Character{
					model:  model,
					exists: false,
				}
				p.processCharacter(character)
				p.processCharacterCorpHistory(character)
				wg.Done()
				return
			}(character)
		}

		p.Logger.Info("Done Dispatching. Waiting for Completion")
		wg.Wait()

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

		p.Logger.InfoF("Completed, sleep for %d seconds", 2)
		time.Sleep(time.Second * 2)

	}

	return
}

func (p *Processor) charUpdater() {
	for {
		var characters []monocle.Character
		p.Logger.DebugF("Current Error Count: %d Remain: %d", p.ESI.Remain, p.ESI.Reset)
		if p.ESI.Remain < 20 {
			p.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
			time.Sleep(time.Second * time.Duration(p.ESI.Reset))
		}
		p.Logger.Info("Start")

		query := boiler.Characters(
			qm.Where(boiler.CharacterColumns.Expires+"<NOW()"),
			qm.And(boiler.CharacterColumns.Ignored+"=?", 0),
			qm.OrderBy(boiler.CharacterColumns.Expires),
			qm.Limit(int(records*workers)),
		)

		// queryStr, args := queries.BuildQuery(query.Query)
		// fmt.Println(queryStr)
		// fmt.Println(args...)

		err := query.Bind(context.Background(), p.DB, &characters)

		if err != nil {
			p.Logger.Errorf("No records returned from database", p.ESI.Reset)
			time.Sleep(time.Minute * 1)
			continue
		}

		if len(characters) < int(threshold) {
			p.Logger.Infof("Minimum threshold of %d for job not met. Sleeping for %d seconds", threshold, sleep)
			time.Sleep(time.Second * time.Duration(sleep))
			continue
		}

		p.Logger.Infof("Successfully Queried %d Characters", len(characters))

		charChunk := chunkCharacterSlice(int(records), characters)

		for _, characters := range charChunk {
			wg.Add(1)
			go func(characters []monocle.Character) {
				for _, model := range characters {
					character := Character{
						model:  model,
						exists: true,
					}
					p.processCharacter(character)
					p.processCharacterCorpHistory(character)
				}
				wg.Done()
			}(characters)
		}

		p.Logger.Info("Waiting")
		wg.Wait()
		p.Logger.Info("Done")
	}
}

func (p *Processor) processCharacter(character Character) {
	var response esi.Response
	var err error

	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("\t%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

	if !character.model.IsExpired() {
		return
	}

	p.Logger.Debugf("\tProcessing Char %d", character.model.ID)

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.Errorf("All Attempts exhuasted for Character %d", character.model.ID)
			return
		}
		response, err = p.ESI.GetCharactersCharacterID(character.model)
		if err != nil {
			p.Logger.Errorf("Error completing request to ESI for Character %d information: %s", character.model.ID, err)
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		p.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting %d request again in 1 second", response.Code, response.Path, attempts)
		time.Sleep(1 * time.Second)
	}

	character.model = response.Data.(monocle.Character)

	if character.model.CorporationID == 1000001 {
		character.model.Ignored = true
	}

	p.Logger.Debugf("\tCharacter: %d:%s\tNew Character: %t", character.model.ID, character.model.Name, !character.exists)

	switch !character.exists {
	case true:
		_, err := p.DB.InsertCharacter(character.model)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to insert new character into database: %s", err)
			return
		}
	case false:
		_, err := p.DB.UpdateCharacterByID(character.model)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to update character in database: %s", err)
			return
		}
	}

}

func (p *Processor) processCharacterCorpHistory(character Character) {
	var history []monocle.CharacterCorporationHistory
	var response esi.Response

	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("\t%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

	historyEtag, err := p.DB.SelectEtagByIDAndResource(character.model.ID, "character_corporation_history")
	historyEtag.Exists = true
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("Unable to query character_corporation_history etag resource for Character %d due to SQL Error: %s", character.model.ID, err)
			return
		}

		historyEtag.ID = character.model.ID
		historyEtag.Resource = "character_corporation_history"
		historyEtag.Exists = false
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

	diff := diffExistingCharCorpHistory(existing, history)

	switch !historyEtag.Exists {
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

func diffExistingCharCorpHistory(a []monocle.CharacterCorporationHistory, b []monocle.CharacterCorporationHistory) []monocle.CharacterCorporationHistory {
	c := convertCharCorpHistToMap(a)
	d := convertCharCorpHistToMap(b)
	result := make([]monocle.CharacterCorporationHistory, 0)
	for recordID, history := range d {
		if _, ok := c[recordID]; !ok {
			result = append(result, history)
		}
	}
	return result
}

func convertCharCorpHistToMap(a []monocle.CharacterCorporationHistory) map[uint]monocle.CharacterCorporationHistory {
	result := make(map[uint]monocle.CharacterCorporationHistory, 0)
	for _, history := range a {
		result[history.RecordID] = history
	}
	return result
}
