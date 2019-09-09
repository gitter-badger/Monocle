package resolvers

import (
	"context"

	"github.com/ddouglas/monocle"
	generated "github.com/ddouglas/monocle/graph/service"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Character() generated.CharacterResolver {
	return &characterResolver{r}
}
func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

type characterResolver struct{ *Resolver }

func (r *characterResolver) ID(ctx context.Context, obj *monocle.Character) (string, error) {
	panic("not implemented")
}
func (r *characterResolver) SecurityStatus(ctx context.Context, obj *monocle.Character) (string, error) {
	panic("not implemented")
}
func (r *characterResolver) AllianceID(ctx context.Context, obj *monocle.Character) (*string, error) {
	panic("not implemented")
}
func (r *characterResolver) CorporationID(ctx context.Context, obj *monocle.Character) (string, error) {
	panic("not implemented")
}
func (r *characterResolver) FactionID(ctx context.Context, obj *monocle.Character) (*string, error) {
	panic("not implemented")
}
func (r *characterResolver) AncestryID(ctx context.Context, obj *monocle.Character) (string, error) {
	panic("not implemented")
}
func (r *characterResolver) BloodlineID(ctx context.Context, obj *monocle.Character) (string, error) {
	panic("not implemented")
}
func (r *characterResolver) RaceID(ctx context.Context, obj *monocle.Character) (string, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Character(ctx context.Context, id int) (*monocle.Character, error) {
	panic("not implemented")
}
