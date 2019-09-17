package processor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

type Alliance struct {
	model  monocle.Alliance
	exists bool
}

func (p *Processor) alliHunter() {

	var value struct {
		Value uint64 `json:"value"`
	}

	kv, err := p.DB.SelectValueByKey("last_good_alliance_id")
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

	for x := begin; x <= 99999999; x++ {
		msg := fmt.Sprintf("Errors: %d Remaining: %d ID: %d", p.ESI.Remain, p.ESI.Reset, x)
		p.Logger.CriticalF("%s", msg)
		attempts := 0
		for {
			if attempts >= 2 {
				// Overriding Sleep to be more appropriate for how often alliances are created
				sleep := 60
				msg := fmt.Sprintf("Head Requests to %d failed. Sleep for %d minutes before trying again", x, sleep)
				p.Logger.Errorf("%s", msg)
				// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
				// p.DGO.ChannelMessageSend("394991263344230411", msg)

				time.Sleep(time.Minute * time.Duration(sleep))
				attempts = 0
			}
			p.Logger.DebugF("Checking for validity of %d", x)
			response, err := p.ESI.HeadAlliancesAllianceID(uint(x))
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

			alliance := Alliance{
				model: monocle.Alliance{
					ID: uint64(x),
				},
				exists: false,
			}
			p.processAlliance(alliance)
			p.processAllianceCorporationMembers(alliance)
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

func (p *Processor) processAlliance(alliance Alliance) {
	var response esi.Response
	var err error

	if p.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
		p.Logger.Errorf("\t%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		// p.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

	if !alliance.model.IsExpired() {
		return
	}

	p.Logger.Debugf("\tProcessing Alliance %d", alliance.model.ID)

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.Errorf("All Attempts exhuasted for Alliance %d", alliance.model.ID)
			return
		}
		response, err = p.ESI.GetAlliancesAllianceID(alliance.model)
		if err != nil {
			p.Logger.Errorf("Error completing request to ESI for Alliance %d information: %s", alliance.model.ID, err)
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		p.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting %d request again in 1 second", response.Code, response.Path, attempts)
		time.Sleep(1 * time.Second)
	}

	alliance.model = response.Data.(monocle.Alliance)

	p.Logger.Debugf("\tAlliance: %d:%s\tNew Alliance: %t", alliance.model.ID, alliance.model.Name, !alliance.exists)

	switch !alliance.exists {
	case true:
		_, err := p.DB.InsertAlliance(alliance.model)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to insert new character into database: %s", err)
			return
		}
	case false:
		_, err := p.DB.UpdateAllianceByID(alliance.model)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to update character in database: %s", err)
			return
		}
	}
}

func (p *Processor) processAllianceCorporationMembers(alliance Alliance) {
	return
}
