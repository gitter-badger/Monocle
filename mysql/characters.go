package mysql

import (
	"fmt"

	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) SelectCharacters(page, perPage uint) ([]monocle.Character, error) {

	characters := make([]monocle.Character, 0)

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

	offset := (page * perPage) - perPage

	q.Limit(int(perPage)).Offset(int(offset))

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

	err := db.Select(&characters, query, args...)
	return characters, err
}

func (db *DB) SelectCharactersLikeName(name string, page, perPage int) ([]monocle.Character, error) {

	var characters []monocle.Character

	sb := sb.NewSelectBuilder()
	q := sb.Select(
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
		sb.Like("name", fmt.Sprintf("%%%s%%", name)),
		sb.G("id", 90848155),
	).Limit(perPage).Offset((page * perPage) - perPage)

	query, args := q.Build()

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
	)

	query, args := s.Build()

	err := db.Get(&character, query, args...)
	return character, err
}

func (db *DB) SelectExpiredCharacterEtags(limit int) ([]monocle.Character, error) {
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
	).
		From(
			"monocle.characters",
		).
		Where(
			s.LessThan("expires", sb.Raw("NOW()")),
			s.E("ignored", 0),
			s.NE("corporation_id", 1000001),
		).
		OrderBy("expires").Asc().
		Limit(limit)

	query, args := s.Build()

	err := db.Select(&characters, query, args...)
	return characters, err
}

func (db *DB) SelectCountOfExpiredCharacterEtags() (uint, error) {
	var count uint

	s := sb.NewSelectBuilder()
	s.Select(
		s.As("COUNT(*)", "count"),
	).From(
		"monocle.characters",
	).Where(
		s.LessThan("expires", sb.Raw("NOW()")),
		s.E("ignored", 0),
		s.NE("corporation_id", 1000001),
	)

	query, args := s.Build()
	err := db.Get(&count, query, args...)
	return count, err
}

func (db *DB) SelectCountOfCharacterEtags() (uint, error) {
	var count uint

	s := sb.NewSelectBuilder()
	s.Select(
		s.As("COUNT(*)", "count"),
	).From(
		"monocle.characters",
	)

	s.Where(
		s.E("ignored", 0),
		s.NE("corporation_id", 1000001),
	)

	query, args := s.Build()
	err := db.Get(&count, query, args...)
	return count, err
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

func (db *DB) SelectCharacterCorporationHistoryByID(id uint64) ([]monocle.CharacterCorporationHistory, error) {

	history := make([]monocle.CharacterCorporationHistory, 0)

	sb := sb.NewSelectBuilder()
	q := sb.Select(
		"id",
		"record_id",
		"corporation_id",
		"start_date",
		"created_at",
		"updated_at",
	).
		From("monocle.character_corporation_history").
		Where(
			sb.E("id", id),
		)

	query, args := q.Build()

	err := db.Select(&history, query, args...)
	return history, err
}

func (db *DB) InsertCharacterCorporationHistory(id uint64, history []monocle.CharacterCorporationHistory) ([]monocle.CharacterCorporationHistory, error) {

	ib := sb.NewInsertBuilder()
	q := ib.InsertIgnoreInto("monocle.character_corporation_history").Cols(
		"id",
		"record_id",
		"corporation_id",
		"start_date",
		"created_at",
		"updated_at",
	)
	for _, v := range history {
		q.Values(
			id,
			v.RecordID,
			v.CorporationID,
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

	return db.SelectCharacterCorporationHistoryByID(id)

}
