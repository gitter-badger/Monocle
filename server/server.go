package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/time/rate"

	"github.com/ddouglas/monocle/core"
	"github.com/ddouglas/monocle/graph/dataloaders"
	"github.com/ddouglas/monocle/graph/resolvers"
	"github.com/ddouglas/monocle/graph/service"
)

type Server struct {
	App      *core.App
	visitors map[string]*visitor
	server   *http.Server
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var mtx sync.Mutex

func New(port uint) (*Server, error) {
	core, err := core.New()
	if err != nil {
		err = errors.Wrap(err, "Unable to create core application")

		return &Server{}, err
	}

	x := &Server{
		App: core,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		visitors: make(map[string]*visitor, 0),
	}

	x.server.Handler = x.RegisterRoutes()
	return x, nil
}

func Serve(c *cli.Context) {
	port := c.Uint("port")

	api, err := New(port)
	if err != nil {
		log.Fatal(err)
	}

	api.App.Logger.Infof("Starting Server on port: %d", port)

	go func() {
		if err := api.server.ListenAndServe(); err != nil {
			api.App.Logger.Infof("unable to start http server: %s", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	api.App.Logger.Info("Shutting Down Server")

	_ = api.GracefullyShutdown(context.Background())
}

func (s *Server) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()

	// r.Use(Cors)
	r.Use(s.RequestLogger)
	r.Use(s.RateLimiter)

	graphSchema := service.NewExecutableSchema(service.Config{
		Resolvers: &resolvers.Common{DB: s.App.DB.DB},
	})

	// One handler to process graphQL queries
	queryHandler := handler.GraphQL(
		graphSchema,
		// handler.IntrospectionEnabled(false),
		// handler.ComplexityLimit(viper.GetInt("api.graphql.complexity_limit")),
	)

	r.Handle("/pg", dataloaders.Dataloader(s.App.DB.DB, handler.Playground("GraphQL playground", "/query")))
	r.Handle("/query", dataloaders.Dataloader(s.App.DB.DB, queryHandler))

	return r
}

// GracefullyShutdown gracefully shuts down the HTTP API.
func (s *Server) GracefullyShutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) WriteSuccess(w http.ResponseWriter, data interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")

	if status != 0 {
		w.WriteHeader(status)
	}

	return json.NewEncoder(w).Encode(data)
}

func (s *Server) WriteError(w http.ResponseWriter, code int, err error) error {
	w.Header().Set("Content-Type", "application-type/json")
	w.WriteHeader(code)

	if err == nil {
		err = errors.New(http.StatusText(code))
	}
	s.App.Logger.Infof("Code: %d Error: %s", code, err)

	res := struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	}

	return json.NewEncoder(w).Encode(res)
}
