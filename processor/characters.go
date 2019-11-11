package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/esi"
	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type Character struct {
	model  *monocle.Character
	exists bool
}

type CharacterCorporationHistory struct {
	model  []*monocle.CharacterCorporationHistory
	exists bool
}

func (p *Processor) charHunter() {
	var value struct {
		Value uint64 `json:"value"`
	}

	const key = "last_good_character_id"

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

	p.Logger.WithField("id", begin).WithField("method", "charHunter").Info("starting...")

	for x := begin; x <= 2147483647; x += records {

		end := x + records
		p.Logger.WithFields(logrus.Fields{
			"errors":    p.ESI.Remain,
			"remaining": p.ESI.Reset,
			"loop":      x,
			"end":       end,
		}).Info()

		attempts := 0
		for {
			if attempts >= 2 {
				p.Logger.WithFields(logrus.Fields{
					"end":      end,
					"sleep":    sleep,
					"attempts": attempts,
				}).Error("head requests failed. sleeping...")
				time.Sleep(time.Minute * time.Duration(sleep))
				attempts = 0
			}

			p.Logger.WithField("end", end).Info("checking for valid end")

			response, err := p.ESI.HeadCharactersCharacterID(uint64(end))
			if err != nil {
				p.Logger.WithError(err).Error("head request for character failed")
				time.Sleep(time.Second * time.Duration(sleep))
				attempts = 3
				continue
			}

			if response.Code >= 500 {
				time.Sleep(time.Second * time.Duration(sleep))
				attempts++
				continue
			}
			break
		}
		p.Logger.WithFields(logrus.Fields{
			"end":     end,
			"current": x,
		}).Info("id is valid, starting loop")

		for y := x; y <= end; y++ {
			p.SleepDuringDowntime(time.Now())
			p.EvaluateESIArtifacts()
			wg.Add(1)

			character := monocle.Character{ID: uint64(y)}
			go func(model monocle.Character) {
				character := &Character{
					model:  &model,
					exists: false,
				}
				p.processCharacter(character)
				p.processCharacterCorpHistory(character)
				wg.Done()
			}(character)
		}

		p.Logger.Info("Done Dispatching. Waiting for Completion")
		wg.Wait()

		value.Value = end

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

		p.Logger.WithField("sleep", sleep).Info("loop complete. entering sleep period")
		time.Sleep(time.Second * time.Duration(sleep))
	}
}

func (p *Processor) charUpdater() {
	for {
		var characters []monocle.Character
		p.Logger.WithFields(logrus.Fields{
			"errors":    p.ESI.Remain,
			"remaining": p.ESI.Reset,
		}).Debug()
		p.SleepDuringDowntime(time.Now())
		p.EvaluateESIArtifacts()
		p.Logger.Info("starting loop...")

		err := boiler.Characters(
			qm.Where(boiler.CharacterColumns.Expires+"<NOW()"),
			qm.And(boiler.CharacterColumns.Ignored+"=?", 0),
			qm.OrderBy(boiler.CharacterColumns.Expires),
			qm.Limit(int(records*workers)),
		).Bind(context.Background(), p.DB, &characters)
		if err != nil {
			p.Logger.WithError(err).Error("no records returned from database")
			time.Sleep(time.Minute * 1)
			p.Logger.Debug("continuing loop")
			continue
		}

		if len(characters) <= 0 {
			temp := sleep * 30
			p.Logger.WithField("sleep", temp).Info("no characters queried. sleeping...")
			time.Sleep(time.Second * time.Duration(temp))
			p.Logger.Debug("continuing loop")
			continue
		}

		p.Logger.WithField("count", len(characters)).Info("character query successful")

		charChunk := chunkCharacterSlice(int(records), characters)

		for _, characters := range charChunk {
			wg.Add(1)

			go p.processCharacterChunk(characters)

		}

		p.Logger.Debug("waiting")
		wg.Wait()
		sleep := time.Second * 1
		p.Logger.WithField("sleep", sleep).Debug("done. sleeping...")
		time.Sleep(sleep)
	}
}

func (p *Processor) processCharacterChunk(characters []monocle.Character) {
	var response esi.Response
	var err error
	defer wg.Done()

	p.SleepDuringDowntime(time.Now())
	p.EvaluateESIArtifacts()

	charMap := characterSliceToMap(characters)
	charIds := charIdsFromSlice(characters)

	attempts := 1
	for {
		if attempts >= 3 {
			p.Logger.WithField("attempts", attempts).Error("all attempts exhausted")
			return
		}
		response, err = p.ESI.PostCharactersAffiliation(charIds)
		if err != nil {
			p.Logger.WithField("count", len(charIds)).WithError(err).Error("PostCharacterAffiliation request failed")
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		sleep := time.Second * 1
		p.Logger.WithFields(logrus.Fields{
			"code":    response.Code,
			"path":    response.Path,
			"attempt": attempts,
			"method":  "processCharacterChunk",
			"sleep":   sleep,
		}).Error("request failed. attempting again after sleeping.")
		time.Sleep(sleep)
	}

	affiliations := response.Data.([]monocle.CharacterAffiliation)
	updated := []monocle.Character{}
	stale := []interface{}{}
	args := []string{}
	for _, affiliation := range affiliations {

		selected := charMap[affiliation.CharacterID]
		updated = append(updated, selected)

		switch {
		case affiliation.CorporationID != selected.CorporationID,
			affiliation.AllianceID.Uint != selected.AllianceID.Uint,
			affiliation.FactionID.Uint != selected.FactionID.Uint:
			updated = append(updated, selected)
		default:
			stale = append(stale, selected.ID)
			args = append(args, "?")
		}
	}

	for _, model := range updated {
		character := &Character{
			model:  &model,
			exists: true,
		}
		p.processCharacter(character)
		p.processCharacterCorpHistory(character)
	}

	if len(stale) > 0 {
		query := `UPDATE characters SET expires = '%v' WHERE id IN (%s)`

		t := time.Now().Add(time.Hour * 12).Format("2006-01-02 15:04:05")

		query = fmt.Sprintf(query, t, strings.Join(args, ", "))
		_, err = p.DB.Exec(query, stale...)
		if err != nil {
			p.Logger.WithError(err).WithField("stale_records", len(stale)).Error("build update of etag expiry failed")
		}
	}
}

func (p *Processor) processCharacter(character *Character) {
	var response esi.Response
	var err error

	p.SleepDuringDowntime(time.Now())
	p.EvaluateESIArtifacts()

	if !character.model.IsExpired() {
		return
	}
	p.Logger.WithField("id", character.model.ID).Debug("processing char")

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.WithField("id", character.model.ID).Info("all attempts exhuasted for character")
			return
		}
		response, err = p.ESI.GetCharactersCharacterID(character.model)
		if err != nil {
			p.Logger.WithField("id", character.model.ID).WithError(err).Error("error completing esi request for character")
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

	character.model = response.Data.(*monocle.Character)
	if character.model.CorporationID == 1000001 {
		character.model.Ignored = true
	}

	p.Logger.WithFields(logrus.Fields{
		"id":     character.model.ID,
		"name":   character.model.Name,
		"exists": character.exists,
	}).Debug()

	switch character.exists {
	case true:
		err := p.DB.UpdateCharacterByID(character.model)
		if err != nil {
			p.Logger.WithError(err).WithField("id", character.model.ID).Error("update query failed for character")
			return
		}
	case false:
		err := p.DB.InsertCharacter(character.model)
		if err != nil {
			p.Logger.WithError(err).WithField("id", character.model.ID).Error("insert query failed for character")
			return
		}
	}

}

func (p *Processor) processCharacterCorpHistory(character *Character) {
	var err error
	var history = CharacterCorporationHistory{
		exists: true,
	}

	var etag = EtagResource{
		model:  &monocle.EtagResource{},
		exists: true,
	}

	var response esi.Response

	p.SleepDuringDowntime(time.Now())
	p.EvaluateESIArtifacts()

	err = boiler.Etags(
		qm.Where("id = ?", uint64(character.model.ID)),
		qm.And("resource = ?", "character_corporation_history"),
	).Bind(context.Background(), p.DB, etag.model)
	if err != nil {
		if err != sql.ErrNoRows {
			p.Logger.
				WithError(err).
				WithFields(logrus.Fields{"id": character.model.ID, "resource": "character_corporation_history"}).
				Error("etag query failed")
			return
		}

		etag.model.ID = character.model.ID
		etag.model.Resource = "character_corporation_history"
		etag.exists = false
	}

	p.Logger.WithField("id", character.model.ID).Debug("processing char corp history")

	attempts := 0
	for {
		if attempts >= 3 {
			p.Logger.WithFields(logrus.Fields{"id": character.model.ID, "attempts": attempts}).Error("all attempts exhausted")
			return
		}
		response, err = p.ESI.GetCharactersCharacterIDCorporationHistory(etag.model)
		if err != nil {
			p.Logger.WithField("id", etag.model.ID).WithError(err).Error("error completing esi request for character corporation history")
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		sleep := time.Second * 1

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

	if _, ok := data["history"].([]monocle.CharacterCorporationHistory); !ok {
		// Log an error
		return
	}
	history.model = data["history"].([]*monocle.CharacterCorporationHistory)

	var boilHistory boiler.CharacterCorporationHistorySlice
	err = copier.Copy(&boilHistory, &history.model)
	if err != nil {
		// Log error
		return
	}

	existing, err := boiler.CharacterCorporationHistories(
		qm.Where("id = ?", etag.model.ID),
	).All(context.Background(), p.DB)
	if err != nil {
		p.Logger.
			WithError(err).
			WithFields(logrus.Fields{"id": character.model.ID, "resource": "character_corporation_history"}).
			Error("history query for character corporation history failed")
		return
	}
	// If we don't know about this characters history at all, loop through the records, perform an insert, and then return
	if len(existing) == 0 {
		sort.Slice(boilHistory, func(i, j int) bool {
			return boilHistory[i].RecordID < boilHistory[j].RecordID
		})
		lastIndex := len(boilHistory) - 1
		for index, history := range boilHistory {
			history.ID = character.model.ID
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
		p.findUnknownCorps(boilHistory)
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
		selected.ID = character.model.ID

		if selected.LeaveDate.Valid {
			continue
		}

		nextIndex := index + 1

		if nextIndex <= lastIndexOfNew {
			selected.LeaveDate.SetValid(boilHistory[nextIndex].StartDate)
		}

		if insert {
			err = selected.Insert(context.Background(), p.DB, boil.Infer())
			if err != nil {
				p.Logger.WithError(err).WithFields(logrus.Fields{
					"id":     selected.ID,
					"record": selected.RecordID,
				}).Error("unable to insert character corporation history record into database")
			}
		} else {
			err = selected.Update(context.Background(), p.DB, boil.Infer())
			if err != nil {
				p.Logger.WithError(err).WithFields(logrus.Fields{
					"id":     selected.ID,
					"record": selected.RecordID,
				}).Error("unable to update character corporation history record in database")
			}
		}
		time.Sleep(time.Millisecond * 100)

	}

}

func (p *Processor) findUnknownCorps(histories boiler.CharacterCorporationHistorySlice) {

	unique := map[uint32]bool{}
	list := []interface{}{}

	for _, history := range histories {
		if _, v := unique[uint32(history.CorporationID)]; !v {
			unique[uint32(history.CorporationID)] = true
			list = append(list, uint32(history.CorporationID))
		}
	}

	knowns := make([]monocle.Corporation, 0)
	err := boiler.Corporations(
		qm.WhereIn("id IN ?", list...),
	).Bind(context.Background(), p.DB, &knowns)
	if err != nil {
		p.Logger.WithError(err).Error("known corporation query failed")
		return
	}

	for _, known := range knowns {
		if v := unique[known.ID]; v {
			delete(unique, known.ID)
		}
	}

	if len(unique) == 0 {
		return
	}

	for i := range unique {
		corporation := &Corporation{
			model: &monocle.Corporation{
				ID: i,
			},
			exists: false,
		}
		p.processCorporation(corporation)
	}

}
