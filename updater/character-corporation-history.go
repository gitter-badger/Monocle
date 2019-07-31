package updater

import (
	"database/sql"
	"time"
)

func (u *Updater) evaluateCharacterCorporationHistory() error {
	for {
		u.Logger.DebugF("Current Error Count: %d Remain: %d", u.count, u.reset)
		if u.count < 20 {
			u.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", u.reset)
			time.Sleep(time.Second * time.Duration(u.reset))
		}
		for x := 1; x <= workers; x++ {
			characters, err := u.DB.SelectExpiredCharacterEtags(x, records)
			if err != nil {
				if err != sql.ErrNoRows {
					u.Logger.Fatalf("Unable to query for characters: %s", err)
				}
				continue
			}

			if len(characters) < threshold {
				u.Logger.Infof("Minimum threshold of %d for job not met. Sleeping for %d seconds", threshold, sleep)
				time.Sleep(time.Second * time.Duration(sleep))
				continue
			}

			u.Logger.Infof("Successfully Queried %d Characters", len(characters))
			wg.Add(1)
			go u.updateCharacters(characters)
		}
		u.Logger.Info("Waiting")
		wg.Wait()
		u.Logger.Info("Done")
	}

}
