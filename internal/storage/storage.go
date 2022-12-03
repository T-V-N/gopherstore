package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	conn *pgxpool.Pool
	cfg  config.Config
}

func InitStorage(cfg config.Config) (*Storage, error) {
	conn, err := pgxpool.New(context.Background(), cfg.DatabaseURI)

	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err.Error())
		return nil, err
	}

	m, err := migrate.New(
		"file://migrations/",
		cfg.DatabaseURI)

	if err != nil {
		return nil, err
	}

	defer m.Close()

	err = m.Up()

	if err != nil {
		if err != migrate.ErrNoChange {
			return nil, err
		}
	}

	return &Storage{conn, cfg}, nil
}

func (st *Storage) CreateUser(ctx context.Context, creds sharedTypes.Credentials) (string, error) {
	sqlStatement := `
	INSERT INTO users (login, password_hash, current_balance, withdrawn)
	VALUES ($1, $2, 0, 0)
	RETURNING uid;`

	var id string
	err := st.conn.QueryRow(ctx, sqlStatement, creds.Login, creds.Password).Scan(&id)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return "", utils.ErrDuplicate
	}

	return id, nil
}

func (st *Storage) GetUser(ctx context.Context, creds sharedTypes.Credentials) (sharedTypes.User, error) {
	sqlStatement := `
	SELECT uid, login, password_hash, current_balance, withdrawn FROM USERS
	WHERE login = $1
	`

	var u sharedTypes.User
	err := st.conn.QueryRow(ctx, sqlStatement, creds.Login).Scan(&u.UID, &u.Login, &u.PasswordHash, &u.CurrentBalance, &u.Withdrawn)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (st *Storage) CreateOrder(ctx context.Context, orderID, uid string) error {
	sqlCheckExists := `
	SELECT ID, UID FROM orders WHERE ID = $1  
	`

	var id string

	var ownerID int

	err := st.conn.QueryRow(ctx, sqlCheckExists, orderID).Scan(&id, &ownerID)
	if err == nil {
		if strconv.Itoa(ownerID) == uid {
			return utils.ErrAlreadyCreated
		} else {
			return utils.ErrDuplicate
		}
	}

	if err != pgx.ErrNoRows {
		return err
	}

	sqlCreate := `
	INSERT INTO orders (uid, id, status, accrual)
	VALUES ($1, $2, 'NEW', 0)`

	_, err = st.conn.Exec(ctx, sqlCreate, uid, orderID)
	if err != nil {
		return err
	}

	return nil
}

func (st *Storage) ListOrders(ctx context.Context, uid string) ([]sharedTypes.Order, error) {
	sqlStatement := `
	SELECT ID, status, accrual, uploaded_at::timestamptz FROM orders WHERE UID = $1 ORDER BY uploaded_at
	`

	rows, err := st.conn.Query(ctx, sqlStatement, uid)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	defer rows.Close()

	orders := []sharedTypes.Order{}

	for rows.Next() {
		entry := sharedTypes.Order{}
		err = rows.Scan(&entry.Number, &entry.Status, &entry.Accrual, &entry.UploadedAt)

		if err != nil {
			return nil, err
		}

		orders = append(orders, entry)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (st *Storage) GetBalance(ctx context.Context, uid string) (sharedTypes.Balance, error) {
	sqlStatement := `
	SELECT uid, login, password_hash, current_balance, withdrawn FROM USERS
	WHERE uid = $1
	`

	var u sharedTypes.User
	err := st.conn.QueryRow(ctx, sqlStatement, uid).Scan(&u.UID, &u.Login, &u.PasswordHash, &u.CurrentBalance, &u.Withdrawn)

	if err != nil {
		return sharedTypes.Balance{}, err
	}

	return sharedTypes.Balance{Current: u.CurrentBalance, Withdrawn: u.Withdrawn}, nil
}

func (st *Storage) WithdrawBalance(ctx context.Context, uid, orderID string, amount, newCurrent, newWithdrawn float32) error {
	tx, err := st.conn.Begin(ctx)

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

func (st *Storage) ListWithdrawals(ctx context.Context, uid string) ([]sharedTypes.Withdrawal, error) {
	sqlStatement := `
	SELECT id, sum, processed_at::timestamptz FROM withdrawals WHERE UID = $1 ORDER BY processed_at
	`

	rows, err := st.conn.Query(ctx, sqlStatement, uid)
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

func (st *Storage) GetUnproccessedOrders(ctx context.Context) ([]sharedTypes.Order, error) {
	sqlStatement := `
	SELECT id, status, accrual, uploaded_at::timestamptz FROM orders WHERE status = 'NEW' or status = 'PROCESSING'
	`

	rows, err := st.conn.Query(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	orders := []sharedTypes.Order{}

	for rows.Next() {
		entry := sharedTypes.Order{}
		err = rows.Scan(&entry.Number, &entry.Status, &entry.Accrual, &entry.UploadedAt)

		if err != nil {
			return nil, err
		}

		orders = append(orders, entry)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (st *Storage) UpdateOrder(ctx context.Context, orderID, status string, accrual float32) error {
	updateOrderSQL := `
	UPDATE orders SET status = $1, accrual = $2  WHERE id = $3
	`
	_, err := st.conn.Exec(ctx, updateOrderSQL, status, accrual, orderID)

	if err != nil {
		return err
	}

	if accrual != 0 {
		updateBalanceSQL := `
		UPDATE USERS SET current_balance = current_balance + $1
		WHERE uid = (select uid from orders WHERE id = $2)
		`

		_, err := st.conn.Exec(ctx, updateBalanceSQL, accrual, orderID)

		if err != nil {
			return err
		}
	}

	return nil
}
