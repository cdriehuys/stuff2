package models_test

import (
	"context"

	"github.com/cdriehuys/stuff2/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type MockDB struct {
	txFactory func() models.Transaction

	beginError error
}

func (db *MockDB) Begin(context.Context) (models.Transaction, error) {
	return db.txFactory(), db.beginError
}

func (db *MockDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (db *MockDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (db *MockDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

type MockTX struct {
	commitError   error
	rollbackError error

	committed  bool
	rolledBack bool

	MockDB
}

func (tx *MockTX) Commit(context.Context) error {
	if tx.commitError != nil {
		return tx.commitError
	}

	if tx.rolledBack {
		return pgx.ErrTxClosed
	}

	tx.committed = true

	return nil
}

func (tx *MockTX) Conn() *pgx.Conn {
	return nil
}

func (tx *MockTX) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (tx *MockTX) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (tx *MockTX) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (tx *MockTX) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}

func (tx *MockTX) Rollback(context.Context) error {
	if tx.rollbackError != nil {
		return tx.rollbackError
	}

	if tx.committed {
		return pgx.ErrTxClosed
	}

	tx.rolledBack = true

	return nil
}
