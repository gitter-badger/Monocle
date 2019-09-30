package resolvers

import (
	"context"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func (r *queryResolver) Stat(ctx context.Context) (*monocle.Total, error) {

	stat := monocle.Total{}
	err := boiler.Totals(
		qm.OrderBy("created_at DESC"),
		qm.Limit(1),
	).Bind(ctx, r.DB, &stat)

	return &stat, err
}

func (r *queryResolver) Stats(ctx context.Context, limit int) ([]*monocle.Total, error) {
	stats := []*monocle.Total{}
	err := boiler.Totals(
		qm.OrderBy("created_at DESC"),
		qm.Limit(10),
	).Bind(ctx, r.DB, &stats)

	return stats, err
}
