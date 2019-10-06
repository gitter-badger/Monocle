package resolvers

import (
	"context"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/graph/dataloaders"
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

type corporationResolver struct {
	*Common
}

func (q *corporationResolver) Alliance(ctx context.Context, obj *monocle.Corporation) (*monocle.Alliance, error) {
	return dataloaders.CtxLoader(ctx).Alliance.Load(obj.AllianceID.Uint32)
}

func (q *corporationResolver) Members(ctx context.Context, obj *monocle.Corporation) ([]*monocle.Character, error) {
	return dataloaders.CtxLoader(ctx).CorporationMembers.Load(obj.ID)
}
