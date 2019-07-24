package populate

import (
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/core"
	"github.com/ddouglas/monocle/esi"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Populator struct {
	*core.App
}

var (
	workers,
	// threshold,
	errorCount,
	records int
	sleep int
	scope string
	wg    sync.WaitGroup
)

func Action(c *cli.Context) error {
	core, err := core.New()
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")
		log.Fatal(err)
		return cli.NewExitError(err, 1)
	}

	populator := Populator{core}

	scope = c.String("scope")
	workers = c.Int("workers")
	records = c.Int("records")
	sleep = c.Int("sleep")

	switch scope {
	case "alliancelist":
		_ = populator.getAlliancelList()
	case "alliancemembers":
		_ = populator.getAllianceMemberList()
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
			go p.processAllianceIds(ids)
		}

		time.Sleep(time.Second * 5)
	}
	p.Logger.Infof("Routines Launched. Awaiting Completion")
	wg.Wait()
	p.Logger.Infof("Done. Returning")
	return nil
}

func (p *Populator) getAllianceMemberList() error {
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

		page++
		p.processAllianceMembers(alliances)

		time.Sleep(time.Second * time.Duration(sleep))
	}

	p.Logger.Debug("Master Loop broken. Waiting for any remaining Routines")

	p.Logger.Debug("Done Waiting. Return nil for errors")

	return nil

}

func (p *Populator) processAllianceMembers(alliances []monocle.Alliance) {

	for _, alliance := range alliances {

		var corpIDs []int

		response, err := p.ESI.GetAllianceMembersByID(alliance.ID, "")
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

			expires, err := esi.RetreiveExpiresHeaderFromResponse(response)
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

		if len(corpIDs) == 0 {

			alliance.Closed = true
			p.Logger.Debugf("%d Corps Detected for Alliance %d. Closing the Corp in DB and continuing to next corp", len(corpIDs), alliance.ID)

			_, err := p.DB.UpdateAllianceByID(alliance)
			if err != nil {
				p.Logger.Errorf("unable to close alliance in database: %s", err)
			}
			continue
		}

		if len(corp)


		chunked := chunkIntSlice(50, corpIDs)
		chunkedLen := len(chunked)
		p.Logger.Debugf("%d corp Ids found and chunked into %d chunkks. Starting Chunk Loop", len(corpIDs), chunkedLen)
		for y, chunk := range chunked {
			p.Logger.Debugf("Starting Loop %d of %d", y+1, chunkedLen)
			missing, err := p.DB.SelectMissingCorporationIdsFromList(chunk)
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
				go p.processCorporationIds(missing)
			}

			time.Sleep(time.Second * 1)
			p.Logger.Debugf("Finishing Loop %d of %d ", y+1, chunkedLen)
		}

	}
	return
}

func (p *Populator) processCorporationIds(ids []int) {
	defer wg.Done()
	for _, v := range ids {
		p.processCorporation(uint(v))
	}
	return
}

func (p *Populator) processAllianceIds(ids []monocle.AllianceIDs) {
	defer wg.Done()
	for _, v := range ids {
		p.processAlliance(v.ID)
	}
	return
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

		expires, err := esi.RetreiveExpiresHeaderFromResponse(response)
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

func (p *Populator) processCharacter(id uint64) {

	var character monocle.Character
	response, err := p.ESI.GetCharactersCharacterID(id, "")
	if err != nil {
		p.Logger.Errorf("Error completing request to ESI for Character information: %s", err)
		return
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &character)
		if err != nil {
			p.Logger.Errorf("unable to unmarshel response body: %s", err)
			return
		}
		expires, err := esi.RetreiveExpiresHeaderFromResponse(response)
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
	default:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	p.Logger.Debugf("\tCharacter: %d:%s", character.ID, character.Name)

	_, err = p.DB.InsertCharacter(character)
	if err != nil {
		p.Logger.Errorf("Error Encountered attempting to insert new character into database: %s", err)
		return
	}
}

func (p *Populator) processCorporation(id uint) {

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

		expires, err := esi.RetreiveExpiresHeaderFromResponse(response)
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

func chunkIntSlice(size int, slice []int) [][]int {

	var chunk [][]int
	chunk = make([][]int, 0)

	if len(slice) <= size {
		chunk = append(chunk, slice)
		return chunk
	}

	for x := 0; x <= len(slice); x += size {

		end := x + size

		if end > len(slice) {
			end = len(slice)
		}

		chunk = append(chunk, slice[x:end])

	}

	return chunk
}
