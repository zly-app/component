package sqlx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type errClient struct {
	err error
}

func (e errClient) GetDB() *sqlx.DB { return nil }

func (e errClient) Unsafe() Client { return e }

func (e errClient) Find(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return e.err
}

func (e errClient) FindOne(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return e.err
}

func (e errClient) FindColumn(ctx context.Context, dest []interface{}, query string, args ...interface{}) error {
	return e.err
}

func (e errClient) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, e.err
}

func (e errClient) Query(ctx context.Context, next NextFunc, query string, args ...interface{}) error {
	return e.err
}

func (e errClient) Transaction(ctx context.Context, fn TxFunc, opts ...TxOption) error {
	return e.err
}

func (e errClient) TransactionX(ctx context.Context, fn TxxFunc, opts ...TxOption) error {
	return e.err
}

func newErrClient(err error) Client {
	return errClient{err}
}
