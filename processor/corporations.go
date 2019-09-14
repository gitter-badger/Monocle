package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/esi"
)

type Corporation struct {
	model  monocle.Corporation
	exists bool
}

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

			corporation := Corporation{
				model: monocle.Corporation{
					ID: uint64(x),
				},
				exists: false,
			}
			p.processCorporation(corporation)
			p.processCorporationAllianceHistory(corporation)
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

		p.Logger.InfoF("Completed, sleep for %d seconds", sleep)
		time.Sleep(time.Second * time.Duration(sleep))

	}

	return
}

func (p *Processor) corpUpdater() {
	for {

		var corporations []monocle.Corporation
		p.Logger.DebugF("Current Error Count: %d Remain: %d", p.ESI.Remain, p.ESI.Reset)
		if p.ESI.Remain < 20 {
			p.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
			time.Sleep(time.Second * time.Duration(p.ESI.Reset))
		}
		err := boiler.Corporations(
			qm.Where(boiler.CorporationColumns.Expires+"<NOW()"),
			qm.And(boiler.CorporationColumns.Ignored+"=?", 0),
			qm.And(boiler.CorporationColumns.Closed+"=?", 0),
			qm.OrderBy(boiler.CorporationColumns.Expires),
			qm.Limit(int(records*workers)),
		).Bind(context.Background(), p.DB, &corporations)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.Fatalf("Unable to query for characters: %s", err)
			}
			continue
		}

		if len(corporations) < int(threshold) {
			p.Logger.Infof("Minimum threshold of %d for job not met. Sleeping for %d seconds", threshold, sleep)
			time.Sleep(time.Second * time.Duration(sleep))
			continue
		}

		p.Logger.Infof("Successfully Queried %d Corporations", len(corporations))

		corpChunk := chunkCorporationSlice(int(records), corporations)

		for _, corporations := range corpChunk {
			wg.Add(1)
			go func(corporations []monocle.Corporation) {
				for _, model := range corporations {
					corporation := Corporation{
						model:  model,
						exists: true,
					}
					p.processCorporation(corporation)
					p.processCorporationAllianceHistory(corporation)
				}
				wg.Done()
			}(corporations)
		}

		p.Logger.Info("Waiting")
		wg.Wait()
		p.Logger.Info("Done")
	}
}

func (p *Processor) processCorporation(corporation Corporation) {

	var response esi.Response
	var err error
	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

	if !corporation.model.IsExpired() {
		return
	}

	p.Logger.Debugf("Processing Corp %d", corporation.model.ID)

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.Errorf("All Attempts exhuasted for Corporation %d", corporation.model.ID)
			return
		}
		response, err = p.ESI.GetCorporationsCorporationID(corporation.model)
		if err != nil {
			p.Logger.Errorf("Error completing request to ESI for Corporation %d information: %s", corporation.model.ID, err)
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		p.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting %d request again in 1 second", response.Code, response.Path, attempts)
		time.Sleep(1 * time.Second)
	}

	corporation.model = response.Data.(monocle.Corporation)

	if corporation.model.MemberCount == 0 {
		corporation.model.Closed = true
	}

	p.Logger.Debugf("Corporation: %d:%s\tNew Corporation: %t", corporation.model.ID, corporation.model.Name, !corporation.exists)

	switch !corporation.exists {
	case true:
		_, err := p.DB.InsertCorporation(corporation.model)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to insert new corporation into database: %s", err)
			return
		}
	case false:
		_, err := p.DB.UpdateCorporationByID(corporation.model)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to update corporation in database: %s", err)
			return
		}
	}
}

func (p *Processor) processCorporationAllianceHistory(corporation Corporation) {
	var history []monocle.CorporationAllianceHistory
	var response esi.Response

	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

	historyEtag, err := p.DB.SelectEtagByIDAndResource(corporation.model.ID, "corporation_alliance_history")
	historyEtag.Exists = true
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("Unable to query corporation_alliance_history etag resource for Character %d due to SQL Error: %s", corporation.model.ID, err)
			return
		}

		historyEtag.ID = corporation.model.ID
		historyEtag.Resource = "corporation_alliance_history"
		historyEtag.Exists = false
	}

	if !historyEtag.IsExpired() {
		return
	}

	p.Logger.Debugf("Processing CorpHistory: %d", historyEtag.ID)

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

	p.Logger.Debugf("Corporation History: %d\tNew Etag: %t", historyEtag.ID, historyEtag.Exists)

	existing, err := p.DB.SelectCorporationAllianceHistoryByID(historyEtag.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("Unable to query corporation_alliance_history etag resource for Character %d due to SQL Error: %s", historyEtag.ID, err)
			return
		}
	}

	diff := diffExistingCorpAlliHistory(existing, history)
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
