package mysql

import (
	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) SelectValueByKey(key string) (monocle.KeyValue, error) {

	var value monocle.KeyValue

	s := sb.NewSelectBuilder()
	s.Select(
		"k",
		"v",
		"created_at",
		"updated_at",
	).From("monocle.kv").Where(
		s.E("k", key),
	)

	query, args := s.Build()

	err := db.Get(&value, query, args...)
	return value, err
}

func (db *DB) UpdateValueByKey(kv monocle.KeyValue) (monocle.KeyValue, error) {

	u := sb.NewUpdateBuilder()
	u.Update("monocle.kv").Set(
		u.E("v", string(kv.Value)),
		u.E("updated_at", sb.Raw("NOW()")),
	).Where(
		u.E("k", kv.Key),
	)

	query, args := u.Build()

	_, err := db.Exec(query, args...)
	if err != nil {
		return kv, err
	}

	return db.SelectValueByKey(kv.Key)

}
