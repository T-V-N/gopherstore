package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/T-V-N/gopherstore/internal/auth"
	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
	"go.uber.org/zap"
)

type UserHandler struct {
	app    sharedTypes.UserApper
	Cfg    *config.Config
	logger *zap.SugaredLogger
}

func InitUserHandler(a sharedTypes.UserApper, cfg *config.Config, logger *zap.SugaredLogger) *UserHandler {
	return &UserHandler{a, cfg, logger}
}

func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	cp := r.Header.Get("Content-Type")
	if cp != "application/json" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.Cfg.ContextCancelTimeout)*time.Second)
	defer cancel()

	cred := sharedTypes.Credentials{}
	err := json.NewDecoder(r.Body).Decode(&cred)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uid, err := h.app.Register(ctx, cred)

	if err != nil {
		switch {
		case errors.Is(err, utils.ErrBadCredentials):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case errors.Is(err, utils.ErrDuplicate):
			http.Error(w, err.Error(), http.StatusConflict)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	token, err := auth.CreateToken(uid, h.Cfg)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Authorization", fmt.Sprintf("Bearer %v", token))
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	cp := r.Header.Get("Content-Type")
	if cp != "application/json" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.Cfg.ContextCancelTimeout)*time.Second)
	defer cancel()

	cred := sharedTypes.Credentials{}

	err := json.NewDecoder(r.Body).Decode(&cred)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.app.Login(ctx, cred)

	if err != nil {
		switch {
		case errors.Is(err, utils.ErrBadCredentials):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case errors.Is(err, utils.ErrNotAuthorized):
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	w.Header().Add("Authorization", fmt.Sprintf("Bearer %v", token))
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.Cfg.ContextCancelTimeout)*time.Second)
	defer cancel()

	uid, _ := r.Context().Value(sharedTypes.UIDKey{}).(string)

	balance, err := h.app.GetBalance(ctx, uid)

	w.Header().Add("Content-Type", "application/json")

	if err != json.NewEncoder(w).Encode(balance) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *UserHandler) HandleBalanceWithdraw(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.Cfg.ContextCancelTimeout)*time.Second)
	defer cancel()

	uid, _ := r.Context().Value(sharedTypes.UIDKey{}).(string)

	withdrawRequest := sharedTypes.WtihdrawRequest{}
	err := json.NewDecoder(r.Body).Decode(&withdrawRequest)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.app.WithdrawBalance(ctx, uid, withdrawRequest.OrderID, withdrawRequest.Sum)

	if err != nil {
		switch {
		case errors.Is(err, utils.ErrPaymentError):
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		case errors.Is(err, utils.ErrWrongFormat):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
