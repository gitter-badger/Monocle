package resolvers

import (
	"context"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/graph/dataloaders"
	"github.com/ddouglas/monocle/graph/models"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func (r *queryResolver) Corporation(ctx context.Context, id int) (*monocle.Corporation, error) {

	var corporation monocle.Corporation

	err := boiler.Corporations(
		qm.Where("id = ?", id),
	).Bind(ctx, r.DB, &corporation)

	return &corporation, err

}

func (r *queryResolver) CorporationsByMemberCount(ctx context.Context, limit int, independent bool, npc bool) ([]*monocle.Corporation, error) {
	corporations := make([]*monocle.Corporation, 0)

	if limit > 50 {
		limit = 50
	}

	mods := []qm.QueryMod{}

	if !independent {
		mods = append(mods, qm.Where("alliance_id IS NOT NULL"))
	}

	if !npc {
		mods = append(mods, qm.Where("id >= 98000000"))
	}

	mods = append(mods, qm.OrderBy("member_count DESC"))
	mods = append(mods, qm.Limit(limit))

	err := boiler.Corporations(mods...).Bind(ctx, r.DB, &corporations)

	return corporations, err
}

func (r *queryResolver) CorporationsByAllianceID(ctx context.Context, allianceID int, page int) ([]*monocle.Corporation, error) {

	corporations := make([]*monocle.Corporation, 0)

	offset := (page * 100) - 100

	err := boiler.Corporations(
		qm.Where("alliance_id = ?", allianceID),
		qm.Limit(100),
		qm.Offset(offset),
	).Bind(ctx, r.DB, &corporations)

	return corporations, err
}

func (r *queryResolver) CorporationAllianceHistoryByAllianceID(ctx context.Context, allianceID int, page *int, limit *int, sort *models.Sort) ([]*monocle.CorporationAllianceHistory, error) {
	// histories := make([]*monocle.CorporationAllianceHistory, 0)

	// if limit == nil || *limit > 50 {
	// 	x := 50
	// 	limit = &x
	// }

	// offset := (*page * *limit) - *limit

	// /**
	// 	SELECT
	// 		*
	// 	FROM `character_corporation_history`
	// 	WHERE (corporation_id = ?)
	// 	AND (leave_date IS NULL)
	// 	ORDER BY record_id DESC
	// 	LIMIT 50
	// 	OFFSET 50;
	// **/

	// err := boiler.CharacterCorporationHistories(
	// 	qm.Where("corporation_id = ?", corporationID),
	// 	qm.And("leave_date IS NULL"),
	// 	qm.Limit(*limit),
	// 	qm.Offset(offset),
	// 	qm.OrderBy(fmt.Sprintf("record_id %s", sort.String())),
	// ).Bind(ctx, r.DB, &histories)

	// return histories, err
	return nil, nil
}

type corporationResolver struct {
	*Common
}

func (q *corporationResolver) Alliance(ctx context.Context, obj *monocle.Corporation) (*monocle.Alliance, error) {
	return dataloaders.CtxLoader(ctx).Alliance.Load(obj.AllianceID.Uint32)
}

func (q *corporationResolver) Members(ctx context.Context, obj *monocle.Corporation) ([]*monocle.Character, error) {
	return dataloaders.CtxLoader(ctx).CorporationMembers.Load(obj.ID)
}

func (q *corporationResolver) History(ctx context.Context, obj *monocle.Corporation) ([]*monocle.CorporationAllianceHistory, error) {
	return dataloaders.CtxLoader(ctx).CorporationAllianceHistory.Load(obj.ID)
}

func (q *corporationResolver) Ceo(ctx context.Context, obj *monocle.Corporation) (*monocle.Character, error) {
	return dataloaders.CtxLoader(ctx).Character.Load(obj.CreatorID)
}
