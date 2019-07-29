package mysql

import (
	"fmt"

	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) SelectCharacters(page, perPage int, where map[string]interface{}) ([]monocle.Character, error) {

	var characters []monocle.Character

	s := sb.NewSelectBuilder()
	q := s.Select(
		"id",
		"name",
		"birthday",
		"gender",
		"security_status",
		"alliance_id",
		"corporation_id",
		"faction_id",
		"ancestry_id",
		"bloodline_id",
		"race_id",
		"ignored",
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).From(
		"monocle.characters",
	)

	for col, val := range where {
		q.Where(q.E(col, val))
	}

	offset := (page * perPage) - perPage

	q.Limit(perPage).Offset(offset)

	query, args := q.Build()

	err := db.Select(&characters, query, args...)
	return characters, err
}

func (db *DB) SelectCharactersFromRange(start int, end int) ([]monocle.Character, error) {
	var characters []monocle.Character

	s := sb.NewSelectBuilder()
	q := s.Select(
		"id",
		"name",
		"birthday",
		"gender",
		"security_status",
		"alliance_id",
		"corporation_id",
		"faction_id",
		"ancestry_id",
		"bloodline_id",
		"race_id",
		"ignored",
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).From(
		"monocle.characters",
	).Where(
		s.Between("id", start, end),
		s.E("ignored", false),
	)

	query, args := q.Build()

	fmt.Println(query)

	err := db.Select(&characters, query, args...)
	return characters, err
}

func (db *DB) SelectCharacterByCharacterID(id uint64) (monocle.Character, error) {

	var character monocle.Character

	s := sb.NewSelectBuilder()
	s.Select(
		"id",
		"name",
		"birthday",
		"gender",
		"security_status",
		"alliance_id",
		"corporation_id",
		"faction_id",
		"ancestry_id",
		"bloodline_id",
		"race_id",
		"ignored",
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).From(
		"monocle.characters",
	).Where(
		s.E("id", id),
	).Limit(1)

	query, args := s.Build()

	err := db.Get(&character, query, args...)
	return character, err
}

func (db *DB) SelectExpiredCharacterEtags(page, perPage int) ([]monocle.Character, error) {
	var characters []monocle.Character

	s := sb.NewSelectBuilder()
	s.Select(
		"id",
		"name",
		"birthday",
		"gender",
		"security_status",
		"alliance_id",
		"corporation_id",
		"faction_id",
		"ancestry_id",
		"bloodline_id",
		"race_id",
		"ignored",
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).From(
		"monocle.characters",
	)

	offset := (page * perPage) - perPage
	s.Where(
		s.LessThan("expires", sb.Raw("NOW()")),
		s.E("ignored", 0),
	).OrderBy("expires").Asc().Limit(perPage).Offset(offset)

	query, args := s.Build()

	err := db.Select(&characters, query, args...)
	return characters, err
}

func (db *DB) SelectCountOfExpiredCharacterEtags() (monocle.Counter, error) {
	var counter monocle.Counter

	s := sb.NewSelectBuilder()
	s.Select(
		s.As("COUNT(*)", "count"),
	).From(
		"monocle.characters",
	)

	s.Where(
		s.LessThan("expires", sb.Raw("NOW()")),
		s.E("ignored", 0),
	)

	query, args := s.Build()
	err := db.Get(&counter, query, args...)
	return counter, err
}

func (db *DB) SelectCountOfCharacterEtags() (monocle.Counter, error) {
	var counter monocle.Counter

	s := sb.NewSelectBuilder()
	s.Select(
		s.As("COUNT(*)", "count"),
	).From(
		"monocle.characters",
	)

	s.Where(
		s.E("ignored", 0),
	)

	query, args := s.Build()
	err := db.Get(&counter, query, args...)
	return counter, err
}

func (db *DB) InsertCharacter(character monocle.Character) (monocle.Character, error) {

	i := sb.NewInsertBuilder()
	i.InsertIgnoreInto("monocle.characters").Cols(
		"id",
		"name",
		"birthday",
		"gender",
		"security_status",
		"alliance_id",
		"corporation_id",
		"faction_id",
		"ancestry_id",
		"bloodline_id",
		"race_id",
		"ignored",
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).Values(
		character.ID,
		character.Name,
		character.Birthday,
		character.Gender,
		character.SecurityStatus,
		character.AllianceID,
		character.CorporationID,
		character.FactionID,
		character.AncestryID,
		character.BloodlineID,
		character.RaceID,
		character.Ignored,
		character.Expires,
		character.Etag,
		sb.Raw("NOW()"),
		sb.Raw("NOW()"),
	)

	query, args := i.Build()
	_, err := db.Exec(query, args...)
	if err != nil {
		return character, err
	}

	return db.SelectCharacterByCharacterID(character.ID)

}

func (db *DB) UpdateCharacterByID(character monocle.Character) (monocle.Character, error) {

	u := sb.NewUpdateBuilder()
	u.Update("monocle.characters").Set(
		u.E("security_status", character.SecurityStatus),
		u.E("alliance_id", character.AllianceID),
		u.E("corporation_id", character.CorporationID),
		u.E("faction_id", character.FactionID),
		u.E("ignored", character.Ignored),
		u.E("expires", character.Expires),
		u.E("etag", character.Etag),
		u.E("updated_at", sb.Raw("NOW()")),
	).Where(
		u.E("id", character.ID),
	)

	query, args := u.Build()

	_, err := db.Exec(query, args...)
	if err != nil {
		return character, err
	}

	return db.SelectCharacterByCharacterID(character.ID)
}

func (db *DB) DeleteCharacterByID(id uint64) error {

	d := sb.NewDeleteBuilder()
	d.DeleteFrom("monocle.characters").Where(d.E("id", id))

	query, args := d.Build()

	_, err := db.Exec(query, args...)
	return err

}
