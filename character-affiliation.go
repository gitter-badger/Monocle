package monocle

import "github.com/volatiletech/null"

type CharacterAffiliation struct {
	CharacterID   uint64      `json:"character_id"`
	CorporationID uint32      `json:"corporation_id"`
	AllianceID    null.Uint32 `json:"alliance_id"`
	FactionID     null.Uint32 `json:"faction_id"`
}
