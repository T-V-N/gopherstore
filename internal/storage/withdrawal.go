package storage

import (
	"context"

	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Withdrawal struct {
	Conn *pgxpool.Pool
}

func InitWithdrawal(conn *pgxpool.Pool) (*Withdrawal, error) {
	return &Withdrawal{conn}, nil
}

func (w *Withdrawal) ListWithdrawals(ctx context.Context, uid string) ([]sharedTypes.Withdrawal, error) {
	sqlStatement := `
	SELECT id, sum, processed_at::timestamptz FROM withdrawals WHERE UID = $1 ORDER BY processed_at
	`

	rows, err := w.Conn.Query(ctx, sqlStatement, uid)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	withdrawals := []sharedTypes.Withdrawal{}

	for rows.Next() {
		entry := sharedTypes.Withdrawal{}
		err = rows.Scan(&entry.ID, &entry.Sum, &entry.ProcessedAt)

		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, entry)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (w *Withdrawal) CreateWithdrawal(ctx context.Context, uid string, amount float32, orderID string) error {
	sqlInsertWd := `
	INSERT INTO WITHDRAWALS (id, sum, uid) 
	VALUES ($1, $2, $3);
	`

	_, err := w.Conn.Exec(ctx, sqlInsertWd, orderID, amount, uid)

	if err != nil {
		return err
	}

	return nil
}
