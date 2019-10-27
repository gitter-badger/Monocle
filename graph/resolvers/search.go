package resolvers

import (
	"context"

	"github.com/ddouglas/monocle"
)

func (q *queryResolver) SearchEntity(ctx context.Context, term string) ([]*monocle.Entity, error) {

	query := `
		(SELECT id, name, "character" AS category FROM characters WHERE NAME LIKE :term LIMIT 10)
		UNION ALL
		(SELECT id, name, "corporation" AS category FROM corporations WHERE NAME LIKE :term LIMIT 10)
		UNION ALL
		(SELECT id, name, "alliance" AS category FROM alliances WHERE NAME LIKE :term LIMIT 10)
	`
	var entities []*monocle.Entity

	stmt, err := q.DB.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	var params = map[string]interface{}{
		"term": term + "%",
	}
	err = stmt.SelectContext(ctx, &entities, params)
	if err != nil {
		return nil, err
	}

	return entities, nil

}
