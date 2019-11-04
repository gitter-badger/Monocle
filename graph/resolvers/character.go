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

func (r *queryResolver) CharacterCorporationHistoryByAllianceID(ctx context.Context, allianceID int, page *int, limit *int) ([]*monocle.CharacterCorporationHistory, error) {
	histories := make([]*monocle.CharacterCorporationHistory, 0)

	if limit == nil || *limit > 50 {
		x := 50
		limit = &x
	}

	offset := (*page * *limit) - *limit
	/**
		SELECT
			`character_corporation_history`.*
		FROM `character_corporation_history`
		INNER JOIN corporations crp on crp.id = character_corporation_history.corporation_id
		WHERE (crp.alliance_id = ?)
		AND (character_corporation_history.leave_date IS NULL)
		ORDER BY record_id DESC
		LIMIT 50
		OFFSET 50;
	**/

	err := boiler.CharacterCorporationHistories(
		qm.InnerJoin(
			fmt.Sprintf(
				"%s %s on %s.%s = %s.%s",
				boiler.TableNames.Corporations,
				"crp",
				"crp",
				"id",
				boiler.TableNames.CharacterCorporationHistory,
				"corporation_id",
			),
		),
		qm.Where("crp.alliance_id = ?", allianceID),
		qm.And(fmt.Sprintf("%s.%s IS NULL", boiler.TableNames.CharacterCorporationHistory, "leave_date")),
		qm.Limit(*limit),
		qm.Offset(offset),
		qm.OrderBy("record_id DESC"),
	).Bind(ctx, r.DB, &histories)

	return histories, err

}

func (r *queryResolver) CharacterCorporationHistoryByCorporationID(ctx context.Context, corporationID int, page *int, limit *int) ([]*monocle.CharacterCorporationHistory, error) {
	histories := make([]*monocle.CharacterCorporationHistory, 0)

	if limit == nil || *limit > 50 {
		x := 50
		limit = &x
	}

	offset := (*page * *limit) - *limit

	/**
		SELECT
			*
		FROM `character_corporation_history`
		WHERE (corporation_id = ?)
		AND (leave_date IS NULL)
		ORDER BY record_id DESC
		LIMIT 50
		OFFSET 50;
	**/

	err := boiler.CharacterCorporationHistories(
		qm.Where("corporation_id = ?", corporationID),
		qm.And("leave_date IS NULL"),
		qm.Limit(*limit),
		qm.Offset(offset),
		qm.OrderBy("record_id DESC"),
	).Bind(ctx, r.DB, &histories)

	return histories, err
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

func (q *corporationHistoryResolver) Character(ctx context.Context, obj *monocle.CharacterCorporationHistory) (*monocle.Character, error) {
	return dataloaders.CtxLoader(ctx).Character.Load(obj.ID)
}

func (q *corporationHistoryResolver) Corporation(ctx context.Context, obj *monocle.CharacterCorporationHistory) (*monocle.Corporation, error) {
	return dataloaders.CtxLoader(ctx).Corporation.Load(obj.CorporationID)
}
