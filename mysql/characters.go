package mysql

import (
	"github.com/ddouglas/eveindex"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) SelectCharacterByCharacterID(id uint64) (eveindex.Character, error) {

	var character eveindex.Character

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
		"expires",
		"etag",
		"created_at",
		"updated_at",
	).From(
		"eveindex.characters",
	).Where(
		s.E("id", id),
	).Limit(1)

	query, args := s.Build()

	err := db.Get(&character, query, args...)
	return character, err
}

func (db *DB) InsertCharacter(character eveindex.Character) (eveindex.Character, error) {

	i := sb.NewInsertBuilder()
	i.InsertIgnoreInto("eveindex.characters").Cols(
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

func (db *DB) UpdateCharacterByID(character eveindex.Character) (eveindex.Character, error) {

	u := sb.NewUpdateBuilder()
	u.Update("eveindex.characters").Set(
		u.E("security_status", character.SecurityStatus),
		u.E("alliance_id", character.AllianceID),
		u.E("corporation_id", character.CorporationID),
		u.E("faction_id", character.FactionID),
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
	d.DeleteFrom("eveindex.characters").Where(d.E("id", id))

	query, args := d.Build()

	_, err := db.Exec(query, args...)
	return err

}
