package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Client interface {
	GetDB() *sqlx.DB
	// 不安全模式, 在安全模式下, 如果 select 语句的字段在 scan 结构体中未定义会报错
	Unsafe() Client

	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, next NextFunc, query string, args ...interface{}) error
	QueryRow(ctx context.Context, dest []interface{}, query string, args ...interface{}) error
	Transaction(ctx context.Context, fn TxFunc, opts ...TxOption) error

	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryToStruct(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryToStructs(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	TransactionX(ctx context.Context, fn TxxFunc, opts ...TxOption) error
}

type (
	// 扫描到的每一行时执行这个handler, 如果返回了err则停止扫描下一行
	NextFunc func(*sql.Rows) error

	// 事务开启成功调用这个handler, 返回nil自动commit, 返回err自动回滚
	TxFunc func(*sql.Tx) error

	// 事务开启成功调用这个handler, 返回nil自动commit, 返回err自动回滚
	TxxFunc func(*sqlx.Tx) error
)

// 事务选项.
type TxOption func(*sql.TxOptions)

// 设置事务隔离级别
func WithTxIsolation(i sql.IsolationLevel) TxOption {
	return func(o *sql.TxOptions) {
		o.Isolation = i
	}
}

// 设置事务为只读
func WithTxReadOnly(readOnly bool) TxOption {
	return func(o *sql.TxOptions) {
		o.ReadOnly = readOnly
	}
}

var ErrBreak = errors.New("mysql scan rows break")

type dbClient struct {
	db *sqlx.DB
}

func (d dbClient) GetDB() *sqlx.DB { return d.db }
func (d dbClient) Unsafe() Client  { return dbClient{db: d.db.Unsafe()} }

func (d dbClient) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d dbClient) Query(ctx context.Context, next NextFunc, query string, args ...interface{}) error {
	var rows *sql.Rows
	rows, err := d.db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err = next(rows)
		if err == ErrBreak {
			break
		}
		if err != nil {
			return err
		}
	}
	err = rows.Err()
	return err
}

func (d dbClient) QueryRow(ctx context.Context, dest []interface{}, query string, args ...interface{}) error {
	row := d.db.DB.QueryRowContext(ctx, query, args...)
	return row.Scan(dest...)
}

func (d dbClient) Transaction(ctx context.Context, fn TxFunc, opts ...TxOption) error {
	txOpts := new(sql.TxOptions)
	for _, o := range opts {
		o(txOpts)
	}

	var tx *sql.Tx
	tx, err := d.db.DB.BeginTx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("begin transaction error: %w", err)
	}
	if err := fn(tx); err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("transaction error: %s, and rollback error: %w", err.Error(), e)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("commit transaction error: %s, and rollback error: %w", err.Error(), e)
		}
		return fmt.Errorf("commit transaction error: %w", err)
	}
	return nil
}

func (d dbClient) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return d.db.GetContext(ctx, dest, query, args...)
}

func (d dbClient) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return d.db.SelectContext(ctx, dest, query, args...)
}

func (d dbClient) QueryToStruct(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return d.db.QueryRowxContext(ctx, query, args...).StructScan(dst)
}

func (d dbClient) QueryToStructs(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := d.db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = sqlx.StructScan(rows, dst)
	return err
}

func (d dbClient) TransactionX(ctx context.Context, fn TxxFunc, opts ...TxOption) error {
	txOpts := new(sql.TxOptions)
	for _, o := range opts {
		o(txOpts)
	}

	tx, err := d.db.BeginTxx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("begin transaction error: %w", err)
	}
	if err := fn(tx); err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("transaction error: %s, and rollback error: %w", err.Error(), e)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("commit transaction error: %s, and rollback error: %w", err.Error(), e)
		}
		return fmt.Errorf("commit transaction error: %w", err)
	}
	return nil
}
