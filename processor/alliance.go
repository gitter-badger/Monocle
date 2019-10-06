package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/monocle/boiler"
	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

type Alliance struct {
	model  monocle.Alliance
	exists bool
}

type AllianceCorporationMembers struct {
	model  monocle.EtagResource
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
					ID: uint32(x),
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

func (p *Processor) alliUpdater() {
	sleep = 1800
	for {
		var alliances []monocle.Alliance

		p.Logger.DebugF("Current Error Count: %d Remain: %d", p.ESI.Remain, p.ESI.Reset)
		if p.ESI.Remain < 20 {
			p.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", p.ESI.Reset)
			time.Sleep(time.Second * time.Duration(p.ESI.Reset))
		}

		err := boiler.Alliances(
			qm.Where(boiler.AllianceColumns.Expires+"<NOW()"),
			qm.And(boiler.AllianceColumns.Ignored+"=?", 0),
			qm.And(boiler.AllianceColumns.Closed+"=?", 0),
			qm.OrderBy(boiler.AllianceColumns.Expires),
			qm.Limit(int(records*workers)),
		).Bind(context.Background(), p.DB, &alliances)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.Fatalf("Unable to query for alliances: %s", err)
			}
			continue
		}

		if len(alliances) == 0 {
			p.Logger.Infof("No alliances were queried. Sleeping for %d seconds", sleep)
			time.Sleep(time.Second * time.Duration(sleep))
			continue
		}

		p.Logger.Infof("Successfully Queried %d Corporations", len(alliances))

		alliChunk := chunkAllianceSlice(int(records), alliances)

		for _, alliances := range alliChunk {
			wg.Add(1)
			go func(alliances []monocle.Alliance) {
				for _, model := range alliances {
					alliance := Alliance{
						model:  model,
						exists: true,
					}
					p.processAlliance(alliance)
					p.processAllianceCorporationMembers(alliance)
				}
				wg.Done()
			}(alliances)
		}

		p.Logger.Info("Waiting")
		wg.Wait()
		p.Logger.Info("Done")
	}
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

	var etagModel AllianceCorporationMembers
	etagModel.exists = true
	etagModel.model, err = p.DB.SelectEtagByIDAndResource(uint64(alliance.model.ID), "alliance_corporation_members")
	if err != nil {
		etagModel.model.ID = uint64(alliance.model.ID)
		etagModel.model.Resource = "alliance_corporation_members"
		etagModel.exists = false
	}

	p.Logger.Debugf("\tProcessing Alliance %d", alliance.model.ID)

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.Errorf("All Attempts exhuasted for Alliance %d", alliance.model.ID)
			return
		}
		response, err = p.ESI.GetAlliancesAllianceIDCorporations(etagModel.model)
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

	data := response.Data.(map[string]interface{})
	if _, ok := data["ids"]; !ok {
		p.Logger.Error("Expected Key ids missing from response")
		return
	}

	if _, ok := data["etag"]; !ok {
		p.Logger.Error("Expected Key etag missing from response")
		return
	}

	etagModel.model = data["etag"].(monocle.EtagResource)

	switch !etagModel.exists {
	case true:
		_, err := p.DB.InsertEtag(etagModel.model)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to insert new character into database: %s", err)
			return
		}
	case false:
		_, err := p.DB.UpdateEtagByIDAndResource(etagModel.model)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to update character in database: %s", err)
			return
		}
	}

	ids := data["ids"].([]uint32)

	idInterface := []interface{}{}
	for _, corpId := range ids {
		idInterface = append(idInterface, corpId)
	}

	corporations := make([]*monocle.Corporation, 0)

	err = boiler.Corporations(
		qm.WhereIn("id IN ?", idInterface...),
	).Bind(context.Background(), p.DB, &corporations)
	if err != nil {
		return
	}

	member_count := 0
	for _, corp := range corporations {
		member_count += int(corp.MemberCount)
	}

	alliance.model.MemberCount = uint32(member_count)
	_, err = p.DB.UpdateAllianceByID(alliance.model)
	if err != nil {
		p.Logger.Errorf("Unable to update member count from alliance %d", alliance.model.ID)
		return
	}

	return
}
