package router

import (
	"net/http"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	m "github.com/rickbassham/goapi/middleware"
)

type router struct {
	address string
	r       *chi.Mux
	logger  m.RequestLogger
}

type RouteCreater interface {
	CreateRoutes(r chi.Router) chi.Router
}

type Router interface {
	ListenAndServe() error
}

func NewRouter(address string, logger m.RequestLogger, routeCreater RouteCreater) Router {
	r := chi.NewRouter()

	recoverer := m.NewRecoverer(logger)

	r.Group(func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "ok"}`))
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.DefaultCompress)
		r.Use(logger.Handler())
		r.Use(recoverer.Handler())

		limiter := tollbooth.NewLimiter(10, nil)

		r.Use(tollbooth_chi.LimitHandler(limiter))

		r.Route("/api", func(r chi.Router) {
			routeCreater.CreateRoutes(r)
		})
	})

	return &router{
		address: address,
		r:       r,
		logger:  logger,
	}
}

func (svc *router) ListenAndServe() error {
	return http.ListenAndServe(svc.address, svc.r)
}
