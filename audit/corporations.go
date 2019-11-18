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

func (a *Auditor) corpUpdater(c *cli.Context) {
	page := c.Int("page")

	limit := 25000
	for i := page; i <= 500; i++ {

		var corporations []monocle.Corporation
		offset := (i * limit) - limit
		a.Logger.WithFields(logrus.Fields{
			"page":   i,
			"limit":  limit,
			"offset": offset,
		}).Info("begin loop")
		query := boiler.Corporations(
			qm.Limit(limit),
			qm.Offset(offset),
			qm.Where("id >= 98000000"),
		)

		err := query.Bind(context.Background(), a.DB, &corporations)
		if err != nil {
			a.Logger.WithError(err).Error("error encountered querying database")
			return
		}

		if len(corporations) <= 0 {
			a.DGO.ChannelMessageSend("394991263344230411", "Done with Auditor, exiting script")
			a.Logger.Info("no corporations queried. done with job....exiting...")
			return
		}

		a.Logger.WithField("count", len(corporations)).Info("corporation query successful")

		corpChunk := tools.ChunkCorporationSlice(2500, corporations)
		for _, corporations := range corpChunk {
			wg.Add(1)

			go a.processCorporationChunk(corporations)
		}

		a.Logger.Info("waiting")
		wg.Wait()
		a.Logger.Info("done")
	}
}

func (a *Auditor) processCorporationChunk(corporations []monocle.Corporation) {
	var err error
	defer wg.Done()

	corpMap := tools.CorporationSliceToMap(corporations)
	corpIds := tools.CorpIDsFromCorpMap(corpMap)

	whereIds := []interface{}{}
	for _, id := range corpIds {
		whereIds = append(whereIds, id)
	}
	histories := []*monocle.CorporationAllianceHistory{}
	err = boiler.CorporationAllianceHistories(
		qm.WhereIn("id IN ?", whereIds...),
	).Bind(context.Background(), a.DB, &histories)
	if err != nil {
		a.Logger.Fatalf("Failed to query history for group of ids")
		return
	}

	x := make(map[uint64][]*monocle.CorporationAllianceHistory)
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
			a.Logger.WithField("id", histories[0].ID).Debug("1 history entry. skipping...")
			a.UpdateCorporation(histories[0].ID)
			continue
		}
		a.Logger.WithField("id", histories[0].ID).Debug("working history")

		for i, history := range histories {
			if i != historiesLen-1 {
				if history.LeaveDate.Valid {
					continue
				}
				j := i + 1
				history.LeaveDate.SetValid(histories[j].StartDate)
			}
			var boilHistory boiler.CorporationAllianceHistory
			err = copier.Copy(&boilHistory, history)
			if err != nil {
				a.Logger.WithFields(logrus.Fields{
					"id":        history.ID,
					"record_id": history.RecordID,
				}).WithError(err).Error("Failed to copy history to boilHistory")
				continue
			}

			err = boilHistory.Update(context.Background(), a.DB, boil.Infer())
			if err != nil {
				a.Logger.WithFields(logrus.Fields{
					"id":        history.ID,
					"record_id": history.RecordID,
				}).WithError(err).Error("Failed to update history in database")
			}

		}
		a.Logger.WithField("id", histories[0].ID).Debug("Done")
		a.UpdateCorporation(histories[0].ID)
	}
}

func (a *Auditor) UpdateCorporation(id uint64) {

	corporation, err := boiler.FindCorporation(context.Background(), a.DB, uint(id))
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"id": id,
		}).WithError(err).Error("unable to update corporation")
	}

	corporation.UpdatedAt = time.Now()

	err = corporation.Update(context.Background(), a.DB, boil.Infer())
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"id": id,
		}).WithError(err).Error("unable to update corporation")
	}
}
