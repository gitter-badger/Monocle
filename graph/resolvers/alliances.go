package resolvers

import (
	"context"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/graph/dataloaders"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func (r *queryResolver) Alliance(ctx context.Context, id int) (*monocle.Alliance, error) {

	var alliance monocle.Alliance
	err := boiler.Alliances(
		qm.Where("id = ?", id),
	).Bind(ctx, r.DB, &alliance)

	return &alliance, err
}

func (r *queryResolver) AlliancesByMemberCount(ctx context.Context, limit int) ([]*monocle.Alliance, error) {

	alliances := make([]*monocle.Alliance, 0)
	err := boiler.Alliances(
		qm.OrderBy("member_count DESC"),
		qm.Limit(limit),
	).Bind(ctx, r.DB, &alliances)

	return alliances, err

}

type allianceResolver struct {
	*Common
}

func (a *allianceResolver) Creator(ctx context.Context, obj *monocle.Alliance) (*monocle.Character, error) {
	return dataloaders.CtxLoader(ctx).Character.Load(obj.CreatorID)
}
func (a *allianceResolver) CreatorCorp(ctx context.Context, obj *monocle.Alliance) (*monocle.Corporation, error) {
	return dataloaders.CtxLoader(ctx).Corporation.Load(obj.CreatorCorporationID)
}
func (a *allianceResolver) Executor(ctx context.Context, obj *monocle.Alliance) (*monocle.Corporation, error) {
	return dataloaders.CtxLoader(ctx).Corporation.Load(obj.ExecutorCorporationID)
}

type allianceHistoryResolver struct {
	*Common
}

func (ah *allianceHistoryResolver) Alliance(ctx context.Context, obj *monocle.CorporationAllianceHistory) (*monocle.Alliance, error) {
	return dataloaders.CtxLoader(ctx).Alliance.Load(obj.AllianceID.Uint)
}
