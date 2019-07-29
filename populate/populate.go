package populate

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/ddouglas/monocle/evewho"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/core"
	"github.com/ddouglas/monocle/esi"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Populator struct {
	*core.App
	count, reset uint64
}

var (
	workers,
	// threshold,
	errorCount,
	records int
	sleep       int
	begin, done int
	scope       string
	wg          sync.WaitGroup
	mx          sync.Mutex
)

func Action(c *cli.Context) error {
	core, err := core.New()
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
		return cli.NewExitError(err, 1)
	}

	populator := Populator{core, 100, 40}

	scope = c.String("scope")
	workers = c.Int("workers")
	records = c.Int("records")
	done = c.Int("done")
	sleep = c.Int("sleep")

	if populator.Config.PopulateBegin > 0 {
		begin = populator.Config.PopulateBegin
	} else if c.Int("begin") > 0 {
		begin = c.Int("begin")
	} else {
		begin = 90000000
	}

	populator.Logger.Infof("Starting process with %d workers", workers)

	switch scope {
	case "getAlliancelList":
		_ = populator.getAlliancelList()
	case "getAllianceCorpMemberList":
		_ = populator.getAllianceCorpList()
	case "getAllianceCharMemberList":
		_ = populator.getAllianceCharList()
	case "getCorpCharList":
		_ = populator.getCorpCharList()
	case "charHunter":
		_ = populator.charHunter()
	}

	return nil
}

func (p *Populator) getAlliancelList() error {

	response, err := p.ESI.GetAlliances("")
	if err != nil {
		p.Logger.Errorf("Error completing request to ESI for List of Alliances: %s", err)
		return err
	}

	var ids []int

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &ids)
		if err != nil {
			p.Logger.Errorf("unable to unmarshal response body: %s", err)
			return err
		}
	default:
		p.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s:", response.Code, response.Path)
		return nil
	}

	chunkedIds := chunkIntSlice(250, ids)

	for _, chunk := range chunkedIds {

		ids, err := p.DB.SelectMissingAllianceIdsFromList(chunk)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.Fatalf("Unable to query for characters: %s", err)
			}
			continue
		}

		if len(ids) > 0 {
			p.Logger.Infof("Deploying Go Routines to Process %d ids", len(ids))
			wg.Add(1)
			go func(id []uint) {
				defer wg.Done()
				for _, v := range ids {
					p.processAlliance(v)
				}
				return
			}(ids)
		}

		time.Sleep(time.Second * 5)
	}
	p.Logger.Infof("Routines Launched. Awaiting Completion")
	wg.Wait()
	p.Logger.Infof("Done. Returning")
	return nil
}

func (p *Populator) getAllianceCorpList() error {
	var page = 1

	for {
		if errorCount >= 10 {
			return errors.New("Error Count High. Exiting Program")
		}

		alliances, err := p.DB.SelectAlliances(page, records)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.Fatalf("Unable to query for alliances: %s", err)
				break
			}
		}

		if len(alliances) == 0 {
			p.Logger.Info("All Alliance Member Corps updated. Breaking Superviser Loop")
			break
		}

		p.Logger.Infof("Successfully Queried %d Alliances", len(alliances))
		wg.Add(1)
		go p.processAllianceCorps(page, alliances)
		page++

		time.Sleep(time.Second * time.Duration(sleep))
	}

	p.Logger.Debug("Master Loop broken. Waiting for any remaining Routines")
	wg.Wait()
	p.Logger.Debug("Done Waiting. Return nil for errors")

	return nil

}

func (p *Populator) getAllianceCharList() error {

	var page = 1
	for {
		alliances, err := p.DB.SelectAlliances(page, records)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.Fatalf("Unable to query for alliances: %s", err)
				continue
			}
		}

		if len(alliances) == 0 {
			p.Logger.Info("All Alliance Member Corps updated. Breaking supervisor loop")
			break
		}

		p.Logger.Infof("Successfully Queried %d Alliances", len(alliances))
		p.processAllianceChars(page, alliances)
		page++
	}

	return nil

}

func (p *Populator) getCorpCharList() error {

	var page = 1
	for {
		corps, err := p.DB.SelectIndependentCorps(page, records)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.Fatalf("Unable to query for alliances: %s", err)
				continue
			}
		}

		if len(corps) == 0 {
			p.Logger.Info("All Alliance Member Corps updated. Breaking supervisor loop")
			break
		}

		p.Logger.Infof("Successfully Queried %d Independent Corps", len(corps))
		p.processCorporationChars(corps)
		page++
	}

	return nil

}

func (p *Populator) processAlliance(id uint) {

	var alliance monocle.Alliance

	response, err := p.ESI.GetAlliancesAllianceID(id, "")
	if err != nil {
		p.Logger.Errorf("Error completing request to ESI for Alliance information: %s", err)
		return
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &alliance)
		if err != nil {
			p.Logger.Errorf("unable to unmarshel response body: %s", err)
			return
		}

		alliance.ID = id

		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}

		alliance.Expires = expires

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		alliance.Etag = etag
		break

	default:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	p.Logger.Debugf("\tAlliance: %d:%s", alliance.ID, alliance.Name)

	_, err = p.DB.InsertAlliance(alliance)
	if err != nil {
		p.Logger.Errorf("Error Encountered attempting to insert new alliance into database: %s", err)
		return
	}
}

func (p *Populator) processAllianceCorps(pid int, alliances []monocle.Alliance) {

	for _, alliance := range alliances {

		var corpIDs []int

		resource := "alliance_corp_list"

		etagResource, err := p.DB.SelectEtagByIDAndResource(alliance.ID, resource)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.Errorf("Unable to query for alliance member etag: %s", err)
			}
			etagResource.Exists = false
			etagResource.ID = alliance.ID
			etagResource.Resource = resource
		}

		if !etagResource.IsExpired() {
			continue
		}

		response, err := p.ESI.GetAllianceMembersByID(alliance.ID, etagResource.Etag)
		if err != nil {
			p.Logger.Errorf("Error completing request to ESI for Alliance information: %s", err)
			return
		}

		switch response.Code {
		case 200:
			err = json.Unmarshal(response.Data.([]byte), &corpIDs)
			if err != nil {
				p.Logger.Errorf("unable to unmarshel response body: %s", err)
				return
			}

			expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
			if err != nil {
				p.Logger.Errorf("Error Encountered attempting to parse expires header: %s", err)
			}

			etagResource.Expires = expires

			etag, err := esi.RetrieveEtagHeaderFromResponse(response)
			if err != nil {
				p.Logger.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
			}

			etagResource.Etag = etag

			_, err = p.DB.InsertEtag(etagResource)
			if err != nil {
				p.Logger.Errorf("Error Received when attempting to insert Etag into database: %s", err)
				return
			}

			break

		case 304:
			expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
			if err != nil {
				p.Logger.Errorf("Error Encountered attempting to parse expires header: %s", err)
			}

			etagResource.Expires = expires

			etag, err := esi.RetrieveEtagHeaderFromResponse(response)
			if err != nil {
				p.Logger.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
			}

			etagResource.Etag = etag

			_, err = p.DB.InsertEtag(etagResource)
			if err != nil {
				p.Logger.Errorf("Error Received when attempting to insert Etag into database: %s", err)
			}

			return
		default:
			p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
			return
		}

		if len(corpIDs) == 0 {

			alliance.Closed = true
			p.Logger.Debugf("%d Corps Detected for Alliance %d. Closing the Corp in DB and continuing to next corp", len(corpIDs), alliance.ID)

			_, err := p.DB.UpdateAllianceByID(alliance)
			if err != nil {
				p.Logger.Errorf("unable to close alliance in database: %s", err)
			}
			continue
		}

		chunked := chunkIntSlice(50, corpIDs)
		chunkedLen := len(chunked)
		p.Logger.Debugf("%d corp Ids found and chunked into %d chunks. Starting Chunk Loop", len(corpIDs), chunkedLen)
		for y, chunk := range chunked {
			p.Logger.Debugf("Starting Loop %d of %d", y+1, chunkedLen)
			missing, err := p.DB.SelectMissingCorporationIdsFromList(pid, chunk)
			if err != nil {
				if err != sql.ErrNoRows {
					p.Logger.Fatalf("Unable to query for characters: %s", err)
				}
				continue
			}

			if len(missing) == 0 {
				p.Logger.Debugf("0 missing ids found for Alliance %d", alliance.ID)
				continue
			}

			if len(missing) > 0 {
				p.Logger.Infof("Deploying Go Routines to Process %d ids", len(missing))
				wg.Add(1)
				go func(missing []int) {
					for _, v := range missing {
						p.processNewCorporation(uint(v))
					}
					wg.Done()
					return
				}(missing)
			}

			time.Sleep(time.Second * 1)
			p.Logger.Debugf("Finishing Loop %d of %d ", y+1, chunkedLen)
		}

	}
	return
}

func (p *Populator) processNewCorporation(id uint) {

	var corporation monocle.Corporation
	response, err := p.ESI.GetCorporationsCorporationID(id, "")
	if err != nil {
		p.Logger.Errorf("Error completing request to ESI for Character information: %s", err)
		return
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &corporation)
		if err != nil {
			p.Logger.Errorf("unable to unmarshel response body: %s", err)
			return
		}

		corporation.ID = id

		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}
		corporation.Expires = expires

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		corporation.Etag = etag

		break

	default:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	p.Logger.Debugf("\tCorporation: %d:%s", corporation.ID, corporation.Name)

	_, err = p.DB.InsertCorporation(corporation)
	if err != nil {
		p.Logger.Errorf("Error Encountered attempting to insert new corporation into database: %s", err)
		return
	}
}

func (p *Populator) processAllianceChars(pid int, alliances []monocle.Alliance) {

	next := make(chan bool)
	limiter := 0

	for _, alliance := range alliances {

		var ewAlliance evewho.AllianceList

		page := 0

		for {
			var characters []uint64
			p.Logger.Infof("Requesting Page %d Member Data for Alliance %d", page, alliance.ID)

			response, err := p.Who.GetAllianceMembersByID(alliance.ID, page)
			if err != nil {
				p.Logger.Errorf("Error completing request to ESI for Alliance information: %s", err)
				return
			}

			p.Logger.Infof("Received Response Code of %d for %d", response.Code, alliance.ID)

			switch response.Code {
			case 200:
				err = json.Unmarshal(response.Data.([]byte), &ewAlliance)
				if err != nil {
					p.Logger.Errorf("unable to unmarshel response body: %s", err)
					return
				}
				break
			default:
				p.Logger.ErrorF("Bad Resposne Code %d received from EveWho API for url %s:", response.Code, response.Path)
				return
			}

			ewCharacters := ewAlliance.Characters
			ewCharactersLen := len(ewCharacters)
			p.Logger.Infof("Alliance %d has approximately %d characters.", alliance.ID, ewCharactersLen)
			if ewCharactersLen == 0 {

				alliance.Closed = true
				p.Logger.Debugf("%d Chars Detected for Alliance %d. Closing the Corp in DB and continuing to next corp", len(ewCharacters), alliance.ID)

				_, err := p.DB.UpdateAllianceByID(alliance)
				if err != nil {
					p.Logger.Errorf("unable to close alliance in database: %s", err)
				}
				break
			}

			for _, ewCharacter := range ewCharacters {
				characterID, err := strconv.ParseUint(ewCharacter.CharacterID, 10, 64)
				if err != nil {
					p.Logger.Errorf("Unable to parse %s to uint64 for esi client", ewCharacter.CharacterID)
					continue
				}
				characters = append(characters, characterID)
			}
			wg.Add(1)
			go p.processCharacterList(characters, next)

			if limiter >= workers {
				select {
				case <-next:
					p.Logger.Info("received value on done chan")

				}
			}
			limiter++
			if ewCharactersLen < 200 {
				break
			}

			page++
			time.Sleep(1 * time.Second)
		}
	}

	p.Logger.Info("Parent Loop Exited. Waiting for inflight goroutines to complete")
	wg.Wait()
	p.Logger.Info("In Flight GoRoutines Done")

	return

}

func (p *Populator) processAllianceCharList(pid int, alliances []monocle.Alliance) {

	next := make(chan bool)
	limiter := 0

	for _, alliance := range alliances {

		var ewAlliance evewho.AllianceList

		page := 0

		for {
			var characters []uint64
			p.Logger.Infof("Requesting Page %d Member Data for Alliance %d", page, alliance.ID)

			response, err := p.Who.GetAllianceMembersByID(alliance.ID, page)
			if err != nil {
				p.Logger.Errorf("Error completing request to ESI for Alliance information: %s", err)
				return
			}

			p.Logger.Infof("Received Response Code of %d for %d", response.Code, alliance.ID)

			switch response.Code {
			case 200:
				err = json.Unmarshal(response.Data.([]byte), &ewAlliance)
				if err != nil {
					p.Logger.Errorf("unable to unmarshel response body: %s", err)
					return
				}
				break
			default:
				p.Logger.ErrorF("Bad Resposne Code %d received from EveWho API for url %s:", response.Code, response.Path)
				return
			}

			ewCharacters := ewAlliance.Characters
			ewCharactersLen := len(ewCharacters)
			p.Logger.Infof("Alliance %d has approximately %d characters.", alliance.ID, ewCharactersLen)
			if ewCharactersLen == 0 {

				alliance.Closed = true
				p.Logger.Debugf("%d Chars Detected for Alliance %d. Closing the Corp in DB and continuing to next corp", len(ewCharacters), alliance.ID)

				_, err := p.DB.UpdateAllianceByID(alliance)
				if err != nil {
					p.Logger.Errorf("unable to close alliance in database: %s", err)
				}
				break
			}

			for _, ewCharacter := range ewCharacters {
				characterID, err := strconv.ParseUint(ewCharacter.CharacterID, 10, 64)
				if err != nil {
					p.Logger.Errorf("Unable to parse %s to uint64 for esi client", ewCharacter.CharacterID)
					continue
				}
				characters = append(characters, characterID)

			}
			wg.Add(1)
			go p.processCharacterList(characters, next)

			if limiter >= workers {
				select {
				case <-next:
					p.Logger.Info("received value on done chan")

				}
			}
			limiter++
			if ewCharactersLen < 200 {
				break
			}

			page++
			time.Sleep(1 * time.Second)
		}
	}

	p.Logger.Info("Parent Loop Exited. Waiting for inflight goroutines to complete")
	wg.Wait()
	p.Logger.Info("In Flight GoRoutines Done")

	return

}

func (p *Populator) processCorporationChars(corporations []monocle.Corporation) {

}

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

			if yresponse.Code > 200 {
				p.Logger.Errorf("Head Request for ID %d resulted in %d", yend, yresponse.Code)

				for z := ystart; z <= yend; z++ {
					p.DB.InsertCharacter(monocle.Character{
						ID:      uint64(ystart),
						Name:    "Invalid Character",
						Expires: time.Now(),
						Ignored: true,
					})
				}
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

	response, err := p.ESI.GetCharactersCharacterID(character.ID, character.Etag)
	if err != nil {
		p.Logger.Errorf("Error completing request to ESI for Character information: %s", err)
		return
	}

	mx.Lock()
	defer mx.Unlock()
	p.reset = esi.RetrieveErrorResetFromResponse(response)
	p.count = esi.RetrieveErrorCountFromResponse(response)

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &character)
		if err != nil {
			p.Logger.Errorf("unable to unmarshel response body: %s", err)
			return
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
		time.Sleep(10 * time.Second)
		return
	default:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		character.Name = "Invalid Character ID!"
		character.Ignored = true
		character.Expires = time.Now()
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
