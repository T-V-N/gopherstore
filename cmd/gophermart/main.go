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

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
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

	router := chi.NewRouter()
	router.Use(chiMw.Compress(cfg.CompressLevel))
	router.Use(middleware.GzipHandle)

	router.Route("/api/user", func(userRouter chi.Router) {
		userRouter.Post("/login", userHn.HandleLogin)
		userRouter.Post("/register", userHn.HandleRegister)
		userRouter.Group(func(r chi.Router) {
			r.Use(authMw)
			r.Post("/orders", orderHn.HandleCreateOrder)
			r.Get("/orders", orderHn.HandleListOrder)
			r.Get("/balance", userHn.HandleGetBalance)
			r.Post("/balance/withdraw", userHn.HandleBalanceWithdraw)
			r.Get("/withdrawals", withdrawalHn.HandleListWithdrawals)
		})
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	go service.InitUpdater(ctx, *cfg, st.Conn, cfg.WorkerLimit, sugar)

	server := http.Server{
		Handler: router,
		Addr:    cfg.RunAddress,
	}

	sugar.Infow("Starting server",
		"Port", cfg.RunAddress,
	)

	go server.ListenAndServe()

	<-ctx.Done()
	stop()

	shutdownCtx, stopShutdownCtx := context.WithTimeout(context.Background(), time.Duration(cfg.ContextCancelTimeout)*time.Second)
	defer stopShutdownCtx()

	err = server.Shutdown(shutdownCtx)

	if err != nil {
		sugar.Fatalw("Unable to shutdown server",
			"Error", err,
		)
	}
}
