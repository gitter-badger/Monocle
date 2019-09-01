package mysql

import (
	"fmt"

	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
	"github.com/pkg/errors"
)

func (db *DB) SelectCorporations(page, perPage uint) ([]monocle.Corporation, error) {

	corporations := make([]monocle.Corporation, 0)

	s := sb.NewSelectBuilder()
	q := s.Select(
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
	).From(
		"monocle.corporations",
	)

	offset := (page * perPage) - perPage

	q.Limit(int(perPage)).Offset(int(offset))

	query, args := q.Build()

	err := db.Select(&corporations, query, args...)
	return corporations, err
}

func (db *DB) SelectCorporationByCorporationID(id uint64) (monocle.Corporation, error) {

	var corporation monocle.Corporation

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
		"tax_rate",
		"war_eligible",
		"ignored",
		"closed",
		"etag",
		"expires",
		"created_at",
		"updated_at",
	).From(
		"monocle.corporations",
	).Where(
		s.E("id", id),
	).Limit(1)

	query, args := s.Build()

	err := db.Get(&corporation, query, args...)
	return corporation, err
}

func (db *DB) SelectIndependentCorps(page, perPage int) ([]monocle.Corporation, error) {
	var corporations []monocle.Corporation

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
		"tax_rate",
		"war_eligible",
		"ignored",
		"closed",
		"etag",
		"expires",
		"created_at",
		"updated_at",
	).From(
		"monocle.corporations",
	)

	offset := (page * perPage) - perPage
	s.Where(
		s.E("closed", 0),
		s.E("ignored", 0),
		s.IsNull("alliance_id"),
	).Limit(perPage).Offset(offset)

	query, args := s.Build()

	err := db.Select(&corporations, query, args...)
	return corporations, err
}

func (db *DB) SelectMissingCorporationIdsFromList(pid int, ids []int) ([]int, error) {
	var results []int
	var table = "temp_ids"

	d := sb.NewDeleteBuilder()
	d.DeleteFrom(table).Where(
		d.E("pid", pid),
	)

	query, args := d.Build()

	_, err := db.Exec(query, args...)
	if err != nil {
		return results, err
	}

	i := sb.NewInsertBuilder()
	i.InsertInto(table).Cols(
		"pid", "id",
	)
	for _, v := range ids {
		i.Values(pid, v)
	}

	query, args = i.Build()

	_, err = db.Exec(query, args...)
	if err != nil {
		err = errors.Wrapf(err, "Unable to insertIds into temporary %s table", table)
		return results, err
	}

	s := sb.NewSelectBuilder()
	s.Select("tmp.id")
	s.From(
		fmt.Sprintf("%s tmp", table),
	)
	s.JoinWithOption(sb.LeftJoin, "corporations corps", "tmp.id = corps.id")
	s.Where(
		s.IsNull("corps.id"),
	)

	query, _ = s.Build()
	err = db.Select(&results, query)
	if err != nil {
		err = errors.Wrapf(err, "Unable perform select operation temporary %s table", table)
		return results, err
	}

	query, args = d.Build()

	_, err = db.Exec(query, args...)

	return results, err
}

func (db *DB) SelectCountOfExpiredCorporationEtags() (uint, error) {
	var count uint

	s := sb.NewSelectBuilder()
	s.Select(
		s.As("COUNT(*)", "count"),
	).From(
		"monocle.corporations",
	)

	s.Where(
		s.LessThan("expires", sb.Raw("NOW()")),
		s.E("ignored", 0),
	)

	query, args := s.Build()
	err := db.Get(&count, query, args...)
	return count, err
}

func (db *DB) SelectCountOfCorporationEtags() (uint, error) {
	var count uint

	s := sb.NewSelectBuilder()
	s.Select(
		s.As("COUNT(*)", "count"),
	).From(
		"monocle.corporations",
	)

	s.Where(
		s.E("ignored", 0),
	)

	query, args := s.Build()
	err := db.Get(&count, query, args...)
	return count, err
}

func (db *DB) SelectExpiredCorporationEtags(page, perPage int) ([]monocle.Corporation, error) {

	var corporations []monocle.Corporation

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
		"tax_rate",
		"war_eligible",
		"ignored",
		"closed",
		"etag",
		"expires",
		"created_at",
		"updated_at",
	).From(
		"monocle.corporations",
	)

	offset := (page * perPage) - perPage
	s.Where(
		s.LessThan("expires", sb.Raw("NOW()")),
		s.E("ignored", 0),
	).OrderBy("expires").Asc().Limit(perPage).Offset(offset)

	query, args := s.Build()

	err := db.Select(&corporations, query, args...)
	return corporations, err
}

func (db *DB) InsertCorporation(corporation monocle.Corporation) (monocle.Corporation, error) {

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
	if err != nil {
		return corporation, err
	}

	return db.SelectCorporationByCorporationID(corporation.ID)
}

func (db *DB) UpdateCorporationByID(corporation monocle.Corporation) (monocle.Corporation, error) {
	u := sb.NewUpdateBuilder()
	u.Update("monocle.corporations").Set(
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
	if err != nil {
		return corporation, err
	}

	return db.SelectCorporationByCorporationID(corporation.ID)
}

func (db *DB) DeleteCorporationByID(id uint) error {
	d := sb.NewDeleteBuilder()
	d.DeleteFrom("monocle.corporations").Where(d.E("id", id))

	query, args := d.Build()

	_, err := db.Exec(query, args...)
	return err
}

func (db *DB) SelectCorporationAllianceHistoryByID(id uint64) ([]monocle.CorporationAllianceHistory, error) {

	history := make([]monocle.CorporationAllianceHistory, 0)

	sb := sb.NewSelectBuilder()
	q := sb.Select(
		"id",
		"record_id",
		"alliance_id",
		"start_date",
		"created_at",
		"updated_at",
	).
		From("monocle.corporation_alliance_history").
		Where(
			sb.E("id", id),
		)

	query, args := q.Build()

	err := db.Select(&history, query, args...)
	return history, err
}

func (db *DB) InsertCorporationAllianceHistory(id uint64, history []monocle.CorporationAllianceHistory) ([]monocle.CorporationAllianceHistory, error) {

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
	if err != nil {
		return history, err
	}

	return db.SelectCorporationAllianceHistoryByID(id)

}
