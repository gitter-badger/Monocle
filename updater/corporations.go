package updater

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
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

			// if len(corporations) < threshold {
			// 	if x == 1 {
			// 		u.Logger.Infof("Minimum threshold of %d for job not met. Sleeping for %d seconds", threshold, sleep)
			// 		time.Sleep(time.Second * time.Duration(sleep))
			// 		continue supervisor
			// 	}
			// 	u.Logger.Info("Breaking from Worker Loop")
			// 	break
			// }

			wg.Add(1)
			go func(corporations []monocle.Corporation) {

				for _, corporation := range corporations {
					if !corporation.IsExpired() {
						continue
					}
					u.updateCorporation(corporation)
				}
				wg.Done()
				return
			}(corporations)
		}
		u.Logger.Info("Waiting")
		wg.Wait()
		u.Logger.Info("Done")
		os.Exit(1)
	}

}

func (u *Updater) updateCorporation(corporation monocle.Corporation) {
	u.Logger.DebugF("Updating Corporation %d", corporation.ID)
	var response esi.Response
	if u.ESI.Remain < 20 {
		msg := fmt.Sprintf("Error Counter is Low, sleeping for %d seconds", u.ESI.Reset)
		u.Logger.Errorf("\t%s", msg)
		// msg = fmt.Sprintf("<@!277968564827324416> %s", msg)
		u.DGO.ChannelMessageSend("394991263344230411", msg)
		time.Sleep(time.Second * time.Duration(u.ESI.Reset))
	}
	attempts := 0
	for {
		if attempts >= 3 {
			u.Logger.Errorf("All Attempts exhuasted for Corporation %d", corporation.ID)
			return
		}
		response, err = u.ESI.GetCorporationsCorporationID(corporation)
		if err != nil {
			u.Logger.Errorf("Error completing request to ESI for Corporation %d information: %s", corporation.ID, err)
			return
		}

		if response.Code < 400 {
			break
		}

		attempts++
		u.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s, attempting %d request again in 1 second", response.Code, response.Path, attempts)
		time.Sleep(1 * time.Second)
	}

	corporation = response.Data.(monocle.Corporation)

	_, err = u.DB.UpdateCorporationByID(corporation)
	if err != nil {
		u.Logger.Errorf("Error Encountered attempting to update corporation %d in database: %s", corporation.ID, err)
		return
	}

	return
}
