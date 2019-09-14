package resolvers

import (
	"context"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	generated "github.com/ddouglas/monocle/graph/service"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Common struct {
	DB *sqlx.DB
}

func (r *Common) Query() generated.QueryResolver {
	return &queryResolver{r}
}

func (r *Common) Character() generated.CharacterResolver {
	return &characterResolver{r}
}

type queryResolver struct{ *Common }

func (r *queryResolver) Character(ctx context.Context, id int) (*monocle.Character, error) {

	var character monocle.Character

	err := boiler.Characters(
		qm.Where(boiler.CharacterColumns.ID+"=?", id),
	).Bind(ctx, r.DB, &character)

	return &character, err
}

func (r *queryResolver) Characters(ctx context.Context, limit int) ([]*monocle.Character, error) {
	var characters []*monocle.Character

	err := boiler.Characters(
		qm.Where(boiler.CharacterColumns.Expires+"<NOW()"),
		qm.And(boiler.CharacterColumns.Ignored+"=?", 0),
		qm.And(boiler.CharacterColumns.CorporationID+"!=?", 1000001),
		qm.OrderBy(boiler.CharacterColumns.Expires),
		qm.Limit(limit),
	).Bind(context.Background(), r.DB, &characters)

	return characters, err
}

func (r *queryResolver) Corporation(ctx context.Context, id int) (*monocle.Corporation, error) {
	return nil, nil
}
