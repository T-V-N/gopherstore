package storage

import (
	"context"
	"fmt"
	"strconv"

	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Order struct {
	Conn *pgxpool.Pool
}

func InitOrder(conn *pgxpool.Pool) (*Order, error) {
	return &Order{conn}, nil
}

func (order *Order) CreateOrder(ctx context.Context, orderID, uid string) error {
	sqlCheckExists := `
	SELECT ID, UID FROM orders WHERE ID = $1  
	`

	var id string

	var ownerID int

	err := order.Conn.QueryRow(ctx, sqlCheckExists, orderID).Scan(&id, &ownerID)
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

	_, err = order.Conn.Exec(ctx, sqlCreate, uid, orderID)
	if err != nil {
		return err
	}

	return nil
}

func (order *Order) ListOrders(ctx context.Context, uid string) ([]sharedTypes.Order, error) {
	sqlStatement := `
	SELECT ID, status, accrual, uploaded_at::timestamptz FROM orders WHERE UID = $1 ORDER BY uploaded_at
	`

	rows, err := order.Conn.Query(ctx, sqlStatement, uid)
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

func (order *Order) GetUnproccessedOrders(ctx context.Context) ([]sharedTypes.Order, error) {
	sqlStatement := `
	SELECT id, status, accrual, uploaded_at::timestamptz FROM orders WHERE status = 'NEW' or status = 'PROCESSING'
	`

	rows, err := order.Conn.Query(ctx, sqlStatement)
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

func (order *Order) UpdateOrder(ctx context.Context, orderID, status string, accrual float32) error {
	updateOrderSQL := `
	UPDATE orders SET status = $1, accrual = $2  WHERE id = $3
	`
	_, err := order.Conn.Exec(ctx, updateOrderSQL, status, accrual, orderID)

	if err != nil {
		return err
	}

	if accrual != 0 {
		updateBalanceSQL := `
		UPDATE USERS SET current_balance = current_balance + $1
		WHERE uid = (select uid from orders WHERE id = $2)
		`

		_, err := order.Conn.Exec(ctx, updateBalanceSQL, accrual, orderID)

		if err != nil {
			return err
		}
	}

	return nil
}