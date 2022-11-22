package storage

import (
	"context"
	"errors"
	"log"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
	"github.com/jackc/pgerrcode"
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

	_, err = conn.Exec(context.Background(), `
	BEGIN;
	CREATE TABLE IF NOT EXISTS 
	USERS
	(
		uid serial primary key,
		login varchar unique,
		password_hash varchar, 
		current_balance real,
		withdrawn real,
		created_at timestamp default current_timestamp
	);

	CREATE TABLE IF NOT EXISTS 
	ORDERS
	(
		uid integer references users(uid),
		id integer primary key,
		status varchar,
		accural real,
		updated_at timestamp default current_timestamp
	);
	COMMIT;
	`)

	if err != nil {
		log.Printf("Unable to create db: %v\n", err.Error())
		return nil, err
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
	return nil
}

func (st *Storage) ListOrders(ctx context.Context, uid string) ([]sharedTypes.Order, error) {
	return []sharedTypes.Order{}, nil
}

func (st *Storage) GetBalance(ctx context.Context, uid string) (sharedTypes.Balance, error) {
	return sharedTypes.Balance{}, nil
}

func (st *Storage) WithdrawBalance(ctx context.Context, uid, orderID string, amount float32) error {
	return nil
}

func (st *Storage) GetListWithdrawals(ctx context.Context, uid string) ([]sharedTypes.Withdrawal, error) {
	return []sharedTypes.Withdrawal{}, nil
}
