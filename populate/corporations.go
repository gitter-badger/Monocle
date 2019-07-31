package populate

import (
	"encoding/json"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/esi"
)

func (p *Populator) processNewCorporation(id uint) {

	var corporation monocle.Corporation
	response, err := p.ESI.GetCorporationsCorporationID(id, "")
	if err != nil {
		p.Logger.Errorf("Error completing request to ESI for Character information: %s", err)
		return
	}

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &corporation)
		if err != nil {
			p.Logger.Errorf("unable to unmarshel response body: %s", err)
			return
		}

		corporation.ID = id

		expires, err := esi.RetrieveExpiresHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to parse expires header: %s", err)
		}
		corporation.Expires = expires

		etag, err := esi.RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			p.Logger.Errorf("Error Encountered attempting to retrieve etag header: %s", err)
		}
		corporation.Etag = etag

		break

	default:
		p.Logger.ErrorF("Bad Resposne Code %d received from ESI API for url %s:", response.Code, response.Path)
		return
	}

	p.Logger.Debugf("\tCorporation: %d:%s", corporation.ID, corporation.Name)

	_, err = p.DB.InsertCorporation(corporation)
	if err != nil {
		p.Logger.Errorf("Error Encountered attempting to insert new corporation into database: %s", err)
		return
	}
}
