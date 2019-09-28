package resolvers

import (
	generated "github.com/ddouglas/monocle/graph/service"
	"github.com/jmoiron/sqlx"
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

func (r *Common) Corporation() generated.CorporationResolver {
	return &corporationResolver{r}
}

func (r *Common) CorporationHistory() generated.CorporationHistoryResolver {
	return &corporationHistoryResolver{r}
}

type queryResolver struct{ *Common }

type mutationResolver struct{ *Common }
