package updater

import (
	"database/sql"
	"time"

	"github.com/ddouglas/monocle"
)

func (u *Updater) evaluateAlliances(sleep, threshold int) error {
	var errorCount int
supervisor:
	for {
		if u.ESI.Remain < 10 {
			u.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", u.ESI.Reset)
			time.Sleep(time.Second * time.Duration(u.ESI.Reset))
		}

		for x := 1; x <= workers; x++ {
			alliances, err := u.DB.SelectExpiredAllianceEtags(x, records)
			if err != nil {
				errorCount++
				if err != sql.ErrNoRows {
					u.Logger.Fatalf("Unable to query for alliances: %s", err)
				}
				time.Sleep(time.Second * time.Duration(sleep))
				continue supervisor
			}

			if len(alliances) < threshold {
				if x == 1 {
					u.Logger.Infof("Minimum threshold of %d for job not met. Sleeping for %d seconds", threshold, sleep)
					time.Sleep(time.Second * time.Duration(sleep))
					continue supervisor
				}
				u.Logger.Info("Breaking from Worker Loop")
				break
			}

			u.Logger.Infof("Successfully Queried %d Alliances", len(alliances))
			wg.Add(1)
			go func(alliances []monocle.Alliance) {
				for _, alliance := range alliances {
					if !alliance.IsExpired() {
						continue
					}
					u.updateAlliance(alliance)
				}
				wg.Done()
				return
			}(alliances)
		}
		u.Logger.Info("Waiting")
		wg.Wait()
		u.Logger.Info("Done")
	}
}

func (u *Updater) updateAlliance(alliance monocle.Alliance) {
	u.Logger.DebugF("Updating Alliance %s", alliance.ID)

	response, err := u.ESI.GetAlliancesAllianceID(alliance)
	if err != nil {
		u.Logger.Errorf("Error completing request to ESI for Alliance %d information: %s", alliance.ID, err)
		return
	}

	alliance = response.Data.(monocle.Alliance)

	_, err = u.DB.UpdateAllianceByID(alliance)
	if err != nil {
		u.Logger.Errorf("Error Encountered attempting to update alliance in database: %s", err)
		return
	}

	return
}
