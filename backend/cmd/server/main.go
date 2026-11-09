package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nik003kk/DeskBooker/backend/internal/auth"
    "github.com/nik003kk/DeskBooker/backend/internal/config"
	"github.com/nik003kk/DeskBooker/backend/internal/database"
	"github.com/nik003kk/DeskBooker/backend/internal/server"
    "github.com/nik003kk/DeskBooker/backend/internal/repository"
    "github.com/nik003kk/DeskBooker/backend/internal/resources"
    "github.com/nik003kk/DeskBooker/backend/internal/reservations"
    "github.com/nik003kk/DeskBooker/backend/internal/worker"
    "github.com/nik003kk/DeskBooker/backend/internal/analytics"
    "github.com/nik003kk/DeskBooker/backend/internal/adminusers"
)

func main() {
	if err := config.LoadLocalEnv(); err != nil {
		log.Fatalf("Napaka konfiguracije okolja: %v", err)
	}
	port, err := config.Port(os.Getenv)
	if err != nil {
		log.Fatalf("Neveljavna konfiguracija API: %v", err)
	}
	allowedOrigins, err := server.AllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if err != nil {
		log.Fatalf("Neveljavna CORS konfiguracija: %v", err)
	}

    tokens, err := auth.NewJWTService(os.Getenv("JWT_SECRET"))
    if err != nil {
        log.Fatalf("Neveljavna konfiguracija JWT: %v", err)
    }

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Open(ctx)
	if err != nil {
		log.Fatalf("Ne morem vzpostaviti povezave s PostgreSQL: %v", err)
	}
	defer pool.Close()
	log.Println("Povezava s PostgreSQL je uspešno vzpostavljena")

    store := repository.New(pool)
    workerDone := make(chan struct{})
    go func() {
        defer close(workerDone)
        if err := worker.Run(ctx, worker.NoShowInterval, store.ExpireNoShows, log.Printf); err != nil {
            log.Print("No-show worker: invalid configuration")
        }
    }()
    defer func() { stop(); <-workerDone }()

    router := server.NewRouterWithReservations(
            pool, allowedOrigins,
            auth.RegisterHandler(store, os.Getenv("REGISTRATION_EMAIL_DOMAIN")),
            auth.LoginHandlerWithJWT(store, tokens),
            auth.RequireCurrentJWT(tokens,store),
            auth.MeHandler(store),
            auth.ProfileHandler(store),
            auth.UpdateProfileHandler(store),
            auth.RequireResourceManager(store),
            resources.CreateHandler(store),
            resources.AddFeaturesHandler(store),
            resources.AdminListHandler(store),
            resources.EditHandler(store),
            resources.ActivationHandler(store),
            resources.AdminDetailHandler(store),
            resources.SearchHandler(store),
            resources.AvailabilityHandler(store),
            reservations.CreateHandler(repository.NewBookingStore(pool)),
            reservations.CheckInHandler(repository.NewBookingStore(pool)),
            reservations.DetailHandler(store),
        )
    server.RegisterAnalytics(router,auth.RequireCurrentJWT(tokens,store),auth.RequireAnalyticsManager(store),analytics.SummaryHandler(store),analytics.CSVHandler(store))
    server.RegisterAdminUsers(router,auth.RequireCurrentJWT(tokens,store),auth.RequireAdministrator(store),adminusers.ListHandler(store),adminusers.ChangeRoleHandler(repository.NewUserAdminStore(pool)),adminusers.ProtectDeletionHandler())
    srv := &http.Server{
        Addr: ":"+port,
        Handler: router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("DeskBooker API posluša na http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Napaka pri zagonu strežnika: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Ustavljam DeskBooker API ...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Napaka pri ustavljanju strežnika: %v", err)
	}
}

