package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
	"go.uber.org/zap"
)

type OrderAppInterface interface {
	CreateOrder(ctx context.Context, orderID string, uid string) error
	ListOrders(ctx context.Context, uid string) ([]sharedTypes.Order, error)
}

type OrderHandler struct {
	app    OrderAppInterface
	Cfg    *config.Config
	logger *zap.SugaredLogger
}

func InitOrderHandler(a OrderAppInterface, cfg *config.Config, logger *zap.SugaredLogger) *OrderHandler {
	return &OrderHandler{a, cfg, logger}
}

func (h *OrderHandler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	cp := r.Header.Get("Content-Type")
	if cp != "text/plain" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.Cfg.ContextCancelTimeout)*time.Second)
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

func (h *OrderHandler) HandleListOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.Cfg.ContextCancelTimeout)*time.Second)
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
