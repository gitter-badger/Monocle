package resolvers

import (
	"context"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/graph/dataloaders"
)

type characterResolver struct {
	*Common
}

func (q *characterResolver) Corporation(ctx context.Context, obj *monocle.Character) (*monocle.Corporation, error) {
	return dataloaders.CtxLoader(ctx).Corporation.Load(int(obj.CorporationID))
}
