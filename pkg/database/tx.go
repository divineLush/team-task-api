package database

import (
	"context"
	"database/sql" // needed for sql.TxOptions
)

type TxManager interface {
	InTx(ctx context.Context, fn func(tx Querier) error) error
}

type sqlTxManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) TxManager {
	return &sqlTxManager{db: db}
}

func (m *sqlTxManager) InTx(ctx context.Context, fn func(tx Querier) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
