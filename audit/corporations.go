package audit

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/davecgh/go-spew/spew"
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
	// page := c.Int("page")

	limit := 1
	i := 1
	for {

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

		// tools.OutputDebugQuery(query.Query)

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

		a.Logger.WithField("count", len(corporations)).Info("character query successful")

		corpChunk := tools.ChunkCorporationSlice(1000, corporations)
		for _, corporations := range corpChunk {
			// wg.Add(1)

			a.processCorporationChunk(corporations)
		}

		a.Logger.Info("waiting")
		// wg.Wait()
		a.Logger.Info("done")
		i++
	}
}

func (a *Auditor) processCorporationChunk(corporations []monocle.Corporation) {
	var err error
	// defer wg.Done()

	corpMap := tools.CorporationSliceToMap(corporations)
	corpIds := tools.CorpIDsFromCorpMap(corpMap)

	whereIds := []interface{}{}
	for _, id := range corpIds {
		whereIds = append(whereIds, id)
	}
	spew.Dump(whereIds)
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

		for i, history := range histories {
			if i != historiesLen-1 {
				if history.LeaveDate.Valid {
					continue
				}
				j := i + 1
				history.LeaveDate.SetValid(histories[j].StartDate)
			}
			spew.Dump(history)
			var boilHistory boiler.CorporationAllianceHistory
			err = copier.Copy(&boilHistory, history)
			if err != nil {
				a.Logger.WithFields(logrus.Fields{
					"id":        history.ID,
					"record_id": history.RecordID,
				}).WithError(err).Error("Failed to copy history to boilHistory")
				continue
			}
			fmt.Println("--------------------------------")
			spew.Dump(boilHistory)
			// boil.DebugMode = true

			// err = boilHistory.Update(context.Background(), a.DB, boil.Infer())
			// if err != nil {
			// 	a.Logger.WithFields(logrus.Fields{
			// 		"id":        history.ID,
			// 		"record_id": history.RecordID,
			// 	}).WithError(err).Error("Failed to update history in database")
			// }
			// boil.DebugMode = false
			DebugSleep("End of Main History Loop", 0)

		}
		DebugSleep("End of Main History Loop", 0)
		a.Logger.WithField("id", histories[0].ID).Debug("Done")
		a.Logger.Debug("---------------")
		// a.UpdateCorporation(histories[0].ID)
	}
	DebugSleep("End of Main Histories Loop", 0)
}

func DebugSleep(msg string, dur int) {
	if dur == 0 {
		dur = 5
	}
	fmt.Println(msg)
	time.Sleep(time.Second * time.Duration(dur))
}

func (a *Auditor) UpdateCorporation(id uint64) {

	corporation, err := boiler.FindCorporation(context.Background(), a.DB, id)
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
