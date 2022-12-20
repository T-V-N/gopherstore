package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/T-V-N/gopherstore/internal/app"
	"github.com/T-V-N/gopherstore/internal/config"
	"github.com/T-V-N/gopherstore/internal/handler"
	"github.com/T-V-N/gopherstore/internal/middleware"
	"github.com/T-V-N/gopherstore/internal/router"
	"github.com/T-V-N/gopherstore/internal/utils"

	service "github.com/T-V-N/gopherstore/internal/services"
	"github.com/T-V-N/gopherstore/internal/storage"
	"go.uber.org/zap"
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

	var userLocks sync.Map

	withdrawalApp, err := app.InitWithdrawal(st.Conn, cfg, sugar)
	if err != nil {
		sugar.Fatalw("Unable to init application",
			"Error", err,
		)
	}

	userApp, err := app.InitUserApp(st.Conn, *withdrawalApp, cfg, sugar, &userLocks)
	if err != nil {
		sugar.Fatalw("Unable to init application",
			"Error", err,
		)
	}

	orderApp, err := app.InitOrderApp(st.Conn, cfg, sugar, utils.RegOrderHTTPInit(cfg.AccrualSystemAddress+"/api/orders"))
	if err != nil {
		sugar.Fatalw("Unable to init application",
			"Error", err,
		)
	}

	userHn := handler.InitUserHandler(userApp, cfg, sugar)
	orderHn := handler.InitOrderHandler(orderApp, cfg, sugar)
	withdrawalHn := handler.InitWithdrawalHandler(withdrawalApp, cfg, sugar)
	authMw := middleware.InitAuth(cfg)
	r := router.InitRouter(cfg, authMw, userHn, orderHn, withdrawalHn)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	server := http.Server{
		Handler: r,
		Addr:    cfg.RunAddress,
	}

	sugar.Infow("Starting server",
		"Port", cfg.RunAddress,
	)

	gr := &sync.WaitGroup{}
	gr.Add(1)

	go func() {
		service.InitUpdater(ctx, *cfg, st.Conn, cfg.WorkerLimit, sugar, userApp, orderApp)
		gr.Done()
	}()

	gr.Add(1)
	go func() {
		err = server.ListenAndServe()

		if err != nil {
			sugar.Error("Unable to run server")
		}
		gr.Done()
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, stopShutdownCtx := context.WithTimeout(context.Background(), time.Duration(cfg.ContextCancelTimeout)*time.Second)
	defer stopShutdownCtx()

	err = server.Shutdown(shutdownCtx)

	if err != nil {
		sugar.Errorw("Unable to shutdown server",
			"Error", err,
		)
	}

	gr.Wait()
	sugar.Info("Server stopped")
}
