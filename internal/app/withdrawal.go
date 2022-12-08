package app

import (
	"context"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/storage"
	"github.com/T-V-N/gopherstore/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type WithdrawalApp struct {
	Withdrawal sharedTypes.WithdrawalStorage
	Cfg        *config.Config
	logger     *zap.SugaredLogger
}

func InitWithdrawal(Conn *pgxpool.Pool, cfg *config.Config, logger *zap.SugaredLogger) (*WithdrawalApp, error) {
	withdrawal, err := storage.InitWithdrawal(Conn)

	if err != nil {
		return nil, err
	}

	return &WithdrawalApp{withdrawal, cfg, logger}, nil
}

func (app *WithdrawalApp) GetListWithdrawals(ctx context.Context, uid string) ([]sharedTypes.Withdrawal, error) {
	list, err := app.Withdrawal.ListWithdrawals(ctx, uid)

	if len(list) == 0 {
		return []sharedTypes.Withdrawal{}, utils.ErrNoData
	}

	return list, err
}
