package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ddouglas/monocle/boiler"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

type Alliance struct {
	model  *monocle.Alliance
	exists bool
}

func (p *Processor) alliHunter() {

	var value struct {
		Value uint64 `json:"value"`
	}

	const key = "last_good_alliance_id"

	kv, err := boiler.KVS(
		qm.Where("k = ?", key),
	).One(context.Background(), p.DB)

	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.WithError(err).WithField("id", key).Error("query for key failed")
			return
		}
	}

	err = json.Unmarshal(kv.V, &value)
	if err != nil {
		p.Logger.WithError(err).WithField("value", kv.V).Error("unable to unmarshal value into struct")
		return
	}

	begin = value.Value
	p.Logger.WithField("id", begin).WithField("method", "alliHunter").Info("starting...")

	for x := begin; x <= 99999999; x++ {

		p.Logger.WithFields(logrus.Fields{
			"errors":    p.ESI.Remain,
			"remaining": p.ESI.Reset,
			"loop":      x,
		}).Info()

		attempts := 0
		for {
			if attempts >= 2 {
				sleep := 60

				p.Logger.WithFields(logrus.Fields{
					"sleep":    sleep,
					"attempts": attempts,
				}).Error("head requests failed. sleeping...")

				time.Sleep(time.Minute * time.Duration(sleep))
				attempts = 0
			}

			p.Logger.WithField("id", x).Debug("head request for id")
			response, err := p.ESI.HeadAlliancesAllianceID(uint(x))
			if err != nil {
				p.Logger.WithError(err).WithField("sleep", sleep).WithField("attempts", attempts).Error("head request returned error. sleeping...")
				time.Sleep(time.Second * time.Duration(sleep))
				attempts = 3
				continue
			}

			if response.Code >= 500 {
				p.Logger.WithError(err).WithField("sleep", sleep).WithField("attempts", attempts).Error("head request returned error. sleeping...")
				attempts++
				time.Sleep(time.Second * time.Duration(sleep))
				continue
			}

			alliance := &Alliance{
				model: &monocle.Alliance{
					ID: uint(x),
				},
				exists: false,
			}
			p.processAlliance(alliance)
			p.processAllianceCorporationMembers(alliance)
			break
		}

		value.Value = x

		kv.V, err = json.Marshal(value)
		if err != nil {
			p.Logger.WithError(err).WithField("value", value).Error("unable to marshal value")
			return
		}

		err = kv.Update(context.Background(), p.DB, boil.Infer())
		if err != nil {
			p.Logger.WithError(err).WithField("key", kv).Error("unable to update database record")
			return
		}

		p.Logger.WithField("sleep", sleep).Info("loop complete. sleeping...")
		time.Sleep(time.Second * time.Duration(sleep))
	}

}

func (p *Processor) alliUpdater() {
	sleep = 1800
	for {
		var alliances []monocle.Alliance
		p.Logger.WithFields(logrus.Fields{
			"errors":    p.ESI.Remain,
			"remaining": p.ESI.Reset,
		}).Debug()

		p.SleepDuringDowntime(time.Now())
		p.EvaluateESIArtifacts()

		query := boiler.Alliances(
			qm.Where(boiler.AllianceColumns.Expires+"<NOW()"),
			qm.And(boiler.AllianceColumns.Ignored+"=?", 0),
			qm.And(boiler.AllianceColumns.Closed+"=?", 0),
			qm.OrderBy(boiler.AllianceColumns.Expires),
			qm.Limit(int(records*workers)),
		)

		err := query.Bind(context.Background(), p.DB, &alliances)
		if err != nil {
			if err != sql.ErrNoRows {
				p.Logger.WithError(err).Error("unable to query for alliances")
			}
			continue
		}

		if len(alliances) == 0 {
			temp := sleep * 30
			p.Logger.WithField("sleep", temp).Info("no alliances queried. sleeping...")
			time.Sleep(time.Minute * time.Duration(temp))
			p.Logger.Debug("continuing loop")
			continue
		}

		p.Logger.WithField("count", len(alliances)).Info("alliances query successful")

		alliChunk := chunkAllianceSlice(int(records), alliances)

		for _, alliances := range alliChunk {
			wg.Add(1)
			go func(alliances []monocle.Alliance) {
				for _, model := range alliances {
					alliance := &Alliance{
						model:  &model,
						exists: true,
					}
					p.processAlliance(alliance)
					p.processAllianceCorporationMembers(alliance)
				}
				wg.Done()
			}(alliances)
		}

		p.Logger.Debug("waiting")
		wg.Wait()
		sleep := time.Second * 1
		p.Logger.WithField("sleep", sleep).Debug("done. sleeping...")
		time.Sleep(sleep)
	}
}

func (p *Processor) processAlliance(alliance *Alliance) {
	var response esi.Response
	var err error

	p.SleepDuringDowntime(time.Now())
	p.EvaluateESIArtifacts()

	if !alliance.model.IsExpired() {
		return
	}

	p.Logger.WithField("id", alliance.model.ID).Debug("processing alliance")

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.WithField("id", alliance.model.ID).Info("all attempts exhuasted for character")
			return
		}
		response, err = p.ESI.GetAlliancesAllianceID(alliance.model)
		if err != nil {
			p.Logger.WithField("id", alliance.model.ID).WithError(err).Error("error completing esi request for alliance")
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		sleep := 1 * time.Second
		p.Logger.WithFields(logrus.Fields{
			"code":     response.Code,
			"path":     response.Path,
			"attempts": attempts,
			"sleep":    sleep,
		}).Info("received bad response code for request. sleeping for trying again...")
		time.Sleep(sleep)
	}

	alliance.model = response.Data.(*monocle.Alliance)

	p.Logger.WithFields(logrus.Fields{
		"id":     alliance.model.ID,
		"name":   alliance.model.Name,
		"exists": alliance.exists,
	}).Debug()

	switch alliance.exists {
	case true:
		err := p.DB.UpdateAllianceByID(alliance.model)
		if err != nil {
			p.Logger.WithError(err).WithField("id", alliance.model.ID).Error("update query failed for alliance")
			return
		}
	case false:
		err := p.DB.InsertAlliance(alliance.model)
		if err != nil {
			p.Logger.WithError(err).WithField("id", alliance.model.ID).Error("insert query failed for alliance")
			return
		}
	}
}

func (p *Processor) processAllianceCorporationMembers(alliance *Alliance) {
	var response esi.Response
	var err error

	p.SleepDuringDowntime(time.Now())
	p.EvaluateESIArtifacts()

	if !alliance.model.IsExpired() {
		return
	}

	var etag = &EtagResource{
		exists: true,
	}

	err = boiler.Etags(
		qm.Where("id = ?", uint64(alliance.model.ID)),
		qm.And("resource = ?", "alliance_corporation_members"),
	).Bind(context.Background(), p.DB, etag.model)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.
				WithError(err).
				WithFields(logrus.Fields{"id": alliance.model.ID, "resource": "alliance_corporation_members"}).
				Error("etag query failed")
			return
		}
		etag.model.ID = uint64(alliance.model.ID)
		etag.model.Resource = "alliance_corporation_members"
		etag.exists = false
	}

	p.Logger.WithField("id", etag.model.ID).Debug("process alliance corporation members")

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.WithFields(logrus.Fields{"id": etag.model.ID, "attempts": attempts}).Error("all attempts exhausted")
			return
		}
		response, err = p.ESI.GetAlliancesAllianceIDCorporations(etag.model)
		if err != nil {
			p.Logger.WithField("id", etag.model.ID).WithError(err).Error("error completing esi request for alliance corporation members")
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		sleep := 1 * time.Second
		p.Logger.WithFields(logrus.Fields{
			"code":    response.Code,
			"path":    response.Path,
			"attempt": attempts,
			"sleep":   sleep,
		}).Error("request failed. attempting again after sleeping.")
		time.Sleep(sleep)
	}

	data := response.Data.(map[string]interface{})
	if _, ok := data["ids"]; !ok {
		p.Logger.WithField("data", data).Error("expected key ids missing from response")
		return
	}

	if _, ok := data["etag"]; !ok {
		p.Logger.WithField("data", data).Error("expected key etag missing from response")
		return
	}

	etag.model = data["etag"].(*monocle.EtagResource)

	switch etag.exists {
	case true:
		err := p.DB.UpdateEtagByIDAndResource(etag.model)
		if err != nil {
			p.Logger.WithError(err).
				WithField("id", etag.model.ID).
				WithField("resource", etag.model.Resource).
				WithField("etag", etag.model.Etag).
				Error("update query failed for etag resource")
			return
		}
	case false:
		err := p.DB.InsertEtag(etag.model)
		if err != nil {
			p.Logger.WithError(err).
				WithField("id", etag.model.ID).
				WithField("resource", etag.model.Resource).
				WithField("etag", etag.model.Etag).
				Error("insert query failed for etag resource")
			return
		}
	}

	ids := data["ids"].([]uint)

	member_count := 0
	if len(ids) > 0 {
		idInterface := []interface{}{}
		for _, corpId := range ids {
			idInterface = append(idInterface, corpId)
		}

		corporations := make([]*monocle.Corporation, 0)

		err = boiler.Corporations(
			qm.WhereIn("id IN ?", idInterface...),
		).Bind(context.Background(), p.DB, &corporations)
		if err != nil {
			p.Logger.WithError(err).WithField("id", idInterface).Error("failed to query corporations for alliance")
			return
		}

		for _, corp := range corporations {
			member_count += int(corp.MemberCount)
		}
	}

	alliance.model.MemberCount = uint(member_count)
	alliance.model.UpdatedAt = time.Now()
	err = p.DB.UpdateAllianceByID(alliance.model)
	if err != nil {
		p.Logger.WithError(err).WithField("id", alliance.model.ID).Error("failed to update member count for alliance")
		return
	}

}
