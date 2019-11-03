package mysql

import (
	"github.com/ddouglas/monocle"
	sb "github.com/huandu/go-sqlbuilder"
)

func (db *DB) InsertCharacter(character *monocle.Character) error {

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

	return err

}

func (db *DB) UpdateCharacterByID(character *monocle.Character) error {

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

	return err
}

func (db *DB) InsertCharacterCorporationHistory(id uint64, history []*monocle.CharacterCorporationHistory) error {

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

	return err

}
