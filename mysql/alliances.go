package mysql

import (
	"fmt"

	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
	"github.com/pkg/errors"
)

func (db *DB) SelectAlliances(page, limit int) ([]monocle.Alliance, error) {
	var alliances []monocle.Alliance

	s := sb.NewSelectBuilder()
	s.Select(
		"id",
		"name",
		"ticker",
		"creator_corporation_id",
		"creator_id",
		"date_founded",
		"executor_corporation_id",
		"ignored",
		"closed",
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).From(
		"monocle.alliances",
	)

	offset := (page * limit) - limit
	s.Where(
		s.E("ignored", 0),
		s.E("closed", 0),
	).Limit(limit).Offset(offset)

	query, args := s.Build()

	err := db.Select(&alliances, query, args...)
	return alliances, err
}

func (db *DB) SelectAllianceByAllianceID(id uint) (monocle.Alliance, error) {

	var alliance monocle.Alliance
	s := sb.NewSelectBuilder()
	s.Select(
		"id",
		"name",
		"ticker",
		"creator_corporation_id",
		"creator_id",
		"date_founded",
		"executor_corporation_id",
		"ignored",
		"closed",
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).From(
		"monocle.alliances",
	).Where(
		s.E("id", id),
	).Limit(1)

	query, args := s.Build()

	err := db.Get(&alliance, query, args...)
	return alliance, err

}

func (db *DB) SelectMissingAllianceIdsFromList(ids []int) ([]monocle.AllianceIDs, error) {
	var results []monocle.AllianceIDs
	var table = "allianceids"

	query := fmt.Sprintf("TRUNCATE %s", table)
	_, err := db.Exec(query)
	if err != nil {
		return results, err
	}

	i := sb.NewInsertBuilder()
	i.InsertInto(table).Cols(
		"id",
	)
	for _, v := range ids {
		i.Values(v)
	}

	query, args := i.Build()

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
	s.JoinWithOption(sb.LeftJoin, "alliances alli", "tmp.id = alli.id")
	s.Where(
		s.IsNull("alli.id"),
	)

	query, _ = s.Build()
	err = db.Select(&results, query)
	if err != nil {
		err = errors.Wrapf(err, "Unable perform select operation temporary %s table", table)
		return results, err
	}

	query = fmt.Sprintf("TRUNCATE %s", table)
	_, err = db.Exec(query)
	return results, err
}

func (db *DB) SelectCountOfExpiredAllianceEtags() (monocle.Counter, error) {
	var counter monocle.Counter

	s := sb.NewSelectBuilder()
	s.Select(
		s.As("COUNT(*)", "count"),
	).From(
		"monocle.alliances",
	)

	s.Where(
		s.LessThan("expires", sb.Raw("NOW()")),
		s.E("ignored", 0),
	)

	query, args := s.Build()
	err := db.Get(&counter, query, args...)
	return counter, err

}

func (db *DB) SelectCountOfAllianceEtags() (monocle.Counter, error) {
	var counter monocle.Counter

	s := sb.NewSelectBuilder()
	s.Select(
		s.As("COUNT(*)", "count"),
	).From(
		"monocle.alliances",
	)

	s.Where(
		s.E("ignored", 0),
	)

	query, args := s.Build()
	err := db.Get(&counter, query, args...)
	return counter, err

}

func (db *DB) SelectExpiredAllianceEtags(page, perPage int) ([]monocle.Alliance, error) {
	var alliances []monocle.Alliance

	s := sb.NewSelectBuilder()
	s.Select(
		"id",
		"name",
		"ticker",
		"creator_corporation_id",
		"creator_id",
		"date_founded",
		"executor_corporation_id",
		"ignored",
		"closed",
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).From(
		"monocle.alliances",
	)

	offset := (page * perPage) - perPage
	s.Where(
		s.LessThan("expires", sb.Raw("NOW()")),
		s.E("ignored", 0),
	).OrderBy("expires").Asc().Limit(perPage).Offset(offset)

	query, args := s.Build()
	err := db.Select(&alliances, query, args...)
	return alliances, err
}

func (db *DB) InsertAlliance(alliance monocle.Alliance) (monocle.Alliance, error) {

	i := sb.NewInsertBuilder()
	i.ReplaceInto("monocle.alliances").Cols(
		"id",
		"name",
		"ticker",
		"creator_corporation_id",
		"creator_id",
		"date_founded",
		"executor_corporation_id",
		"ignored",
		"closed",
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).Values(
		alliance.ID,
		alliance.Name,
		alliance.Ticker,
		alliance.CreatorCorporationID,
		alliance.CreatorID,
		alliance.DateFounded,
		alliance.ExecutorCorporationID,
		alliance.Ignored,
		alliance.Closed,
		alliance.Expires,
		alliance.Etag,
		sb.Raw("NOW()"),
		sb.Raw("NOW()"),
	)

	query, args := i.Build()

	_, err := db.Exec(query, args...)
	if err != nil {
		return alliance, err
	}

	return db.SelectAllianceByAllianceID(alliance.ID)

}

func (db *DB) UpdateAllianceByID(alliance monocle.Alliance) (monocle.Alliance, error) {

	u := sb.NewUpdateBuilder()
	u.Update("monocle.alliances").Set(
		u.E("executor_corporation_id", alliance.ExecutorCorporationID),
		u.E("ignored", alliance.Ignored),
		u.E("closed", alliance.Closed),
		u.E("expires", alliance.Expires),
		u.E("etag", alliance.Etag),
	).Where(
		u.E("id", alliance.ID),
	)

	query, args := u.Build()

	_, err := db.Exec(query, args...)
	if err != nil {
		return alliance, err
	}

	return db.SelectAllianceByAllianceID(alliance.ID)

}

func (db *DB) DeleteAllianceByID(id uint) error {
	d := sb.NewDeleteBuilder()
	d.DeleteFrom("monocle.alliances").Where(d.E("id", id))

	query, args := d.Build()

	_, err := db.Exec(query, args...)
	return err
}
