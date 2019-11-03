package mysql

import (
	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) InsertAlliance(alliance *monocle.Alliance) error {

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
		return err
	}

	return err

}

func (db *DB) UpdateAllianceByID(alliance *monocle.Alliance) error {

	u := sb.NewUpdateBuilder()
	u.Update("monocle.alliances").Set(
		u.E("executor_corporation_id", alliance.ExecutorCorporationID),
		u.E("member_count", alliance.MemberCount),
		u.E("ignored", alliance.Ignored),
		u.E("closed", alliance.Closed),
		u.E("expires", alliance.Expires),
		u.E("etag", alliance.Etag),
		u.E("updated_at", sb.Raw("NOW()")),
	).Where(
		u.E("id", alliance.ID),
	)

	query, args := u.Build()

	_, err = db.Exec(query, args...)
	if err != nil {
		return err
	}

	return err

}

func (db *DB) DeleteAllianceByID(id uint) error {
	d := sb.NewDeleteBuilder()
	d.DeleteFrom("monocle.alliances").Where(d.E("id", id))

	query, args := d.Build()

	_, err := db.Exec(query, args...)
	return err
}
