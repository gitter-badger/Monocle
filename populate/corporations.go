package populate

import (
	"github.com/ddouglas/monocle"
)

func (p *Populator) processNewCorporation(id uint) {

	var corporation monocle.Corporation
	corporation.ID = id
	response, err := p.ESI.GetCorporationsCorporationID(corporation)
	if err != nil {
		p.Logger.Errorf("Error completing request to ESI for Character information: %s", err)
		return
	}

	corporation = response.Data.(monocle.Corporation)

	p.Logger.Debugf("\tCorporation: %d:%s", corporation.ID, corporation.Name)

	_, err = p.DB.InsertCorporation(corporation)
	if err != nil {
		p.Logger.Errorf("Error Encountered attempting to insert new corporation into database: %s", err)
		return
	}
}
