package resolvers

import (
	"context"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func (r *queryResolver) Alliances(ctx context.Context, limit int) ([]*monocle.Alliance, error) {

	alliances := make([]*monocle.Alliance, 0)
	err := boiler.Alliances(
		qm.OrderBy("member_count DESC"),
		qm.Limit(limit),
	).Bind(ctx, r.DB, &alliances)

	return alliances, err
}
