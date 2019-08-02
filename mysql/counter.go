package mysql

import (
	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) InsertCounter(counter monocle.Counter) error {

	ib := sb.NewInsertBuilder()
	q := ib.InsertIgnoreInto("monocle.counter").Cols(
		"char_count",
		"corp_count",
		"alli_count",
		"created_at",
		"updated_at",
	).Values(
		counter.CharCount,
		counter.CorpCount,
		counter.AlliCount,
		sb.Raw("NOW()"),
		sb.Raw("NOW()"),
	)

	query, args := q.Build()

	_, err := db.Exec(query, args...)
	return err

}
