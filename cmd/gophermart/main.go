package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/T-V-N/gopherstore/internal/app"
	"github.com/T-V-N/gopherstore/internal/config"
	"github.com/T-V-N/gopherstore/internal/handler"
	"github.com/T-V-N/gopherstore/internal/middleware"
	service "github.com/T-V-N/gopherstore/internal/services"
	"github.com/T-V-N/gopherstore/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sugar := logger.Sugar()

	cfg, err := config.Init()
	if err != nil {
		sugar.Fatalw("Unable to load config",
			"Error", err,
		)
	}

	st, err := storage.InitStorage(*cfg)
	if err != nil {
		sugar.Fatalw("Unable to migrate db and init pgx connection",
			"DB URL", cfg.DatabaseURI,
			"Migrations path", cfg.MigrationsPath,
			"Error", err,
		)
	}

	defer st.Conn.Close()

	withdrawalApp, err := app.InitWithdrawal(st.Conn, cfg, sugar)
	if err != nil {
		sugar.Fatalw("Unable to init application",
			"Error", err,
		)
	}

	userApp, err := app.InitUserApp(st.Conn, *withdrawalApp, cfg, sugar)
	if err != nil {
		sugar.Fatalw("Unable to init application",
			"Error", err,
		)
	}

	orderApp, err := app.InitOrderApp(st.Conn, cfg, sugar)
	if err != nil {
		sugar.Fatalw("Unable to init application",
			"Error", err,
		)
	}

	userHn := handler.InitUserHandler(userApp, cfg, sugar)
	orderHn := handler.InitOrderHandler(orderApp, cfg, sugar)
	withdrawalHn := handler.InitWithdrawalHandler(withdrawalApp, cfg, sugar)
	authMw := middleware.InitAuth(cfg)
	router := initRouter(cfg, authMw, userHn, orderHn, withdrawalHn)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	server := http.Server{
		Handler: router,
		Addr:    cfg.RunAddress,
	}

	sugar.Infow("Starting server",
		"Port", cfg.RunAddress,
	)

	gr, grCtx := errgroup.WithContext(ctx)

	gr.Go(func() error {
		return service.InitUpdater(grCtx, *cfg, st.Conn, cfg.WorkerLimit, sugar)
	})
	gr.Go(server.ListenAndServe)

	if err := gr.Wait(); err != nil {
		sugar.Errorw("Server or updater crashed", "Error", err)
		return
	}
	stop()

	shutdownCtx, stopShutdownCtx := context.WithTimeout(context.Background(), time.Duration(cfg.ContextCancelTimeout)*time.Second)
	defer stopShutdownCtx()

	err = server.Shutdown(shutdownCtx)

	if err != nil {
		sugar.Errorw("Unable to shutdown server",
			"Error", err,
		)
	}
}
