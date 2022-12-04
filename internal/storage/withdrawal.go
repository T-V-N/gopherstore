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

func (w *Withdrawal) WithdrawBalance(ctx context.Context, uid, orderID string, amount, newCurrent, newWithdrawn float32) error {
	tx, err := w.Conn.Begin(ctx)

	if err != nil {
		return err
	}

	sqlUpdateUser := `	
	UPDATE USERS
    SET current_balance = $1, withdrawn = $2
	WHERE uid = $3;
	`

	sqlInsertWd := `
	INSERT INTO WITHDRAWALS (id, sum, uid) 
	VALUES ($1, $2, $3);
	`

	_, err = tx.Exec(ctx, sqlUpdateUser, newCurrent, newWithdrawn, uid)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, sqlInsertWd, orderID, amount, uid)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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
