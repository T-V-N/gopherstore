package storage

// import (
// 	"context"
// 	"log"

// 	"github.com/T-V-N/gourlshortener/internal/config"
// 	"github.com/jackc/pgx/v5"
// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// type DBStorage struct {
// 	conn *pgxpool.Pool
// 	cfg  config.Config
// }

// func InitDBStorage(cfg *config.Config) (*DBStorage, error) {
// 	conn, err := pgxpool.New(context.Background(), cfg.DatabaseURI)
// 	if err != nil {
// 		log.Printf("Unable to connect to database: %v\n", err.Error())
// 		return nil, err
// 	}

// 	_, err = conn.Exec(context.Background(), `
// 	BEGIN;

// 	CREATE TABLE IF NOT EXISTS
// 	USER
// 	(
// 		uid serial primary key,
// 		login varchar,
// 		password_hash varchar,
// 		created_at timestamp default current_timestamp
// 	);

// 	COMMIT;
// 	`)

// 	if err != nil {
// 		log.Printf("Unable to create db: %v\n", err.Error())
// 		return nil, err
// 	}

// 	return &DBStorage{conn, *cfg}, nil
// }

// func (db *DBStorage) SaveURL(ctx context.Context, url, uid, hash string) error {
// 	sqlStatement := `
// 	INSERT INTO urls (user_uid, url_hash, original_url)
// 	VALUES ($1, $2, $3)`

// 	_, err := db.conn.Exec(ctx, sqlStatement, uid, hash, url)

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (db *DBStorage) GetURL(ctx context.Context, hash string) (URL, error) {
// 	row := db.conn.QueryRow(ctx, "Select * from urls where url_hash = $1", hash)

// 	u := URL{}
// 	err := row.Scan(&u.UID, &u.ShortURL, &u.URL, &u.IsDeleted)

// 	if err != nil {
// 		return URL{}, err
// 	}

// 	return u, nil
// }

// func (db *DBStorage) GetUrlsByUID(ctx context.Context, uid string) ([]URL, error) {
// 	urls := make([]URL, 0)

// 	rows, err := db.conn.Query(ctx, "SELECT hash, original_url from urls where user_uid = $1", uid)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer rows.Close()

// 	for rows.Next() {
// 		var u URL
// 		err = rows.Scan(&u.ShortURL, &u.URL)

// 		if err != nil {
// 			return nil, err
// 		}

// 		urls = append(urls, u)
// 	}

// 	err = rows.Err()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return urls, nil
// }

// func (db *DBStorage) IsAlive(ctx context.Context) (bool, error) {
// 	err := db.conn.Ping(ctx)

// 	if err != nil {
// 		return false, err
// 	}

// 	return true, nil
// }

// func (db *DBStorage) BatchSaveURL(ctx context.Context, urls []URL) error {
// 	tx, err := db.conn.Begin(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	defer tx.Rollback(ctx)

// 	stmt, err := tx.Prepare(ctx, "batch insert", "INSERT INTO urls(user_uid, url_hash, original_url) VALUES($1,$2,$3)")
// 	if err != nil {
// 		return err
// 	}

// 	for _, u := range urls {
// 		if _, err = tx.Exec(ctx, stmt.Name, u.UID, u.ShortURL, u.URL); err != nil {
// 			return err
// 		}
// 	}

// 	return tx.Commit(ctx)
// }

// func (db *DBStorage) KillConn() error {
// 	db.conn.Close()
// 	return nil
// }

// func (db *DBStorage) DeleteURLs(ctx context.Context, entries []DeletionEntry) error {
// 	b := pgx.Batch{}
// 	for _, e := range entries {
// 		b.Queue("UPDATE urls set is_deleted = true WHERE user_uid = $1 and url_hash = $2", e.UID, e.Hash)
// 	}

// 	br := db.conn.SendBatch(context.Background(), &b)
// 	err := br.Close()

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
