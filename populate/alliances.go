package populate

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
	"github.com/ddouglas/monocle/evewho"
	"github.com/pkg/errors"
)

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

// func (p *Populator) alliHunter() error {

// }

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

		etagResource, err := p.DB.SelectEtagByIDAndResource(uint64(alliance.ID), resource)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.Errorf("Unable to query for alliance member etag: %s", err)
			}
			etagResource.Exists = false
			etagResource.ID = uint64(alliance.ID)
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
