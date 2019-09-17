package resolvers

import (
	"context"
	"fmt"

	"github.com/ddouglas/monocle"
	"github.com/ddouglas/monocle/boiler"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type corporationResolver struct {
	*Common
}

func (q *corporationResolver) Members(ctx context.Context, obj *monocle.Corporation) ([]*monocle.Character, error) {

	var characters []*monocle.Character

	query := boiler.Characters(
		qm.Where(boiler.CharacterColumns.CorporationID+"=?", obj.ID),
	)

	queryStr, args := queries.BuildQuery(query.Query)
	fmt.Println(queryStr)
	fmt.Println(args...)

	err := query.Bind(ctx, q.DB, &characters)

	return characters, err

}
