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
		p.processCharacter(id, false)
	}
	wg.Done()
	next <- true
	return
}

func (p *Populator) charHunter() error {

	for x := begin; x < done; x += workers * records {
		msg := fmt.Sprintf("Errors: %d Remaining: %d Loop: %d - %d", p.count, p.reset, x, x+(workers*records))
		p.Logger.CriticalF("\t%s", msg)

		for y := 1; y <= workers; y++ {
			if p.count < 20 {
				msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.reset)
				p.Logger.Errorf("\t%s", msg)
				msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
				p.DGO.ChannelMessageSend("394991263344230411", msg)
				time.Sleep(time.Second * time.Duration(p.reset))
			}
			ystart := (y * records) - records + x
			yend := (y * records) + x

			yresponse, err := p.ESI.HeadCharactersCharacterID(uint64(yend))
			if err != nil {
				p.Logger.Errorf("Error completing request to ESI for Character information: %s", err)
				continue
			}

			mx.Lock()
			p.reset = esi.RetrieveErrorResetFromResponse(yresponse)
			p.count = esi.RetrieveErrorCountFromResponse(yresponse)
			mx.Unlock()

			if yresponse.Code >= 400 && yresponse.Code < 500 {
				p.Logger.Errorf("Head Request for ID %d resulted in %d", yend, yresponse.Code)

				// for z := ystart; z <= yend; z++ {
				// 	p.DB.InsertCharacter(monocle.Character{
				// 		ID:      uint64(ystart),
				// 		Name:    "Invalid Character",
				// 		Expires: time.Now(),
				// 		Ignored: true,
				// 	})
				// }
				continue
			}

			wg.Add(1)
			go func(start, end int) {

				for z := start; z <= end; z++ {
					p.processCharacter(uint64(z), false)
				}
				// next <- true
				wg.Done()
			}(ystart, yend)
		}

		wg.Wait()
		time.Sleep(time.Millisecond * 500)
	}
	return nil
}

func (p *Populator) missingCharHunter() error {

	page := 1
	for {
		p.Logger.Infof("On Page: %d", page)
		characters, err := p.DB.SelectCharactersLikeName("Invalid Character", page, records)
		if err != nil {
			p.Logger.Errorf("Unable to query DB for character information: %s", err)
			continue
		}

		if len(characters) == 0 {
			break
		}

		for _, character := range characters {
			for z := character.ID; z < character.ID+10; z++ {
				p.processCharacter(character.ID, false)
				time.Sleep(time.Millisecond * 250)
			}
		}
		page++
	}

	// for x := begin; x < done; x += workers * records {
	// 	msg := fmt.Sprintf("Errors: %d Remaining: %d Loop: %d - %d", p.count, p.reset, x, x+(workers*records))
	// 	p.Logger.CriticalF("\t%s", msg)

	// 	for y := 1; y <= workers; y++ {
	// 		if p.count < 20 {
	// 			msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.reset)
	// 			p.Logger.Errorf("\t%s", msg)
	// 			msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
	// 			p.DGO.ChannelMessageSend("394991263344230411", msg)
	// 			time.Sleep(time.Second * time.Duration(p.reset))
	// 		}
	// 		ystart := (y * records) - records + x
	// 		yend := (y * records) + x

	// 		yresponse, err := p.ESI.HeadCharactersCharacterID(uint64(yend))
	// 		if err != nil {
	// 			p.Logger.Errorf("Error completing request to ESI for Character information: %s", err)
	// 			continue
	// 		}

	// 		mx.Lock()
	// 		p.reset = esi.RetrieveErrorResetFromResponse(yresponse)
	// 		p.count = esi.RetrieveErrorCountFromResponse(yresponse)
	// 		mx.Unlock()

	// 		if yresponse.Code >= 400 && yresponse.Code < 500 {
	// 			p.Logger.Errorf("Head Request for ID %d resulted in %d", yend, yresponse.Code)

	// 			// for z := ystart; z <= yend; z++ {
	// 			// 	p.DB.InsertCharacter(monocle.Character{
	// 			// 		ID:      uint64(ystart),
	// 			// 		Name:    "Invalid Character",
	// 			// 		Expires: time.Now(),
	// 			// 		Ignored: true,
	// 			// 	})
	// 			// }
	// 			continue
	// 		}

	// 		wg.Add(1)
	// 		go func(start, end int) {

	// 			for z := start; z <= end; z++ {
	// 				p.processCharacter(uint64(z), false)
	// 			}
	// 			// next <- true
	// 			wg.Done()
	// 		}(ystart, yend)
	// 	}

	// 	wg.Wait()
	// 	time.Sleep(time.Millisecond * 500)
	// }
	return nil
}

// func (p *Populator) corpHistory() error {
// 	page := 0

// 	for {
// 		p.Logger.DebugF("Current Error Count: %d Remain: %d", p.count, p.reset)
// 		if p.count < 20 {
// 			p.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", p.reset)
// 			time.Sleep(time.Second * time.Duration(p.reset))
// 		}

// 		for x := 1; x <= workers; x++ {
// 			characters, err := p.DB.SelectCharactersFromRange(x, records)
// 			if err != nil {
// 				if err != sql.ErrNoRows {
// 					p.Logger.Fatalf("Unable to query for characters: %s", err)
// 				}
// 				continue
// 			}

// 			wg.Add(1)
// 			go func(characters []monocle.Character) {
// 				for _, character := range characters {
// 					p.processCharacterCorpHistory(character.ID)
// 				}
// 				wg.Done()
// 				return
// 			}(characters)
// 		}

// 		wg.Wait()
// 		page++
// 	}
// }

func (p *Populator) processCharacter(id uint64, newCharacter bool) {

	if p.count < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.reset)
		p.Logger.Errorf("\t%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.reset))
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

	response, err := p.ESI.GetCharactersCharacterID(character.ID, "")
	if err != nil {
		p.Logger.Errorf("Error completing request to ESI for Character information: %s", err)
		return
	}

	character, proceed := p.processCharacterResponse(response, character)
	if !proceed {
		return
	}

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

func (p *Populator) processCharacterResponse(response esi.Response, character monocle.Character) (monocle.Character, bool) {

	mx.Lock()
	defer mx.Unlock()
	p.reset = esi.RetrieveErrorResetFromResponse(response)
	p.count = esi.RetrieveErrorCountFromResponse(response)

	switch response.Code {
	case 200:
		err := json.Unmarshal(response.Data.([]byte), &character)
		if err != nil {
			p.Logger.Errorf("unable to unmarshel response body: %s", err)
			return character, false
		}
		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		character.Etag = etag

		character.Expires = expires
		break
	case 304:
		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}
		character.Expires = expires

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		character.Etag = etag

		break
	case 420:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return character, false
	case 400:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return character, false
	case 404:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return character, false
	default:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		character.Name = "Invalid Character ID!"
		character.Ignored = true
		character.Expires = time.Now()
	}

	return character, true

}

// func (p *Populator) processCharacterCorpHistory(id uint64) {

// 	if p.count < 20 {
// 		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.reset)
// 		p.Logger.Errorf("\t%s", msg)
// 		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
// 		// p.DGO.ChannelMessageSend("394991263344230411", msg)
// 		time.Sleep(time.Second * time.Duration(p.reset))
// 	}

// 	// var newHistory bool
// 	var history []monocle.CharacterCorporationHistory
// 	var etag monocle.EtagResource

// 	etag, err := p.DB.SelectEtagByIDAndResource(id, "character_corporation_history")
// 	if err != nil {
// 		if err != sql.ErrNoRows {
// 			return
// 		}

// 		newHistory = true
// 		etag.ID = id
// 	}

// 	response, err := p.ESI.GetCharactersCharacterIDCorporationHistory(id, etag.Etag)
// 	if err != nil {
// 		p.Logger.Errorf("Error completeing request to ESI for Character &d corporation history: %s", id, err)
// 		return
// 	}

// 	history, etag, proceed := p.processCharacterCorpHistoryResponse(response, history, etag)
// 	if !proceed {
// 		return
// 	}

// }

func (p *Populator) processCharacterCorpHistoryResponse(response esi.Response, history []monocle.CharacterCorporationHistory, etag monocle.EtagResource) ([]monocle.CharacterCorporationHistory, monocle.EtagResource, bool) {

	mx.Lock()
	defer mx.Unlock()
	p.reset = esi.RetrieveErrorResetFromResponse(response)
	p.count = esi.RetrieveErrorCountFromResponse(response)

	switch response.Code {
	case 200:
		err := json.Unmarshal(response.Data.([]byte), &history)
		if err != nil {
			p.Logger.Errorf("unable to unmarshel response body: %s", err)
			return history, etag, false
		}
		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}

		etagStr, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		etag.Etag = etagStr

		etag.Expires = expires
		break
	case 304:
		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}
		etag.Expires = expires

		etagStr, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		etag.Etag = etagStr

		break
	case 420:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return history, etag, false
	default:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		etag.Expires = time.Now()
	}

	return history, etag, true

}
