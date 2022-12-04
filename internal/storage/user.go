package storage

import (
	"context"
	"errors"

	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	Conn *pgxpool.Pool
}

func InitUser(conn *pgxpool.Pool) (*User, error) {
	return &User{conn}, nil
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
