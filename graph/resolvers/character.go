package resolvers

import (
	"context"
	"fmt"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/graph/dataloaders"
	"github.com/ddouglas/monocle/graph/models"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func (r *queryResolver) Character(ctx context.Context, id int) (*monocle.Character, error) {

	var character monocle.Character

	err := boiler.Characters(
		qm.Where(boiler.CharacterColumns.ID+"=?", id),
	).Bind(ctx, r.DB, &character)

	return &character, err
}

func (r *queryResolver) CharactersByID(ctx context.Context, limit int, order models.Sort) ([]*monocle.Character, error) {

	var characters []*monocle.Character

	if limit > 100 {
		limit = 100
	}

	err := boiler.Characters(
		qm.Limit(limit),
		qm.OrderBy(boiler.CharacterColumns.ID+" "+order.String()),
	).Bind(ctx, r.DB, &characters)

	return characters, err
}

func (r *queryResolver) CharactersByBirthday(ctx context.Context, limit int, order *models.Sort) ([]*monocle.Character, error) {
	var characters []*monocle.Character

	if limit > 100 {
		limit = 100
	}

	err := boiler.Characters(
		qm.Where("birthday = DATE('%c-%d', CURDATE())"),
		qm.Limit(limit),
		qm.OrderBy(fmt.Sprintf(
			"birthday %s",
			order.String(),
		)),
	).Bind(ctx, r.DB, &characters)

	return characters, err
}

func (r *queryResolver) CharactersByAllianceID(ctx context.Context, allianceID int, page *int) ([]*monocle.Character, error) {

	characters := make([]*monocle.Character, 0)

	limit := 50
	offset := (*page * limit) - limit

	err := boiler.Characters(
		qm.Where("alliance_id = ?", allianceID),
		qm.Limit(limit),
		qm.Offset(offset),
		qm.OrderBy("birthday DESC"),
	).Bind(ctx, r.DB, &characters)

	return characters, err

}

func (r *queryResolver) CharactersByCorporationID(ctx context.Context, corporationID int, page *int) ([]*monocle.Character, error) {
	characters := make([]*monocle.Character, 0)

	limit := 50
	offset := (*page * limit) - limit

	err := boiler.Characters(
		qm.Where("corporation_id = ?", corporationID),
		qm.Limit(limit),
		qm.Offset(offset),
	).Bind(ctx, r.DB, &characters)

	return characters, err
}

type characterResolver struct {
	*Common
}

func (q *characterResolver) Corporation(ctx context.Context, obj *monocle.Character) (*monocle.Corporation, error) {
	return dataloaders.CtxLoader(ctx).Corporation.Load(obj.CorporationID)
}

func (q *characterResolver) History(ctx context.Context, obj *monocle.Character) ([]*monocle.CharacterCorporationHistory, error) {
	return dataloaders.CtxLoader(ctx).CharacterCorporationHistory.Load(obj.ID)
}

type corporationHistoryResolver struct {
	*Common
}

func (q *corporationHistoryResolver) Corporation(ctx context.Context, obj *monocle.CharacterCorporationHistory) (*monocle.Corporation, error) {
	return dataloaders.CtxLoader(ctx).Corporation.Load(obj.CorporationID)
}
