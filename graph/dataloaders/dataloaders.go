package dataloaders

import (
	"context"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/volatiletech/sqlboiler/queries"

	"github.com/ddouglas/monocle/boiler"
	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/ddouglas/monocle"

	"github.com/ddouglas/monocle/graph/dataloaders/generated"
	"github.com/jmoiron/sqlx"
)

const defaultWait = 2 * time.Millisecond
const defaultMaxBatch = 100

func corporationsLoader(ctx context.Context, db *sqlx.DB) *generated.CorporationLoader {
	return generated.NewCorporationLoader(generated.CorporationLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []int) ([]*monocle.Corporation, []error) {
			corporations := make([]*monocle.Corporation, 1)
			errors := make([]error, len(ids))

			var whereIds []interface{}
			for _, c := range ids {
				whereIds = append(whereIds, c)
			}

			query := boiler.Corporations(
				qm.WhereIn(boiler.CorporationColumns.ID+" IN ?", whereIds...),
			)

			queryStr, args := queries.BuildQuery(query.Query)
			fmt.Println(queryStr)
			fmt.Println(args...)

			err := query.Bind(ctx, db, &corporations)
			if err != nil {
				errors = append(errors, err)
			}
			spew.Dump(corporations)

			return corporations, errors
		},
	})
}
