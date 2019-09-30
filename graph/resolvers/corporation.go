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

func (r *queryResolver) CorporationsByMemberCount(ctx context.Context, limit int) ([]*monocle.Corporation, error) {
	corporations := make([]*monocle.Corporation, 0)

	if limit > 50 {
		limit = 50
	}

	err := boiler.Corporations(
		qm.OrderBy("member_count DESC"),
		qm.Limit(limit),
	).Bind(ctx, r.DB, &corporations)

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
