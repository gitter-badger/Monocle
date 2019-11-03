package mysql

import (
	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) InsertCorporation(corporation *monocle.Corporation) error {

	i := sb.NewInsertBuilder()
	i.ReplaceInto("monocle.corporations").Cols(
		"id",
		"name",
		"ticker",
		"member_count",
		"ceo_id",
		"alliance_id",
		"date_founded",
		"creator_id",
		"home_station_id",
		"tax_rate",
		"war_eligible",
		"ignored",
		"closed",
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
		corporation.TaxRate,
		corporation.WarEligible,
		corporation.Ignored,
		corporation.Closed,
		corporation.Etag,
		corporation.Expires,
		sb.Raw("NOW()"),
		sb.Raw("NOW()"),
	)

	query, args := i.Build()

	_, err := db.Exec(query, args...)

	return err
}

func (db *DB) UpdateCorporationByID(corporation *monocle.Corporation) error {
	u := sb.NewUpdateBuilder()
	u.Update("monocle.corporations").Set(
		u.E("name", corporation.Name),
		u.E("ticker", corporation.Ticker),
		u.E("date_founded", corporation.DateFounded),
		u.E("creator_id", corporation.CreatorID),
		u.E("member_count", corporation.MemberCount),
		u.E("ceo_id", corporation.CeoID),
		u.E("alliance_id", corporation.AllianceID),
		u.E("home_station_id", corporation.HomeStationID),
		u.E("tax_rate", corporation.TaxRate),
		u.E("war_eligible", corporation.WarEligible),
		u.E("ignored", corporation.Ignored),
		u.E("closed", corporation.Closed),
		u.E("expires", corporation.Expires),
		u.E("etag", corporation.Etag),
		u.E("updated_at", sb.Raw("NOW()")),
	).Where(
		u.E("id", corporation.ID),
	)

	query, args := u.Build()

	_, err := db.Exec(query, args...)

	return err
}

func (db *DB) InsertCorporationAllianceHistory(id uint64, history []*monocle.CorporationAllianceHistory) error {

	ib := sb.NewInsertBuilder()
	q := ib.InsertIgnoreInto("monocle.corporation_alliance_history").Cols(
		"id",
		"record_id",
		"alliance_id",
		"start_date",
		"created_at",
		"updated_at",
	)
	for _, v := range history {
		q.Values(
			id,
			v.RecordID,
			v.AllianceID,
			v.StartDate,
			sb.Raw("NOW()"),
			sb.Raw("NOW()"),
		)
	}

	query, args := ib.Build()

	_, err := db.Exec(query, args...)

	return err

}
