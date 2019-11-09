package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/esi"
	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type Corporation struct {
	model  *monocle.Corporation
	exists bool
}

type CorporationAllianceHistory struct {
	model  []*monocle.CorporationAllianceHistory
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

		p.SleepDuringDowntime(time.Now())
		p.EvaluateESIArtifacts()

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

	p.SleepDuringDowntime(time.Now())
	p.EvaluateESIArtifacts()

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
	var err error
	var history = &CorporationAllianceHistory{
		exists: true,
	}

	var etag = &EtagResource{
		model:  &monocle.EtagResource{},
		exists: true,
	}

	var response esi.Response

	p.SleepDuringDowntime(time.Now())
	p.EvaluateESIArtifacts()

	const resource = "corporation_alliance_history"

	err = boiler.Etags(
		qm.Where("id = ?", uint64(corporation.model.ID)),
		qm.And("resource = ?", resource),
	).Bind(context.Background(), p.DB, etag.model)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.
				WithError(err).
				WithFields(logrus.Fields{"id": corporation.model.ID, "resource": resource}).
				Error("etag query failed")
			return
		}

		etag.model.ID = uint64(corporation.model.ID)
		etag.model.Resource = resource
		etag.exists = false
	}

	if !etag.model.IsExpired() {
		return
	}

	p.Logger.WithField("id", etag.model.ID).Debug("processing corporation alliance history")

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

	if response.Data == nil {
		p.Logger.WithFields(logrus.Fields{
			"code":    response.Code,
			"path":    response.Path,
			"attempt": attempts,
		}).Error("Data property on response is nil")
		return
	}

	if _, ok := response.Data.(map[string]interface{}); !ok {
		p.Logger.WithFields(logrus.Fields{
			"code":    response.Code,
			"path":    response.Path,
			"attempt": attempts,
			"data":    response.Data,
		}).Error("data is not of expected type")
		return
	}

	data := response.Data.(map[string]interface{})
	if _, ok := data["history"]; !ok {
		p.Logger.WithField("data", data).Error("expected key history missing from response")
		return
	}

	if _, ok := data["etag"]; !ok {
		p.Logger.WithField("data", data).Error("expected key etag missing from response")
		return
	}

	if _, ok := data["etag"].(*monocle.EtagResource); !ok {
		p.Logger.WithFields(logrus.Fields{
			"code":    response.Code,
			"path":    response.Path,
			"attempt": attempts,
		}).Error("Etag Index on response is not set.")
		return
	}
	etag.model = data["etag"].(*monocle.EtagResource)

	var boilEtag boiler.Etag
	err = copier.Copy(&boilEtag, &etag.model)
	if err != nil {
		// Log an error
		return
	}

	switch etag.exists {
	case true:
		err = boilEtag.Update(context.Background(), p.DB, boil.Infer())
		if err != nil {
			p.Logger.WithError(err).
				WithField("id", etag.model.ID).
				WithField("resource", etag.model.Resource).
				WithField("etag", etag.model.Etag).
				Error("update query failed for history etag")
			return
		}
	case false:
		err = boilEtag.Insert(context.Background(), p.DB, boil.Infer())
		if err != nil {
			p.Logger.WithError(err).
				WithField("id", etag.model.ID).
				WithField("resource", etag.model.Resource).
				WithField("etag", etag.model.Etag).
				Error("insert query failed for history etag")
			return
		}
	}

	// If the response code is a not a 200, there is no new data, so return here
	if response.Code > 200 {
		return
	}

	if _, ok := data["history"].([]monocle.CorporationAllianceHistory); !ok {
		// Log an error
		return
	}
	history.model = data["history"].([]*monocle.CorporationAllianceHistory)

	var boilHistory boiler.CorporationAllianceHistorySlice
	err = copier.Copy(&boilHistory, &history.model)
	if err != nil {
		// Log error
		return
	}

	existing, err := boiler.CorporationAllianceHistories(
		qm.Where("id = ?", etag.model.ID),
	).All(context.Background(), p.DB)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.Errorf("Unable to query corporation_alliance_history etag resource for Character %d due to SQL Error: %s", etag.model.ID, err)
			return
		}
	}

	if len(existing) == 0 {
		sort.Slice(boilHistory, func(i, j int) bool {
			return boilHistory[i].RecordID < boilHistory[j].RecordID
		})

		lastIndex := len(boilHistory) - 1
		for index, history := range boilHistory {
			history.ID = uint64(corporation.model.ID)
			if index < lastIndex {
				history.LeaveDate.SetValid(boilHistory[index+1].StartDate)
			}

			err = history.Insert(context.Background(), p.DB, boil.Infer())
			if err != nil {
				p.Logger.WithError(err).WithFields(logrus.Fields{
					"id":     history.ID,
					"record": history.RecordID,
				}).Error("unable to insert character corporation history record into database")
			}
			time.Sleep(time.Millisecond * 100)
		}
		return
	}

	if len(existing) < len(boilHistory) {
		err = p.findUnknownAlliances(boilHistory)
		if err != nil {
			p.Logger.WithError(err).Error("known corporation query failed")
			return
		}
	}

	sort.Slice(existing, func(i, j int) bool {
		return existing[i].RecordID < existing[j].RecordID
	})

	sort.Slice(boilHistory, func(i, j int) bool {
		return boilHistory[i].RecordID < boilHistory[j].RecordID
	})

	exLen := len(existing)
	neLen := len(boilHistory)
	lastIndexOfExisting := exLen - 1
	lastIndexOfNew := neLen - 1
	insert := false
	for index := range boilHistory {
		if index > lastIndexOfExisting {
			existing = append(existing, boilHistory[index])
			insert = true
		}

		selected := existing[index]

		if selected.LeaveDate.Valid {
			continue
		}

		nextIndex := index + 1

		if nextIndex <= lastIndexOfNew {
			selected.LeaveDate.SetValid(boilHistory[nextIndex].StartDate)
		}

		if insert {
			selected.ID = uint64(corporation.model.ID)
			err = selected.Insert(context.Background(), p.DB, boil.Infer())
			if err != nil {
				p.Logger.WithError(err).WithFields(logrus.Fields{
					"id":     selected.ID,
					"record": selected.RecordID,
				}).Error("unable to insert character corporation history record into database")
			} else {
				err = selected.Update(context.Background(), p.DB, boil.Infer())
				if err != nil {
					p.Logger.WithError(err).WithFields(logrus.Fields{
						"id":     selected.ID,
						"record": selected.RecordID,
					}).Error("unable to update character corporation history record in database")
				}
			}
		}
	}

}

func (p *Processor) findUnknownAlliances(histories boiler.CorporationAllianceHistorySlice) error {
	unique := map[uint64]bool{}
	list := []interface{}{}

	for _, history := range histories {
		if !history.AllianceID.Valid {
			continue
		}

		if _, v := unique[uint64(history.AllianceID.Uint)]; !v {
			unique[uint64(history.AllianceID.Uint)] = true
			list = append(list, uint64(history.AllianceID.Uint))
		}
	}

	knowns, err := boiler.Alliances(
		qm.WhereIn("id IN ?", list...),
	).All(context.Background(), p.DB)
	if err != nil {
		return err
	}

	for _, known := range knowns {
		if v := unique[known.ID]; v {
			delete(unique, known.ID)
		}
	}

	if len(unique) == 0 {
		return nil
	}

	for i := range unique {
		alliance := &Alliance{
			model: &monocle.Alliance{
				ID: uint32(i),
			},
			exists: false,
		}
		p.processAlliance(alliance)
	}

	return nil
}
