package mysql

import (
	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) InsertEtag(etag *monocle.EtagResource) error {

	i := sb.NewInsertBuilder()
	i.InsertInto("monocle.etags").Cols(
		"id",
		"resource",
		"etag",
		"expires",
		"created_at",
		"updated_at",
	).Values(
		etag.ID,
		etag.Resource,
		etag.Etag,
		etag.Expires,
		sb.Raw("NOW()"),
		sb.Raw("NOW()"),
	)

	query, args := i.Build()

	_, err := db.Exec(query, args...)

	return err

}

func (db *DB) UpdateEtagByIDAndResource(etag *monocle.EtagResource) error {

	u := sb.NewUpdateBuilder()
	u.Update("monocle.etags").Set(
		u.E("etag", etag.Etag),
		u.E("expires", etag.Expires),
		u.E("updated_at", sb.Raw("NOW()")),
	).Where(
		u.E("id", etag.ID),
		u.E("resource", etag.Resource),
	)

	query, args := u.Build()

	_, err := db.Exec(query, args...)

	return err

}
