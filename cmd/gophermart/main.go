package main

import (
	"log"
	"net/http"

	"github.com/T-V-N/gopherstore/internal/app"
	"github.com/T-V-N/gopherstore/internal/config"
	"github.com/T-V-N/gopherstore/internal/handler"
	"github.com/T-V-N/gopherstore/internal/middleware"
	service "github.com/T-V-N/gopherstore/internal/services"
	"github.com/T-V-N/gopherstore/internal/storage"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Panic("Unable to load config. Check if it was passed")
	}

	st, err := storage.InitStorage(*cfg)
	if err != nil {
		log.Fatal(err)
	}

	app := app.InitApp(st, cfg)
	hn := handler.InitHandler(app, cfg)
	authMw := middleware.InitAuth(cfg)

	router := chi.NewRouter()
	router.Use(chiMw.Compress(5))
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

	go service.InitUpdater(*cfg, *st, 1)

	log.Panic(http.ListenAndServe(cfg.RunAddress, router))
}
