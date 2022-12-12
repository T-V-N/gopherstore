package main

import (
	"net/http"

	"github.com/T-V-N/gopherstore/internal/config"
	"github.com/T-V-N/gopherstore/internal/handler"
	"github.com/T-V-N/gopherstore/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
)

func initRouter(cfg *config.Config,
	authMw func(next http.Handler) http.Handler,
	userHn *handler.UserHandler,
	orderHn *handler.OrderHandler,
	withdrawalHn *handler.WithdrawalHandler) chi.Router {
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

	return router
}
