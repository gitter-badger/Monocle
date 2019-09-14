package resolvers

import (
	"context"
	"fmt"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/ddouglas/monocle/graph/models"
	generated "github.com/ddouglas/monocle/graph/service"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/queries"
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

func (r *queryResolver) Characters(ctx context.Context, limit int, order models.Sort) ([]*monocle.Character, error) {
	var characters []*monocle.Character

	if !order.IsValid() {
		return characters, fmt.Errorf("%s is not a valid sort value", order.String())
	}

	query := boiler.Characters(
		qm.Limit(limit),
		qm.OrderBy(boiler.CharacterColumns.Expires+" "+order.String()),
	)

	queryStr, args := queries.BuildQuery(query.Query)
	fmt.Println(queryStr)
	fmt.Println(args...)

	err := query.Bind(context.Background(), r.DB, &characters)

	return characters, err
}

func (r *queryResolver) Corporation(ctx context.Context, id int) (*monocle.Corporation, error) {
	return nil, nil
}
