package routes

import (
	"log"
	"log/slog"
	"net/http"
	"service-atlas/api/debt"
	"service-atlas/api/dependencies"
	"service-atlas/api/helloworld"
	"service-atlas/api/releases"
	"service-atlas/api/reports"
	"service-atlas/api/services"
	"service-atlas/api/system"
	"service-atlas/api/teams"
	"service-atlas/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/service-atlas/go-common/corsconfig"
	"github.com/service-atlas/go-common/httplog"
)

func SetupRouter(driver neo4j.DriverWithContext) http.Handler {
	slog.Debug("Setting up router")
	router := chi.NewRouter()

	router.Use(httplog.WebRequestLogger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))

	setupCORS(router)
	authCfg, err := auth.NewAuthConfig()
	setupSystemCalls(router, authCfg)

	serviceHandler := services.New(driver)
	debtHandler := debt.New(driver)
	dependencyHandler := dependencies.New(driver)
	releaseHandler := releases.New(driver)
	reportHandler := reports.New(driver)
	teamHandler := teams.New(driver)

	if err != nil {
		log.Fatal(err)
	}

	router.Group(func(r chi.Router) {

		r.Use(auth.Middleware(authCfg))

		r.Get("/releases/{startDate}/{endDate}", releaseHandler.GetReleasesInDateRange)
		r.Get("/reports/services/{id}/risk", reportHandler.GetComprehensiveRiskReport)
		r.Get("/reports/services/{id}/change_risk", reportHandler.GetServiceChangeRisk)
		r.Get("/reports/services/debt", reportHandler.GetServiceDebtReport)
		r.Get("/reports/services/tier", reportHandler.GetServicesByTier)
		r.Patch("/debt/{id}", debtHandler.UpdateDebtStatus)

		r.Route("/services", func(r chi.Router) {
			r.Get("/", serviceHandler.GetAllServices)
			r.Post("/", serviceHandler.CreateService)
			r.Get("/search", serviceHandler.Search)
			r.Get("/types", reportHandler.GetServiceTypes)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", serviceHandler.GetById)
				r.Put("/", serviceHandler.UpdateService)
				r.Delete("/", serviceHandler.DeleteServiceById)
				r.Get("/teams", serviceHandler.GetTeamsByServiceId)

				r.Get("/dependencies", dependencyHandler.GetDependencies)
				r.Get("/dependents", dependencyHandler.GetDependents)
				r.Post("/dependency", dependencyHandler.CreateDependency)
				r.Delete("/dependency/{id2}", dependencyHandler.DeleteDependency)

				r.Route("/debt", func(r chi.Router) {
					r.Post("/", debtHandler.CreateDebt)
					r.Get("/", debtHandler.GetDebtByServiceId)
				})

				r.Route("/release", func(r chi.Router) {
					r.Post("/", releaseHandler.CreateRelease)
					r.Get("/", releaseHandler.GetReleasesByServiceId)
				})

			})
		})

		r.Route("/teams", func(r chi.Router) {
			r.Post("/", teamHandler.CreateTeam)
			r.Get("/", teamHandler.GetTeams)
			r.Delete("/{id}", teamHandler.DeleteTeam)
			r.Get("/{id}", teamHandler.GetTeam)
			r.Put("/{id}", teamHandler.UpdateTeam)
			r.Route("/{teamId}/services/{serviceId}", func(r chi.Router) {
				r.Put("/", teamHandler.CreateTeamAssociation)
				r.Delete("/", teamHandler.DeleteTeamAssociation)
			})
			r.Get("/{teamId}/services", reportHandler.GetServicesByTeam)
		})
	})
	return router
}

func setupCORS(r chi.Router) {
	corsConfig := corsconfig.GetCORSConfig()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsConfig.AllowedOrigins,
		AllowedMethods:   corsConfig.AllowedMethods,
		AllowedHeaders:   corsConfig.AllowedHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           300, // Maximum value not ignored by any major browser
	}))
}

func setupSystemCalls(r chi.Router, cfg *auth.Config) {
	slog.Debug("Setting up system calls")
	r.Get("/time", system.GetTime)
	r.Get("/database", system.GetDbAddress)
	r.Get("/version", system.GetVersion)
	r.Get("/helloworld", helloworld.HelloWorld)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/auth/mcp/config", system.CreateMCPAuthEndpoint(cfg))
}
