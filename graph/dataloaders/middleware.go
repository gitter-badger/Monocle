package dataloaders

import (
	"context"
	"net/http"

	"github.com/ddouglas/monocle/graph/dataloaders/generated"
	"github.com/jmoiron/sqlx"
)

type ctxKeyType struct{ name string }

var ctxKey = ctxKeyType{"userCtx"}

type Loaders struct {
	Alliance                    *generated.AllianceLoader
	Character                   *generated.CharacterLoader
	CharacterCorporationHistory *generated.CharacterCorporationHistoryLoader
	Corporation                 *generated.CorporationLoader
	CorporationMembers          *generated.CorporationMembersLoader
	CorporationAllianceHistory  *generated.CorporationAllianceHistoryLoader
}

func Dataloader(db *sqlx.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		loaders := Loaders{
			Alliance:                    allianceLoader(ctx, db),
			Character:                   characterLoader(ctx, db),
			CharacterCorporationHistory: characterCorporationHistoryLoader(ctx, db),
			Corporation:                 corporationLoader(ctx, db),
			CorporationMembers:          corporationMembersLoader(ctx, db),
			CorporationAllianceHistory:  corporationAllianceHistoryLoader(ctx, db),
		}

		dataloadersCtx := context.WithValue(ctx, ctxKey, loaders)
		next.ServeHTTP(w, r.WithContext(dataloadersCtx))
	})
}

func CtxLoader(ctx context.Context) Loaders {
	return ctx.Value(ctxKey).(Loaders)
}
