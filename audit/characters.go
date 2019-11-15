package audit

import (
	"context"
	"sort"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/tools"
	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func (a *Auditor) charUpdater(c *cli.Context) {
	page := c.Int("page") // 1
	end := c.Int("end")
	if end == 0 {
		// If end was not passed from the cli, set a default value
		//  instead of using the zero value of the type
		end = 1000
	}

	limit := 50000
	for i := page; i <= end; i++ {
		var characters []monocle.Character
		offset := (i * limit) - limit
		a.Logger.WithFields(logrus.Fields{
			"page":   i,
			"limit":  limit,
			"offset": offset,
		}).Info("begin loop")

		query := boiler.Characters(
			qm.Limit(limit),
			qm.Offset(offset),
		)

		err := query.Bind(context.Background(), a.DB, &characters)
		if err != nil {
			a.Logger.WithError(err).Error("error encountered querying database")
			return
		}

		if len(characters) <= 0 {
			a.DGO.ChannelMessageSend("394991263344230411", "Done with Auditor, exiting script")
			a.Logger.Info("no characters queried. done with job....exiting...")
			return
		}

		a.Logger.WithField("count", len(characters)).Info("character query successful")

		charChunk := tools.ChunkCharacterSlice(1000, characters)

		for _, characters := range charChunk {
			wg.Add(1)
			go a.processCharacterChunk(characters)
		}

		a.Logger.Info("waiting")
		wg.Wait()
		a.Logger.Info("done")
	}
}

func (a *Auditor) processCharacterChunk(characters []monocle.Character) {
	var err error
	defer wg.Done()

	charMap := tools.CharacterSliceToMap(characters)
	charIds := tools.CharIDsFromCharMap(charMap)
	whereIds := []interface{}{}
	for _, id := range charIds {
		whereIds = append(whereIds, id)
	}
	histories := []*monocle.CharacterCorporationHistory{}
	err = boiler.CharacterCorporationHistories(
		qm.WhereIn("id IN ?", whereIds...),
	).Bind(context.Background(), a.DB, &histories)
	if err != nil {
		a.Logger.Fatalf("Failed to query history for group of ids")
		return
	}

	x := make(map[uint64][]*monocle.CharacterCorporationHistory)
	for _, history := range histories {
		x[history.ID] = append(x[history.ID], history)
	}

	for id := range x {
		sort.Slice(x[id], func(i, j int) bool {
			return x[id][i].RecordID < x[id][j].RecordID
		})
	}

	for _, histories := range x {
		historiesLen := len(histories)
		if historiesLen == 1 {
			a.Logger.WithField("id", histories[0].ID).Debug("1 history entry, skipping....")
			a.UpdateCharacter(histories[0].ID)
			continue
		}
		for i, history := range histories {

			if i != historiesLen-1 {
				if history.LeaveDate.Valid {
					continue
				}
				j := i + 1
				history.LeaveDate.SetValid(histories[j].StartDate)
			}

			var boilHistory boiler.CharacterCorporationHistory
			err = copier.Copy(&boilHistory, history)
			if err != nil {
				a.Logger.WithFields(logrus.Fields{
					"id":        history.ID,
					"record_id": history.RecordID,
				}).WithError(err).Error("Failed to copier history to boilHistory")
				continue
			}

			err = boilHistory.Update(context.Background(), a.DB, boil.Infer())
			if err != nil {
				a.Logger.WithFields(logrus.Fields{
					"id":        history.ID,
					"record_id": history.RecordID,
				}).WithError(err).Error("Failed to copier history to boilHistory")
				continue
			}
		}
		a.Logger.WithField("id", histories[0].ID).Debug("Done")
		a.Logger.Debug("---------------")
		a.UpdateCharacter(histories[0].ID)
	}

}

func (a *Auditor) UpdateCharacter(id uint64) {

	character, err := boiler.FindCharacter(context.Background(), a.DB, id)
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"id": id,
		}).WithError(err).Error("unable to update character")
	}

	character.UpdatedAt = time.Now()

	err = character.Update(context.Background(), a.DB, boil.Infer())
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"id": id,
		}).WithError(err).Error("unable to update character")
	}
}
