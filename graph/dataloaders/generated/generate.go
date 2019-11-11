//go:generate dataloaden AllianceLoader uint *github.com/ddouglas/monocle.Alliance
//go:generate dataloaden CharacterCorporationHistoryLoader uint64 []*github.com/ddouglas/monocle.CharacterCorporationHistory
//go:generate dataloaden CharacterLoader uint64 *github.com/ddouglas/monocle.Character
//go:generate dataloaden CorporationAllianceHistoryLoader uint []*github.com/ddouglas/monocle.CorporationAllianceHistory
//go:generate dataloaden CorporationLoader uint *github.com/ddouglas/monocle.Corporation
//go:generate dataloaden CorporationMembersLoader uint []*github.com/ddouglas/monocle.Character

package generated
