package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gasflow/api/internal/gateway"
	"github.com/gasflow/api/internal/infra/auth"
	"github.com/gasflow/api/internal/infra/db"
	"github.com/gasflow/api/internal/infra/messaging"

	analyticsM "github.com/gasflow/api/internal/modules/analytics"
	billingM "github.com/gasflow/api/internal/modules/billing"
	clientsM "github.com/gasflow/api/internal/modules/clients"
	inventoryM "github.com/gasflow/api/internal/modules/inventory"
	ordersM "github.com/gasflow/api/internal/modules/orders"
)

func main() {
	// ── Logging estruturado ───────────────────────────────────────────────────
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if getenv("ENV", "production") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"})
	}

	log.Info().Str("env", getenv("ENV", "production")).Msg("gasflow: starting")

	// ── Banco de dados ────────────────────────────────────────────────────────
	database := db.MustConnect(db.Config{
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenv("DB_PORT", "3306"),
		User:     getenv("DB_USER", "gasflow"),
		Password: getenv("DB_PASSWORD", "gasflow"),
		Name:     getenv("DB_NAME", "gasflow"),
	})
	defer database.Close()

	// ── RabbitMQ ──────────────────────────────────────────────────────────────
	mq := messaging.MustConnect(messaging.Config{
		URL: getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	})
	defer mq.Close()

	// ── Auth JWT ──────────────────────────────────────────────────────────────
	authSvc := auth.NewService(auth.Config{
		Secret:          getenv("JWT_SECRET", "change-me-in-production-min-32-chars!!"),
		AccessTokenTTL:  8 * time.Hour,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	})

	// ── Repositórios ──────────────────────────────────────────────────────────
	clientsRepo   := clientsM.NewRepository(database)
	ordersRepo    := ordersM.NewRepository(database)
	inventoryRepo := inventoryM.NewRepository(database)
	billingRepo   := billingM.NewRepository(database)
	analyticsRepo := analyticsM.NewRepository(database)

	// ── Serviços ──────────────────────────────────────────────────────────────
	defaultDeposit := getenv("DEFAULT_DEPOSIT_ID", "dep-sp-001")

	clientsSvc   := clientsM.NewService(clientsRepo)
	ordersSvc    := ordersM.NewService(ordersRepo, mq)
	inventorySvc := inventoryM.NewService(inventoryRepo, mq, defaultDeposit)
	billingSvc   := billingM.NewService(billingRepo, mq)
	analyticsSvc := analyticsM.NewService(analyticsRepo)

	// ── Context com cancelamento para graceful shutdown ───────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Consumers RabbitMQ ────────────────────────────────────────────────────
	consumers := []func(context.Context) error{
		ordersM.NewConsumer(ordersSvc, mq).Start,
		inventoryM.NewConsumer(inventorySvc, mq).Start,
		billingM.NewConsumer(billingSvc, mq, 10500).Start, // preço padrão P13 em centavos
	}
	for _, start := range consumers {
		if err := start(ctx); err != nil {
			log.Fatal().Err(err).Msg("failed to start consumer")
		}
	}

	// ── Job periódico: marca cobranças vencidas ───────────────────────────────
	go func() {
		// Roda imediatamente na inicialização, depois a cada hora
		billingSvc.RunOverdueJob(context.Background())
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				billingSvc.RunOverdueJob(context.Background())
			}
		}
	}()

	// ── Handlers HTTP ─────────────────────────────────────────────────────────
	handlers := gateway.Handlers{
		Clients:   clientsM.NewHandler(clientsSvc),
		Orders:    ordersM.NewHandler(ordersSvc),
		Inventory: inventoryM.NewHandler(inventorySvc),
		Billing:   billingM.NewHandler(billingSvc),
		Analytics: analyticsM.NewHandler(analyticsSvc),
		Auth:      authSvc,
		DB:        database,
	}

	allowedOrigins := []string{getenv("FRONTEND_URL", "http://localhost:3000")}
	router := gateway.New(handlers, allowedOrigins)

	// ── Servidor HTTP ─────────────────────────────────────────────────────────
	addr := ":" + getenv("PORT", "8080")
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", addr).Msg("server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("shutting down...")

	cancel() // encerra consumers

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("server stopped")
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}