package repositories

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	Key        string        `json:"key"`
	Capacity   int64         `json:"capacity"`
	RefillRate time.Duration `json:"refill_rate"`
	Unlimited  bool          `json:"unlimited"`
	CreatedAt  time.Time     `json:"created_at"`
}

type DBInterface interface {
	AddClient(ctx context.Context, client Client) error
	GetClient(ctx context.Context, key string) (Client, error)
	ListClients(ctx context.Context) ([]Client, error)
	DeleteClient(ctx context.Context, key string) error
	UpdateClient(ctx context.Context, client Client) error
}

type DB struct {
	Log  *slog.Logger
	Conn *pgxpool.Pool
}

var _ DBInterface = (*DB)(nil)

func New(log *slog.Logger, address string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Error("failed to ping database", "error", err)
		return nil, err
	}

	log.Info("successfully connected to database", "address", address)

	return &DB{
		Log:  log,
		Conn: pool,
	}, nil
}

func (db *DB) AddClient(ctx context.Context, client Client) error {
	db.Log.Debug("Started adding client to DB")

	query := `
        INSERT INTO clients (key, capacity, refill_rate, unlimited, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := db.Conn.Exec(ctx, query,
		client.Key,
		client.Capacity,
		client.RefillRate,
		client.Unlimited,
		client.CreatedAt,
	)

	if err != nil {
		db.Log.Error("Failed to add client", "error", err)
		return err
	}

	db.Log.Debug("Ended adding client to DB")
	return nil
}

func (db *DB) UpdateClient(ctx context.Context, client Client) error {
	db.Log.Debug("Started updating client in DB", "key", client.Key)

	query := `
        UPDATE clients
        SET 
            capacity = $1,
            refill_rate = $2,
            unlimited = $3
        WHERE key = $4
        RETURNING key, capacity, refill_rate, unlimited, created_at
    `

	var updated Client
	err := db.Conn.QueryRow(ctx, query,
		client.Capacity,
		client.RefillRate,
		client.Unlimited,
		client.Key,
	).Scan(
		&updated.Key,
		&updated.Capacity,
		&updated.RefillRate,
		&updated.Unlimited,
		&updated.CreatedAt,
	)

	if err != nil {
		db.Log.Error("Failed to update client", "error", err)
		return err
	}

	db.Log.Debug("Ended updating client in DB")
	return nil
}

func (db *DB) GetClient(ctx context.Context, key string) (Client, error) {
	db.Log.Debug("Started getting client from DB", "key", key)

	var client Client

	query := `
        SELECT key, capacity, refill_rate, unlimited, created_at
        FROM clients
        WHERE key = $1
    `

	err := db.Conn.QueryRow(ctx, query, key).Scan(
		&client.Key,
		&client.Capacity,
		&client.RefillRate,
		&client.Unlimited,
		&client.CreatedAt,
	)

	if err != nil {
		db.Log.Error("Failed to get client", "error", err)
		return Client{}, err
	}

	db.Log.Debug("Ended getting client from DB")
	return client, nil
}

func (db *DB) ListClients(ctx context.Context) ([]Client, error) {
	db.Log.Debug("Started listing clients from DB")

	var clients []Client

	query := `
        SELECT key, capacity, refill_rate, unlimited, created_at
        FROM clients
        ORDER BY created_at DESC
    `

	rows, err := db.Conn.Query(ctx, query)
	if err != nil {
		db.Log.Error("Failed to list clients", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var client Client
		err := rows.Scan(
			&client.Key,
			&client.Capacity,
			&client.RefillRate,
			&client.Unlimited,
			&client.CreatedAt,
		)
		if err != nil {
			db.Log.Error("Failed to scan client row", "error", err)
			return nil, err
		}
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		db.Log.Error("Error while iterating over client rows", "error", err)
		return nil, err
	}

	db.Log.Debug("Ended listing clients from DB")
	return clients, nil
}

func (db *DB) DeleteClient(ctx context.Context, key string) error {
	db.Log.Debug("Started deleting client from DB", "key", key)

	query := `
        DELETE FROM clients
        WHERE key = $1
    `

	result, err := db.Conn.Exec(ctx, query, key)
	if err != nil {
		db.Log.Error("Failed to delete client", "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		db.Log.Warn("No client found with the given key", "key", key)
		return errors.New("client not found")
	}

	db.Log.Debug("Ended deleting client from DB")
	return nil
}
