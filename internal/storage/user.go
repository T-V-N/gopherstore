package storage

import (
	"context"
	"errors"
	"sync"

	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	Conn     *pgxpool.Pool
	lockedMu map[string]*sync.Mutex
}

func InitUser(conn *pgxpool.Pool) (*User, error) {
	return &User{conn, map[string]*sync.Mutex{}}, nil
}

func (user *User) CreateUser(ctx context.Context, creds sharedTypes.Credentials) (string, error) {
	sqlStatement := `
	INSERT INTO users (login, password_hash, current_balance, withdrawn)
	VALUES ($1, $2, 0, 0)
	RETURNING uid;`

	var id string
	err := user.Conn.QueryRow(ctx, sqlStatement, creds.Login, creds.Password).Scan(&id)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return "", utils.ErrDuplicate
	}

	return id, nil
}

func (user *User) UpdateUser(ctx context.Context, orderID, uid string, accrual float32) error {
	if user.lockedMu[uid] == nil {
		user.lockedMu[uid] = &sync.Mutex{}
	}

	user.lockedMu[uid].Lock()

	defer func() {
		user.lockedMu[uid].Unlock()
		user.lockedMu[uid] = nil
	}()

	updateBalanceSQL := `
	UPDATE USERS SET current_balance = current_balance + $1
	WHERE uid = (select uid from orders WHERE id = $2)
	`

	_, err := user.Conn.Exec(ctx, updateBalanceSQL, accrual, orderID)

	if err != nil {
		return err
	}

	return nil
}

func (user *User) GetUser(ctx context.Context, creds sharedTypes.Credentials) (sharedTypes.User, error) {
	sqlStatement := `
	SELECT uid, login, password_hash, current_balance, withdrawn FROM USERS
	WHERE login = $1
	`

	var u sharedTypes.User
	err := user.Conn.QueryRow(ctx, sqlStatement, creds.Login).Scan(&u.UID, &u.Login, &u.PasswordHash, &u.CurrentBalance, &u.Withdrawn)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (user *User) GetBalance(ctx context.Context, uid string) (sharedTypes.Balance, error) {
	sqlStatement := `
	SELECT uid, login, password_hash, current_balance, withdrawn FROM USERS
	WHERE uid = $1
	`

	var u sharedTypes.User
	err := user.Conn.QueryRow(ctx, sqlStatement, uid).Scan(&u.UID, &u.Login, &u.PasswordHash, &u.CurrentBalance, &u.Withdrawn)

	if err != nil {
		return sharedTypes.Balance{}, err
	}

	return sharedTypes.Balance{Current: u.CurrentBalance, Withdrawn: u.Withdrawn}, nil
}

func (user *User) GetBalanceAndLock(ctx context.Context, uid string) (sharedTypes.Balance, error) {
	if user.lockedMu[uid] == nil {
		user.lockedMu[uid] = &sync.Mutex{}
	}

	user.lockedMu[uid].Lock()

	return user.GetBalance(ctx, uid)
}

func (user *User) WithdrawBalance(ctx context.Context, uid, orderID string, amount, newCurrent, newWithdrawn float32, withdrawal sharedTypes.WithdrawalStorage) error {
	defer user.lockedMu[uid].Unlock()

	sqlUpdateUser := `	
	UPDATE USERS
    SET current_balance = $1, withdrawn = $2
	WHERE uid = $3;
	`

	_, err := user.Conn.Exec(ctx, sqlUpdateUser, newCurrent, newWithdrawn, uid)

	if err != nil {
		return err
	}

	return withdrawal.CreateWithdrawal(ctx, uid, amount, orderID)
}
