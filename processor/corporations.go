package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/esi"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type Corporation struct {
	model  *monocle.Corporation
	exists bool
}

type CorporationAllianceHistory struct {
	model  *monocle.EtagResource
	exists bool
}

func (p *Processor) corpHunter() {
	var value struct {
		Value uint64 `json:"value"`
	}

	const key = "last_good_corporation_id"

	kv, err := boiler.KVS(
		qm.Where("k = ?", key),
	).One(context.Background(), p.DB)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.WithError(err).WithField("key", key).Error("query for key failed")

			return
		}
	}

	err = json.Unmarshal(kv.V, &value)
	if err != nil {
		p.Logger.WithError(err).WithField("value", kv.V).Error("unable to unmarshal value into struct")
		return
	}

	begin = value.Value

	p.Logger.WithField("id", begin).WithField("method", "corpHunter").Info("starting...")

	for x := begin; x <= 98999999; x++ {

		p.Logger.WithFields(logrus.Fields{
			"errors":    p.ESI.Remain,
			"remaining": p.ESI.Reset,
			"loop":      x,
		}).Info()

		attempts := 0
		for {
			if attempts >= 2 {
				p.Logger.WithFields(logrus.Fields{
					"sleep":    sleep,
					"attempts": attempts,
				}).Error("head requests failed. sleeping...")
				time.Sleep(time.Minute * time.Duration(sleep))
				attempts = 0
			}

			p.Logger.WithField("id", x).Debug("head request for id")
			response, err := p.ESI.HeadCorporationsCorporationID(uint(x))
			if err != nil {
				p.Logger.WithError(err).WithField("sleep", sleep).WithField("attempts", attempts).Error("head request returned error. sleeping...")
				time.Sleep(time.Second * 5)
				attempts = 3
				continue
			}

			if response.Code >= 500 {
				p.Logger.WithError(err).WithField("sleep", sleep).WithField("attempts", attempts).Error("head request returned error. sleeping...")
				attempts++
				time.Sleep(time.Second * 5)
				continue
			}

			corporation := &Corporation{
				model: &monocle.Corporation{
					ID: uint32(x),
				},
				exists: false,
			}
			p.processCorporation(corporation)
			p.processCorporationAllianceHistory(corporation)
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

func (p *Processor) corpUpdater() {
	for {

		var corporations []monocle.Corporation
		p.Logger.WithFields(logrus.Fields{
			"errors":    p.ESI.Remain,
			"remaining": p.ESI.Reset,
		}).Debug()
		if p.ESI.Remain < 20 {
			p.Logger.WithFields(logrus.Fields{
				"errors":    p.ESI.Remain,
				"remaining": p.ESI.Reset,
			}).Error("error count is low. sleeping...")
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
				p.Logger.WithError(err).Error("unable to query for corporations")
			}
			continue
		}

		if len(corporations) <= 0 {
			temp := sleep * 10
			p.Logger.WithField("sleep", temp).Info("no corporations queried. sleeping...")
			time.Sleep(time.Minute * time.Duration(temp))
			p.Logger.Debug("continuing loop")
			continue
		}
		p.Logger.WithField("count", len(corporations)).Info("corporations query successful")

		corpChunk := chunkCorporationSlice(int(records), corporations)

		for _, corporations := range corpChunk {
			wg.Add(1)
			go func(corporations []monocle.Corporation) {
				for _, model := range corporations {
					corporation := &Corporation{
						model:  &model,
						exists: true,
					}
					p.processCorporation(corporation)
					p.processCorporationAllianceHistory(corporation)
				}
				wg.Done()
			}(corporations)
		}

		p.Logger.Debug("waiting")
		wg.Wait()
		sleep := time.Second * 1
		p.Logger.WithField("sleep", sleep).Debug("done. sleeping...")
		time.Sleep(sleep)
	}
}

func (p *Processor) processCorporation(corporation *Corporation) {

	var response esi.Response
	var err error
	if p.ESI.Remain < 20 {
		p.Logger.WithFields(logrus.Fields{
			"errors":    p.ESI.Remain,
			"remaining": p.ESI.Reset,
		}).Error("error count is low. sleeping...")
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

	if !corporation.model.IsExpired() {
		return
	}

	p.Logger.WithField("id", corporation.model.ID).Debug("processing corporation")

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.WithField("id", corporation.model.ID).Info("all attempts exhuasted for corporation")
			return
		}
		response, err = p.ESI.GetCorporationsCorporationID(corporation.model)
		if err != nil {
			p.Logger.WithField("id", corporation.model.ID).WithError(err).Error("error completing esi request for corporation")
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

	corporation.model = response.Data.(*monocle.Corporation)

	if corporation.model.MemberCount == 0 {
		corporation.model.Closed = true
	}

	p.Logger.WithFields(logrus.Fields{
		"id":     corporation.model.ID,
		"name":   corporation.model.Name,
		"exists": corporation.exists,
	}).Debug()

	switch corporation.exists {
	case true:
		err := p.DB.UpdateCorporationByID(corporation.model)
		if err != nil {
			p.Logger.WithError(err).WithField("id", corporation.model.ID).Error("update query failed for corporation")
			return
		}
	case false:
		err := p.DB.InsertCorporation(corporation.model)
		if err != nil {
			p.Logger.WithError(err).WithField("id", corporation.model.ID).Error("insert query failed for corporation")
			return
		}
	}
}

func (p *Processor) processCorporationAllianceHistory(corporation *Corporation) {
	var history []*monocle.CorporationAllianceHistory
	var response esi.Response

	if p.ESI.Remain < 20 {
		p.Logger.WithFields(logrus.Fields{
			"errors":    p.ESI.Remain,
			"remaining": p.ESI.Reset,
		}).Error("error count is low. sleeping...")
		time.Sleep(time.Second * time.Duration(p.ESI.Reset))
	}

	if !corporation.model.IsExpired() {
		return
	}

	var etag = &CorporationAllianceHistory{
		exists: true,
	}

	err := boiler.Etags(
		qm.Where("id = ?", uint64(corporation.model.ID)),
		qm.And("resource = ?", "corporation_alliance_history"),
	).Bind(context.Background(), p.DB, etag.model)

	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.
				WithError(err).
				WithFields(logrus.Fields{"id": corporation.model.ID, "resource": "corporation_alliance_history"}).
				Error("etag query failed")
			return
		}

		etag.model.ID = uint64(corporation.model.ID)
		etag.model.Resource = "corporation_alliance_history"
		etag.exists = false
	}

	if !etag.model.IsExpired() {
		return
	}

	p.Logger.WithField("id", etag.model.ID).Debug("process character corporation history")

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.WithFields(logrus.Fields{"id": etag.model.ID, "attempts": attempts}).Error("all attempts exhausted")
			return
		}
		response, err = p.ESI.GetCorporationsCorporationIDAllianceHistory(etag.model)
		if err != nil {
			p.Logger.WithField("id", etag.model.ID).WithError(err).Error("error completing esi request for character corporation history")
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
	if _, ok := data["history"]; !ok {
		p.Logger.WithField("data", data).Error("expected key history missing from response")
		return
	} else if _, ok := data["etag"]; !ok {
		p.Logger.WithField("data", data).Error("expected key etag missing from response")
		return
	}

	history = data["history"].([]*monocle.CorporationAllianceHistory)
	etag.model = data["etag"].(*monocle.EtagResource)

	p.Logger.WithFields(logrus.Fields{
		"id":     etag.model.ID,
		"exists": etag.model.Exists,
	}).Debug()

	var existing []*monocle.CorporationAllianceHistory
	err = boiler.CorporationAllianceHistories(
		qm.Where("id = ?", etag.model.ID),
	).Bind(context.Background(), p.DB, &existing)

	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("Unable to query corporation_alliance_history etag resource for Character %d due to SQL Error: %s", etag.model.ID, err)
			return
		}
	}

	diff := diffExistingCorpAlliHistory(existing, history)

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

	if len(diff) > 0 {
		err = p.DB.InsertCorporationAllianceHistory(etag.model.ID, diff)
		if err != nil {
			p.Logger.WithError(err).
				WithField("id", etag.model.ID).
				WithField("diffCount", diff).
				Error("insert query failed for corporation alliance history")
			return
		}
	}
}

func diffExistingCorpAlliHistory(a []*monocle.CorporationAllianceHistory, b []*monocle.CorporationAllianceHistory) []*monocle.CorporationAllianceHistory {
	c := convertCorpAlliHistToMap(a)
	d := convertCorpAlliHistToMap(b)
	result := make([]*monocle.CorporationAllianceHistory, 0)
	for recordID, history := range d {
		if _, ok := c[recordID]; !ok {
			result = append(result, history)
		}
	}
	return result
}

func convertCorpAlliHistToMap(a []*monocle.CorporationAllianceHistory) map[uint]*monocle.CorporationAllianceHistory {
	result := make(map[uint]*monocle.CorporationAllianceHistory)
	for _, history := range a {
		result[history.RecordID] = history
	}
	return result
}
