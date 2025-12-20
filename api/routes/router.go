package routes

import (
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
	"service-atlas/internal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func SetupRouter(driver neo4j.DriverWithContext) http.Handler {
	slog.Debug("Setting up router")
	router := chi.NewRouter()

	router.Use(internal.RequestIDLogger)
	router.Use(internal.StructuredLoggerFromContext())
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))
	setupSystemCalls(router)

	serviceHandler := services.New(driver)
	debtHandler := debt.New(driver)
	dependencyHandler := dependencies.New(driver)
	releaseHandler := releases.New(driver)
	reportHandler := reports.New(driver)
	teamHandler := teams.New(driver)

	router.Get("/releases/{startDate}/{endDate}", releaseHandler.GetReleasesInDateRange)
	router.Get("/reports/services/{id}/risk", reportHandler.GetServiceRiskReport)
	router.Get("/reports/services/debt", reportHandler.GetServiceDebtReport)
	router.Get("/reports/services/tier", reportHandler.GetServicesByTier)
	router.Patch("/debt/{id}", debtHandler.UpdateDebtStatus)

	router.Route("/services", func(r chi.Router) {
		r.Get("/", serviceHandler.GetAllServices)
		r.Post("/", serviceHandler.CreateService)
		r.Get("/search", serviceHandler.Search)

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

	router.Route("/teams", func(r chi.Router) {
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
	return router
}

func setupSystemCalls(r chi.Router) {
	slog.Debug("Setting up system calls")
	r.Get("/time", system.GetTime)
	r.Get("/database", system.GetDbAddress)
	r.Get("/helloworld", helloworld.HelloWorld)
}
