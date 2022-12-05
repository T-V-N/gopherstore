package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	a, err := app.InitApp(st.Conn, cfg, sugar)
	if err != nil {
		sugar.Fatalw("Unable to init application",
			"Error", err,
		)
	}

	hn := handler.InitHandler(a, cfg, sugar)
	authMw := middleware.InitAuth(cfg)

	router := chi.NewRouter()
	router.Use(chiMw.Compress(cfg.CompressLevel))
	router.Use(middleware.GzipHandle)

	router.Route("/api/user", func(userRouter chi.Router) {
		userRouter.Post("/login", hn.HandleLogin)
		userRouter.Post("/register", hn.HandleRegister)
		userRouter.Group(func(r chi.Router) {
			r.Use(authMw)
			r.Post("/orders", hn.HandleCreateOrder)
			r.Get("/orders", hn.HandleListOrder)
			r.Get("/balance", hn.HandleGetBalance)
			r.Post("/balance/withdraw", hn.HandleBalanceWithdraw)
			r.Get("/withdrawals", hn.HandleListWithdrawals)
		})
	})

	go service.InitUpdater(*cfg, st.Conn, 1, sugar, ctx)

	server := http.Server{
		Handler: router,
		Addr:    cfg.RunAddress,
	}

	err = server.Shutdown(ctx)
	if err != nil {
		sugar.Fatalw("Unable to shutdown server",
			"Error", err,
		)
	}

	sugar.Infow("Starting server",
		"Port", cfg.RunAddress,
	)

	err = server.ListenAndServe()

	if err != nil {
		sugar.Info("Server stopped",
			"MSG", err,
		)
	}
}
