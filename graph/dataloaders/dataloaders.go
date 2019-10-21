package dataloaders

import (
	"context"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ddouglas/monocle/boiler"
	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/ddouglas/monocle"

	"github.com/ddouglas/monocle/graph/dataloaders/generated"
	"github.com/jmoiron/sqlx"
)

const defaultWait = 2 * time.Millisecond
const defaultMaxBatch = 100

func allianceLoader(ctx context.Context, db *sqlx.DB) *generated.AllianceLoader {
	return generated.NewAllianceLoader(generated.AllianceLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint32) ([]*monocle.Alliance, []error) {

			alliances := make([]*monocle.Alliance, len(ids))
			errors := make([]error, len(ids))

			var whereIDs []interface{}
			for _, c := range ids {
				whereIDs = append(whereIDs, c)
			}

			allAlliances := make([]*monocle.Alliance, 0)
			err := boiler.Alliances(
				qm.WhereIn("id IN ?", whereIDs...),
			).Bind(ctx, db, &allAlliances)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			allianceByAllianceID := map[uint32]*monocle.Alliance{}
			for _, c := range allAlliances {
				allianceByAllianceID[c.ID] = c
			}

			for i, x := range ids {
				alliances[i] = allianceByAllianceID[x]
			}

			return alliances, nil

		},
	})
}

func characterLoader(ctx context.Context, db *sqlx.DB) *generated.CharacterLoader {
	return generated.NewCharacterLoader(generated.CharacterLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([]*monocle.Character, []error) {

			characters := make([]*monocle.Character, len(ids))
			errors := make([]error, len(ids))

			var whereIDs []interface{}
			for _, c := range ids {
				whereIDs = append(whereIDs, c)
			}

			allCharacters := make([]*monocle.Character, 0)
			err := boiler.Characters(
				qm.WhereIn("id IN ?", whereIDs...),
			).Bind(ctx, db, &allCharacters)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			characterByCharacterID := map[uint64]*monocle.Character{}
			for _, c := range allCharacters {
				characterByCharacterID[c.ID] = c
			}

			for i, x := range ids {
				characters[i] = characterByCharacterID[x]
			}

			return characters, nil

		},
	})
}

func corporationLoader(ctx context.Context, db *sqlx.DB) *generated.CorporationLoader {
	return generated.NewCorporationLoader(generated.CorporationLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint32) ([]*monocle.Corporation, []error) {
			corporations := make([]*monocle.Corporation, len(ids))
			errors := make([]error, len(ids))

			var whereIds []interface{}
			for _, c := range ids {
				whereIds = append(whereIds, c)
			}

			allCorporations := make([]*monocle.Corporation, 0)
			err := boiler.Corporations(
				qm.WhereIn("id IN ?", whereIds...),
			).Bind(ctx, db, &allCorporations)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			corporationByCorporationID := map[uint32]*monocle.Corporation{}
			for _, c := range allCorporations {
				corporationByCorporationID[c.ID] = c
			}

			for i, x := range ids {
				corporations[i] = corporationByCorporationID[x]
			}

			return corporations, nil
		},
	})
}

func corporationMembersLoader(ctx context.Context, db *sqlx.DB) *generated.CorporationMembersLoader {
	return generated.NewCorporationMembersLoader(generated.CorporationMembersLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint32) ([][]*monocle.Character, []error) {
			corporationMembers := make([][]*monocle.Character, len(ids))
			errors := make([]error, len(ids))

			var whereIDs []interface{}
			for _, i := range ids {
				whereIDs = append(whereIDs, i)
			}

			allCharacters := make([]*monocle.Character, 0)
			err := boiler.Characters(
				qm.WhereIn("corporation_id IN ?", whereIDs...),
			).Bind(ctx, db, &allCharacters)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			membersByCorporationID := map[uint32][]*monocle.Character{}
			for _, c := range allCharacters {
				membersByCorporationID[c.CorporationID] = append(membersByCorporationID[c.CorporationID], c)
			}

			for i, x := range ids {
				corporationMembers[i] = membersByCorporationID[x]
			}

			return corporationMembers, nil
		},
	})
}

func corporationAllianceHistoryLoader(ctx context.Context, db *sqlx.DB) *generated.CorporationAllianceHistoryLoader {
	return generated.NewCorporationAllianceHistoryLoader(generated.CorporationAllianceHistoryLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(corporationIDs []uint32) ([][]*monocle.CorporationAllianceHistory, []error) {
			histories := make([][]*monocle.CorporationAllianceHistory, len(corporationIDs))
			errors := make([]error, len(corporationIDs))

			var whereIDs []interface{}
			for _, i := range corporationIDs {
				whereIDs = append(whereIDs, i)
			}

			allHistories := make([]*monocle.CorporationAllianceHistory, 0)
			err := boiler.CorporationAllianceHistories(
				qm.WhereIn("id IN ?", whereIDs...),
				qm.OrderBy("record_id DESC"),
			).Bind(ctx, db, &allHistories)
			if err != nil {
				errors = append(errors, err)
				spew.Dump(errors)
				return nil, errors
			}

			historiesByCorporationID := map[uint32][]*monocle.CorporationAllianceHistory{}
			for _, h := range allHistories {
				historiesByCorporationID[uint32(h.ID)] = append(historiesByCorporationID[uint32(h.ID)], h)
			}

			for i, x := range corporationIDs {
				histories[i] = historiesByCorporationID[x]
			}

			return histories, nil
		},
	})
}

func characterCorporationHistoryLoader(ctx context.Context, db *sqlx.DB) *generated.CharacterCorporationHistoryLoader {
	return generated.NewCharacterCorporationHistoryLoader(generated.CharacterCorporationHistoryLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([][]*monocle.CharacterCorporationHistory, []error) {

			characterHistory := make([][]*monocle.CharacterCorporationHistory, len(ids))
			errors := make([]error, len(ids))

			var whereIds []interface{}
			for _, i := range ids {
				whereIds = append(whereIds, i)
			}

			allHistories := make([]*monocle.CharacterCorporationHistory, 0)
			err := boiler.CharacterCorporationHistories(
				qm.WhereIn("id IN ?", whereIds...),
				qm.OrderBy("record_id DESC"),
			).Bind(ctx, db, &allHistories)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			historyByCharacterID := map[uint64][]*monocle.CharacterCorporationHistory{}
			for _, c := range allHistories {
				historyByCharacterID[c.ID] = append(historyByCharacterID[c.ID], c)
			}

			for i, c := range ids {
				characterHistory[i] = historyByCharacterID[c]
			}

			return characterHistory, nil

		},
	})
}
