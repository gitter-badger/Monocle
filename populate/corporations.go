package populate

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/monocle"
)

// func (p *Populator) corpHunter() error {

// }

func (p *Populator) corpHunter2() error {
	var value struct {
		Value int `json:"value"`
	}
	kv, err := p.DB.SelectValueByKey("last_good_corporation_id")
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

	for x := begin; x <= 98999999; x += workers * records {
		end := x + (workers * records)
		msg := fmt.Sprintf("Errors: %d Remaining: %d Loop: %d - %d", p.ESI.Remain, p.ESI.Reset, x, x+(workers*records))
		p.Logger.CriticalF("\t%s", msg)
		for y := 1; y <= workers; y++ {
			if p.ESI.Remain < 20 {
				msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
				p.Logger.Errorf("\t%s", msg)
				// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
				// p.DGO.ChannelMessageSend("394991263344230411", msg)
				time.Sleep(time.Second * time.Duration(p.ESI.Reset))
			}
			ystart := (y * records) - records + x
			yend := (y * records) + x

			wg.Add(1)
			go func(start, end int) {

				for z := start; z <= end; z++ {
					p.processCorporation(uint(z))
				}
				// next <- true
				wg.Done()
			}(ystart, yend)
		}

		p.Logger.Debug("Done Dispatching. Waiting for Completion")
		wg.Wait()
		p.Logger.DebugF("Completed, sleep for %d seconds", sleep)
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

	msg := fmt.Sprintf("<@!277968564827324416> %s", "Corporation Hunter reached end of range")
	p.DGO.ChannelMessageSend("394991263344230411", msg)

	return nil
}

func (p *Populator) processCorporation(id uint) {

	var newCorporation bool

	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("\t%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// // p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}
	corporation, err := p.DB.SelectCorporationByCorporationID(id)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("DB Query for Corporation ID %d Failed with Error %s", id, err)
			return
		}
		corporation.ID = id
		newCorporation = true
	}

	if !corporation.IsExpired() {
		return
	}

	p.Logger.Debugf("\tProcessing Corp %d", corporation.ID)

	response, err := p.ESI.GetCorporationsCorporationID(corporation)
	if err != nil {
		p.Logger.Errorf("Error completing request to ESI for Corporation information: %s", err)
		return
	}

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.Errorf("All Attempts exhuasted for Corporation %d", corporation.ID)
			return
		}
		response, err = p.ESI.GetCorporationsCorporationID(corporation)
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

	p.Logger.Debugf("\tCorporation: %d:%s\tNew Corporation: %t", corporation.ID, corporation.Name, newCorporation)

	switch newCorporation {
	case true:
		_, err := p.DB.InsertCorporation(corporation)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to insert new corporation %d into database: %s", corporation.ID, err)
			return
		}
	case false:
		_, err := p.DB.UpdateCorporationByID(corporation)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to update corporation %d in database: %s", corporation.ID, err)
			return
		}
	}

}
