package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
	"go.uber.org/zap"
)

type WithdrawalAppInterface interface {
	GetListWithdrawals(ctx context.Context, uid string) ([]sharedTypes.Withdrawal, error)
}

type WithdrawalHandler struct {
	app    WithdrawalAppInterface
	Cfg    *config.Config
	logger *zap.SugaredLogger
}

func InitWithdrawalHandler(a WithdrawalAppInterface, cfg *config.Config, logger *zap.SugaredLogger) *WithdrawalHandler {
	return &WithdrawalHandler{a, cfg, logger}
}

func (h *WithdrawalHandler) HandleListWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.Cfg.ContextCancelTimeout)*time.Second)
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
	err = json.NewEncoder(w).Encode(withdrawalsList)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
