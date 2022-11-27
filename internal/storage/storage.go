package storage

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
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
		id bigint primary key,
		status varchar,
		accural real,
		uploaded_at timestamp default current_timestamp
	);

	CREATE TABLE IF NOT EXISTS 
	WITHDRAWALS
	(
		id bigint primary key,
		order_id integer references orders(id),
		sum real,
		processed_at timestamp default current_timestamp
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
	INSERT INTO orders (uid, id, status, accural)
	VALUES ($1, $2, 'NEW', 0)`

	_, err = st.conn.Exec(ctx, sqlCreate, uid, orderID)
	if err != nil {
		return err
	}

	return nil
}

func (st *Storage) ListOrders(ctx context.Context, uid string) ([]sharedTypes.Order, error) {
	sqlStatement := `
	SELECT ID, status, accural, uploaded_at::timestamptz FROM orders WHERE UID = $1 ORDER BY uploaded_at
	`

	rows, err := st.conn.Query(ctx, sqlStatement, uid)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	orders := []sharedTypes.Order{}

	for rows.Next() {
		entry := sharedTypes.Order{}
		err = rows.Scan(&entry.Number, &entry.Status, &entry.Accural, &entry.UploadedAt)

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
	sqlStatement := `
	BEGIN;
	
	INSERT INTO USERS (curent_balance, withdrawn)
	VALUES ($1, $2)
	WHERE uid = $3;

	INSERT INTO WITHDRAWALS (order_id, sum) 
	VALUES ($4, $5)

	COMMIT;
	`

	_, err := st.conn.Exec(ctx, sqlStatement, newCurrent, newWithdrawn, uid, orderID, amount)

	if err != nil {
		return err
	}

	return nil
}

func (st *Storage) ListWithdrawals(ctx context.Context, uid string) ([]sharedTypes.Withdrawal, error) {
	sqlStatement := `
	SELECT order_id, sum, processed_at::timestamptz FROM orders WHERE UID = $1 ORDER BY processed_at
	`

	rows, err := st.conn.Query(ctx, sqlStatement, uid)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	withdrawals := []sharedTypes.Withdrawal{}

	for rows.Next() {
		entry := sharedTypes.Withdrawal{}
		err = rows.Scan(&entry.OrderID, &entry.Sum, &entry.ProcessedAt)

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
