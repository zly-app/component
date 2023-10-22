package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/zly-app/zapp/filter"
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

	clientType string
	clientName string
}

type clientReq struct {
	Query string
	Args  []interface{}
}
type clientRsp struct {
	DestList []interface{}
	Dest     interface{}
	Result   sql.Result
}

func (d dbClient) GetDB() *sqlx.DB { return d.db }
func (d dbClient) Unsafe() Client  { return dbClient{db: d.db.Unsafe()} }

func (d dbClient) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	meta := &filter.Meta{
		ClientType: d.clientType,
		ClientName: d.clientName,
		MethodName: "Exec",
	}
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp, err := filter.TriggerClientFilter(ctx, meta, req, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		r := req.(*clientReq)
		result, err := d.db.ExecContext(ctx, r.Query, r.Args...)
		if err != nil {
			return nil, err
		}
		return &clientRsp{Result: result}, nil
	})
	if err != nil {
		return nil, err
	}
	return rsp.(*clientRsp).Result, nil
}

func (d dbClient) Query(ctx context.Context, next NextFunc, query string, args ...interface{}) error {
	meta := &filter.Meta{
		ClientType: d.clientType,
		ClientName: d.clientName,
		MethodName: "Query",
	}
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{}
	err := filter.TriggerClientFilterInject(ctx, meta, req, rsp, func(ctx context.Context, req, _ interface{}) error {
		r := req.(*clientReq)
		rows, err := d.db.DB.QueryContext(ctx, r.Query, r.Args...)
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
		return nil
	})
	return err
}

func (d dbClient) QueryRow(ctx context.Context, dest []interface{}, query string, args ...interface{}) error {
	meta := &filter.Meta{
		ClientType: d.clientType,
		ClientName: d.clientName,
		MethodName: "QueryRow",
	}
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		DestList: dest,
	}
	err := filter.TriggerClientFilterInject(ctx, meta, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		row := d.db.DB.QueryRowContext(ctx, r.Query, r.Args...)
		return row.Scan(sp.DestList...)
	})
	return err
}

func (d dbClient) Transaction(ctx context.Context, fn TxFunc, opts ...TxOption) error {
	meta := &filter.Meta{
		ClientType: d.clientType,
		ClientName: d.clientName,
		MethodName: "Transaction",
	}
	req := &clientReq{}
	rsp := &clientRsp{}
	err := filter.TriggerClientFilterInject(ctx, meta, req, rsp, func(ctx context.Context, _, _ interface{}) error {
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
	})
	return err
}

func (d dbClient) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	meta := &filter.Meta{
		ClientType: d.clientType,
		ClientName: d.clientName,
		MethodName: "Get",
	}
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		Dest: dest,
	}
	err := filter.TriggerClientFilterInject(ctx, meta, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		return d.db.GetContext(ctx, sp.Dest, r.Query, r.Args...)
	})
	return err
}

func (d dbClient) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	meta := &filter.Meta{
		ClientType: d.clientType,
		ClientName: d.clientName,
		MethodName: "Select",
	}
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		Dest: dest,
	}
	err := filter.TriggerClientFilterInject(ctx, meta, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		return d.db.SelectContext(ctx, sp.Dest, r.Query, r.Args...)
	})
	return err
}

func (d dbClient) QueryToStruct(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	meta := &filter.Meta{
		ClientType: d.clientType,
		ClientName: d.clientName,
		MethodName: "QueryToStruct",
	}
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		Dest: dest,
	}
	err := filter.TriggerClientFilterInject(ctx, meta, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		return d.db.QueryRowxContext(ctx, r.Query, r.Args...).StructScan(sp.Dest)
	})
	return err
}

func (d dbClient) QueryToStructs(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	meta := &filter.Meta{
		ClientType: d.clientType,
		ClientName: d.clientName,
		MethodName: "QueryToStructs",
	}
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		Dest: dest,
	}
	err := filter.TriggerClientFilterInject(ctx, meta, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)

		rows, err := d.db.DB.QueryContext(ctx, r.Query, r.Args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		err = sqlx.StructScan(rows, sp.Dest)
		return err
	})
	return err
}

func (d dbClient) TransactionX(ctx context.Context, fn TxxFunc, opts ...TxOption) error {
	meta := &filter.Meta{
		ClientType: d.clientType,
		ClientName: d.clientName,
		MethodName: "TransactionX",
	}
	req := &clientReq{}
	rsp := &clientRsp{}
	err := filter.TriggerClientFilterInject(ctx, meta, req, rsp, func(ctx context.Context, _, _ interface{}) error {
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
	})
	return err
}
