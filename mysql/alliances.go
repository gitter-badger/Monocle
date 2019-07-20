package mysql

import (
	"github.com/ddouglas/eveindex"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) SelectAllianceByAllianceID(id uint) (eveindex.Alliance, error) {

	var alliance eveindex.Alliance
	s := sb.NewSelectBuilder()
	s.Select(
		"id",
		"name",
		"ticker",
		"creator_corporation_id",
		"creator_id",
		"date_founded",
		"executor_corporation_id",
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).From(
		"eveindex.alliances",
	).Where(
		s.E("id", id),
	).Limit(1)

	query, args := s.Build()

	err := db.Get(&alliance, query, args...)
	return alliance, err

}

func (db *DB) InsertAlliance(alliance eveindex.Alliance) (eveindex.Alliance, error) {

	i := sb.NewInsertBuilder()
	i.ReplaceInto("eveindex.alliances").Cols(
		"id",
		"name",
		"ticker",
		"creator_corporation_id",
		"creator_id",
		"date_founded",
		"executor_corporation_id",
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

func (db *DB) UpdateAllianceByID(alliance eveindex.Alliance) (eveindex.Alliance, error) {

	u := sb.NewUpdateBuilder()
	u.Update("eveindex.alliances").Set(
		u.E("executor_corporation_id", alliance.ExecutorCorporationID),
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
	d.DeleteFrom("eveindex.alliances").Where(d.E("id", id))

	query, args := d.Build()

	_, err := db.Exec(query, args...)
	return err
}
