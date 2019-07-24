package mysql

import (
	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) SelectEtagByIDAndResource(id uint, resource string) (monocle.EtagResource, error) {
	var resource monocle.EtagResource

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
		s.E("resource", resource),
	).Limit(1)

	query, args := s.Build()

	err := db.Select(&resource, query, args...)
	return resource, err
}

// SelectCountOfExpiredIAREtags select the count of expired etags by id and resource
// func (db *DB) SelectCountOfExpiredIAREtags() ([]monocle.EtagRourse, error) {

// }
