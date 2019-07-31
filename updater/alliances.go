package updater

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

func (u *Updater) evaluateAlliances(sleep, threshold int) error {
	var errorCount int
supervisor:
	for {
		if u.count < 10 {
			u.Logger.Errorf("Error Counter is Low, sleeping for %d seconds", u.reset)
			time.Sleep(time.Second * time.Duration(u.reset))
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
			go u.updateAlliances(alliances)
		}
		u.Logger.Info("Waiting")
		wg.Wait()
		u.Logger.Info("Done")
	}
}

func (u *Updater) updateAlliances(alliances []monocle.Alliance) {
	defer wg.Done()
	for _, alliance := range alliances {
		if !alliance.IsExpired() {
			continue
		}
		u.updateAlliance(alliance)
	}
	return
}

func (u *Updater) updateAlliance(alliance monocle.Alliance) {
	u.Logger.DebugF("Updating Alliance %s", alliance.ID)

	response, err := u.ESI.GetAlliancesAllianceID(alliance.ID, alliance.Etag)
	if err != nil {
		u.Logger.Errorf("Error completing request to ESI for Alliance %d information: %s", alliance.ID, err)
		return
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &alliance)
		if err != nil {
			u.Logger.Errorf("unable to unmarshel response body for %d: %s", alliance.ID, err)
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
		alliance.Etag = etag

		alliance.Expires = expires
		break
	case 304:
		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to parse expires header for url %s: %s", response.Path, err)
		}

		alliance.Expires = expires

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			u.Logger.Errorf("Error Encountered attempting to retrieve etag header for url %s: %s", response.Path, err)
		}
		alliance.Etag = etag
		break
	default:
		u.Logger.ErrorF("Bad Response Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	_, err = u.DB.UpdateAllianceByID(alliance)
	if err != nil {
		u.Logger.Errorf("Error Encountered attempting to update alliance in database: %s", err)
		return
	}

	return
}
