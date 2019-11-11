package monocle

import "github.com/volatiletech/null"

type CharacterAffiliation struct {
	CharacterID   uint64    `json:"character_id"`
	CorporationID uint      `json:"corporation_id"`
	AllianceID    null.Uint `json:"alliance_id"`
	FactionID     null.Uint `json:"faction_id"`
}
