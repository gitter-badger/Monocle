package mysql

import (
	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) SelectEtagByIDAndResource(id uint64, resource string) (monocle.EtagResource, error) {
	var etag monocle.EtagResource

	s := sb.NewSelectBuilder()
	s.Select(
		"id",
		"resource",
		"etag",
		"expires",
		"created_at",
		"updated_at",
	).From(
		"monocle.etags",
	).Where(
		s.E("id", id),
		s.E("resource", resource),
	).Limit(1)

	query, args := s.Build()

	err := db.Get(&etag, query, args...)
	return etag, err
}

// // SelectCountOfExpiredIAREtags select the count of expired etags by id and resource
// func (db *DB) SelectCountOfExpiredIAREtags() (int, error) {
// 	var count int

// 	s := sb.NewSelectBuilder()
// 	s.Select(
// 		s.As("COUNT(*)", "count"),
// 	).From(
// 		"monocle.etags",
// 	).Where(
// 		s.LessThan("expires", sb.Raw("NOW()")),
// 	).Limit(1)

// 	query, args := s.Build()

// 	err := db.Get(&count, query, args...)
// 	return count, err
// }

func (db *DB) InsertEtag(etag monocle.EtagResource) (monocle.EtagResource, error) {

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
	if err != nil {
		return etag, err
	}

	return db.SelectEtagByIDAndResource(etag.ID, etag.Resource)

}

func (db *DB) UpdateEtagByIDAndResource(etag monocle.EtagResource) (monocle.EtagResource, error) {

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
	if err != nil {
		return etag, err
	}

	return db.SelectEtagByIDAndResource(etag.ID, etag.Resource)

}

func (db *DB) DeleteEtagByIDAndResource(id int, resource string) error {

	d := sb.NewDeleteBuilder()
	d.DeleteFrom("monocle.etags").Where(
		d.E("id", id),
		d.E("resource", resource),
	)

	query, args := d.Build()

	_, err := db.Exec(query, args...)
	return err
}
