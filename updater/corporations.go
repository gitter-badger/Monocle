package updater

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

func (u *Updater) evaluateCorporations(sleep, threshold int) error {

	var errorCount int
supervisor:
	for {
		u.Logger.DebugF("Current Error Count: %d Remain: %d", u.count, u.reset)
		if u.count < 10 {
			u.Logger.Error("Error Counter is Low, sleeping for 30 seconds")
			time.Sleep(time.Second * 30)
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

	response, err := u.ESI.GetCorporationsCorporationID(corporation.ID, corporation.Etag)
	if err != nil {
		u.Logger.Errorf("Error completing request to ESI for Corporation %d information: %s", corporation.ID, err)
		return
	}

	mx.Lock()
	defer mx.Unlock()
	u.reset = esi.RetrieveErrorResetFromResponse(response)
	u.count = esi.RetrieveErrorCountFromResponse(response)

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &corporation)
		if err != nil {
			u.Logger.Errorf("unable to unmarshel response body for %d: %s", corporation.ID, err)
			return
		}
		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to parse expires header for url %s: %s", response.Path, err)
		}

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to retrieve etag header for url %s: %s", response.Path, err)
		}
		corporation.Etag = etag

		corporation.Expires = expires
		break
	case 304:
		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to parse expires header for url %s: %s", response.Path, err)
		}

		corporation.Expires = expires

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to retrieve etag header for url %s: %s", response.Path, err)
		}
		corporation.Etag = etag
		break
	default:
		u.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	_, err = u.DB.UpdateCorporationByID(corporation)
	if err != nil {
		u.Logger.Errorf("Error Encountered attempting to update corporation in database: %s", err)
		return
	}

	return
}
