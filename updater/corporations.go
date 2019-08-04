package updater

import (
	"database/sql"
	"time"

	"github.com/ddouglas/monocle"
)

func (u *Updater) evaluateCorporations(sleep, threshold int) error {

	var errorCount int
supervisor:
	for {
		u.Logger.DebugF("Current Error Count: %d Remain: %d", u.ESI.Remain, u.ESI.Reset)
		if u.ESI.Remain < 10 {
			u.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", u.ESI.Reset)
			time.Sleep(time.Second * time.Duration(u.ESI.Reset))
		}
		for x := 1; x <= workers; x++ {
			corporations, err := u.DB.SelectExpiredCorporationEtags(x, records)
			if err != nil {
				errorCount++
				if err != sql.ErrNoRows {
					u.Logger.Errorf("Unable to query for corporations: %s", err)
				}
				continue supervisor
			}

			if len(corporations) < threshold {
				if x == 1 {
					u.Logger.Infof("Minimum threshold of %d for job not met. Sleeping for %d seconds", threshold, sleep)
					time.Sleep(time.Second * time.Duration(sleep))
					continue supervisor
				}
				u.Logger.Info("Breaking from Worker Loop")
				break
			}

			wg.Add(1)
			go u.updateCorporations(corporations)
		}
		u.Logger.Info("Waiting")
		wg.Wait()
		u.Logger.Info("Done")
	}

}

func (u *Updater) updateCorporations(corporations []monocle.Corporation) {
	defer wg.Done()
	for _, corporation := range corporations {
		if !corporation.IsExpired() {
			continue
		}
		u.updateCorporation(corporation)
	}
	return
}

func (u *Updater) updateCorporation(corporation monocle.Corporation) {
	u.Logger.DebugF("Updating Corporation %d", corporation.ID)

	response, err := u.ESI.GetCorporationsCorporationID(corporation)
	if err != nil {
		u.Logger.Errorf("Error completing request to ESI for Corporation %d information: %s", corporation.ID, err)
		return
	}

	corporation = response.Data.(monocle.Corporation)

	return
}
