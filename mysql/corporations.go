package mysql

import (
	"github.com/ddouglas/eveindex"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) SelectCorporationByCorporationID(id uint) (eveindex.Corporation, error) {

	var corporation eveindex.Corporation

	s := sb.NewSelectBuilder()
	s.Select(
		"id",
		"name",
		"ticker",
		"member_count",
		"ceo_id",
		"alliance_id",
		"date_founded",
		"creator_id",
		"home_station_id",
		"url",
		"tax_rate",
		"war_eligible",
		"etag",
		"expires",
		"created_at",
		"updated_at",
	).From(
		"eveindex.corporations",
	).Where(
		s.E("id", id),
	).Limit(1)

	query, args := s.Build()

	err := db.Get(&corporation, query, args...)
	return corporation, err
}

func (db *DB) InsertCorporation(corporation eveindex.Corporation) (eveindex.Corporation, error) {

	i := sb.NewInsertBuilder()
	i.ReplaceInto("eveindex.corporations").Cols(
		"id",
		"name",
		"ticker",
		"member_count",
		"ceo_id",
		"alliance_id",
		"date_founded",
		"creator_id",
		"home_station_id",
		"url",
		"tax_rate",
		"war_eligible",
		"etag",
		"expires",
		"created_at",
		"updated_at",
	).Values(
		corporation.ID,
		corporation.Name,
		corporation.Ticker,
		corporation.MemberCount,
		corporation.CeoID,
		corporation.AllianceID,
		corporation.DateFounded,
		corporation.CreatorID,
		corporation.HomeStationID,
		corporation.URL,
		corporation.TaxRate,
		corporation.WarEligible,
		corporation.Etag,
		corporation.Expires,
		sb.Raw("NOW()"),
		sb.Raw("NOW()"),
	)

	query, args := i.Build()

	_, err := db.Exec(query, args...)
	if err != nil {
		return corporation, err
	}

	return db.SelectCorporationByCorporationID(corporation.ID)
}

func (db *DB) UpdateCorporationByID(corporation eveindex.Corporation) (eveindex.Corporation, error) {
	u := sb.NewUpdateBuilder()
	u.Update("eveindex.corporations").Set(
		u.E("member_count", corporation.MemberCount),
		u.E("ceo_id", corporation.CeoID),
		u.E("alliance_id", corporation.AllianceID),
		u.E("home_station_id", corporation.HomeStationID),
		u.E("url", corporation.URL),
		u.E("tax_rate", corporation.TaxRate),
		u.E("war_eligible", corporation.WarEligible),
		u.E("expires", corporation.Expires),
		u.E("etag", corporation.Etag),
	).Where(
		u.E("id", corporation.ID),
	)

	query, args := u.Build()

	_, err := db.Exec(query, args...)
	if err != nil {
		return corporation, err
	}

	return db.SelectCorporationByCorporationID(corporation.ID)
}

func (db *DB) DeleteCorporationByID(id uint) error {
	d := sb.NewDeleteBuilder()
	d.DeleteFrom("eveindex.corporations").Where(d.E("id", id))

	query, args := d.Build()

	_, err := db.Exec(query, args...)
	return err
}
