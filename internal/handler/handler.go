package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/T-V-N/gopherstore/internal/app"
	"github.com/T-V-N/gopherstore/internal/auth"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
)

type Handler struct {
	app *app.App
}

func InitHandler(a *app.App) *Handler {
	return &Handler{a}
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	cp := r.Header.Get("Content-Type")
	if cp != "application/json" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
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

	token, err := auth.CreateToken(uid, h.app.Cfg)

	w.Header().Add("Authorization", fmt.Sprintf("Bearer %v", token))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	cp := r.Header.Get("Content-Type")
	if cp != "application/json" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	cred := sharedTypes.Credentials{}
	json.NewDecoder(r.Body).Decode(&cred)

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

func (h *Handler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	cp := r.Header.Get("Content-Type")
	if cp != "text/plain" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	uid, _ := r.Context().Value(sharedTypes.UIDKey{}).(string)
	err = h.app.CreateOrder(ctx, string(body), uid)

	if err != nil {
		switch {
		case errors.Is(err, utils.ErrAlreadyCreated):
			http.Error(w, err.Error(), http.StatusOK)
			return
		case errors.Is(err, utils.ErrDuplicate):
			http.Error(w, err.Error(), http.StatusConflict)
			return
		case errors.Is(err, utils.ErrWrongFormat):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) HandleListOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	uid, _ := r.Context().Value(sharedTypes.UIDKey{}).(string)

	list, err := h.app.ListOrders(ctx, uid)

	if err == utils.ErrNoData {
		http.Error(w, "No content", http.StatusNoContent)

		return
	}

	w.Header().Add("Content-Type", "application/json")

	if err != json.NewEncoder(w).Encode(list) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	uid, _ := r.Context().Value(sharedTypes.UIDKey{}).(string)

	balance, err := h.app.GetBalance(ctx, uid)

	if err != json.NewEncoder(w).Encode(balance) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleBalanceWithdraw(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
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

func (h *Handler) HandleListWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	uid, _ := r.Context().Value(sharedTypes.UIDKey{}).(string)

	withdrawalsList, err := h.app.GetListWithdrawals(ctx, uid)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrNoData):
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Add("Content-Type", "application/json")

	if err != json.NewEncoder(w).Encode(withdrawalsList) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
