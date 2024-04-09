package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zly-app/zapp/filter"
)

var ErrNoRows = sql.ErrNoRows

type Client interface {
	GetDB() *sqlx.DB

	// 不安全模式, 在安全模式下, 如果 select 语句的字段在 scan 结构体中未定义会报错
	Unsafe() Client

	// 查询出多行记录并扫描到 dest 列表中, 记录未找到不会报错
	Find(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	// 查询出一行记录或一个列并扫描到 dest 中, 记录未找到会返回 ErrNoRows
	FindOne(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	// 查询一条记录的多个列并依次扫描到 dest 内, 列数量和 dest 长度必须相同, 记录未找到会返回 ErrNoRows
	FindColumn(ctx context.Context, dest []interface{}, query string, args ...interface{}) error

	// 执行一条语句
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Query(ctx context.Context, next NextFunc, query string, args ...interface{}) error

	// 开启事务
	Transaction(ctx context.Context, fn TxFunc, opts ...TxOption) error
	// 开启事务
	TransactionX(ctx context.Context, fn TxxFunc, opts ...TxOption) error
}

type Tx interface {
	Tx() *sql.Tx

	// 查询一条记录的多个列并依次扫描到 dest 内, 列数量和 dest 长度必须相同, 记录未找到会返回 ErrNoRows
	FindColumn(ctx context.Context, dest []interface{}, query string, args ...interface{}) error
	// 同 Find, 要求 dest 必须是 []struct
	FindToStructs(ctx context.Context, dst interface{}, query string, args ...interface{}) error

	// 执行一条语句
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Query(ctx context.Context, next NextFunc, query string, args ...interface{}) error
}

type Txx interface {
	Tx() *sql.Tx
	Txx() *sqlx.Tx

	// 不安全模式, 在安全模式下, 如果 select 语句的字段在 scan 结构体中未定义会报错
	Unsafe() Txx

	// 查询出多行记录并扫描到 dest 列表中, 记录未找到不会报错
	Find(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	// 查询出一行记录或一个列并扫描到 dest 中, 记录未找到会返回 ErrNoRows
	FindOne(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	// 查询一条记录的多个列并依次扫描到 dest 内, 列数量和 dest 长度必须相同, 记录未找到会返回 ErrNoRows
	FindColumn(ctx context.Context, dest []interface{}, query string, args ...interface{}) error

	// 执行一条语句
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Query(ctx context.Context, next NextFunc, query string, args ...interface{}) error
}

type (
	// 扫描到的每一行时执行这个handler, 如果返回了err则停止扫描下一行
	NextFunc func(ctx context.Context, rows *sql.Rows) error

	// 事务开启成功调用这个handler, 返回nil自动commit, 返回err自动回滚
	TxFunc func(ctx context.Context, tx Tx) error

	// 事务开启成功调用这个handler, 返回nil自动commit, 返回err自动回滚
	TxxFunc func(ctx context.Context, txx Txx) error
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

// 停止调用 NextFunc
var ErrBreakNext = errors.New("mysql scan rows break")

type dbClient struct {
	db *sqlx.DB

	name string
}

type clientReq struct {
	Query string
	Args  []interface{}
}
type clientRsp struct {
	IsNoRows bool
	DestList []interface{}
	Dest     interface{}
	Result   sql.Result
}

func (d dbClient) GetDB() *sqlx.DB { return d.db }
func (d dbClient) Unsafe() Client  { return dbClient{db: d.db.Unsafe(), name: d.name} }

func (d dbClient) Find(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		Dest: dest,
	}
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "Find")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		return d.db.SelectContext(ctx, sp.Dest, r.Query, r.Args...)
	})
	return err
}
func (d dbClient) FindOne(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		Dest: dest,
	}
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "FindOne")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		err := d.db.GetContext(ctx, sp.Dest, r.Query, r.Args...)
		if err == ErrNoRows {
			sp.IsNoRows = true
			return nil
		}
		return err
	})
	if rsp.IsNoRows {
		return ErrNoRows
	}
	return err
}
func (d dbClient) FindColumn(ctx context.Context, dest []interface{}, query string, args ...interface{}) error {
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		DestList: dest,
	}
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "FindColumn")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		row := d.db.DB.QueryRowContext(ctx, r.Query, r.Args...)
		err := row.Scan(sp.DestList...)
		if err == ErrNoRows {
			sp.IsNoRows = true
			return nil
		}
		return err
	})
	if rsp.IsNoRows {
		return ErrNoRows
	}
	return err
}

func (d dbClient) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	req := &clientReq{
		Query: query,
		Args:  args,
	}

	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "Exec")
	rsp, err := chain.Handle(ctx, req, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
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
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{}

	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "Query")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, _ interface{}) error {
		r := req.(*clientReq)
		rows, err := d.db.DB.QueryContext(ctx, r.Query, r.Args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			err = next(ctx, rows)
			if err == ErrBreakNext {
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

func (d dbClient) Transaction(ctx context.Context, fn TxFunc, opts ...TxOption) error {
	req := &clientReq{}
	rsp := &clientRsp{}
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "Transaction")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, _, _ interface{}) error {
		txOpts := new(sql.TxOptions)
		for _, o := range opts {
			o(txOpts)
		}

		var tx *sql.Tx
		tx, err := d.db.DB.BeginTx(ctx, txOpts)
		if err != nil {
			return fmt.Errorf("begin transaction error: %w", err)
		}
		if err := fn(ctx, &dbTx{tx: tx, name: d.name}); err != nil {
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
func (d dbClient) TransactionX(ctx context.Context, fn TxxFunc, opts ...TxOption) error {
	req := &clientReq{}
	rsp := &clientRsp{}
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "TransactionX")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, _, _ interface{}) error {
		txOpts := new(sql.TxOptions)
		for _, o := range opts {
			o(txOpts)
		}

		tx, err := d.db.BeginTxx(ctx, txOpts)
		if err != nil {
			return fmt.Errorf("begin transaction error: %w", err)
		}
		if err := fn(ctx, dbTxx{txx: tx, name: d.name}); err != nil {
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

type dbTx struct {
	tx   *sql.Tx
	name string
}

func (d dbTx) Tx() *sql.Tx { return d.tx }

func (d dbTx) FindColumn(ctx context.Context, dest []interface{}, query string, args ...interface{}) error {
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		DestList: dest,
	}
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "FindColumn")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		row := d.tx.QueryRowContext(ctx, r.Query, r.Args...)
		err := row.Scan(sp.DestList...)
		if err == ErrNoRows {
			sp.IsNoRows = true
			return nil
		}
		return err
	})
	if rsp.IsNoRows {
		return ErrNoRows
	}
	return err
}
func (d dbTx) FindToStructs(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		Dest: dest,
	}
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "FindToStructs")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)

		rows, err := d.tx.QueryContext(ctx, r.Query, r.Args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		err = sqlx.StructScan(rows, sp.Dest)
		return err
	})
	return err
}

func (d dbTx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	req := &clientReq{
		Query: query,
		Args:  args,
	}

	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "Exec")
	rsp, err := chain.Handle(ctx, req, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		r := req.(*clientReq)
		result, err := d.tx.ExecContext(ctx, r.Query, r.Args...)
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

func (d dbTx) Query(ctx context.Context, next NextFunc, query string, args ...interface{}) error {
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{}

	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "Query")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, _ interface{}) error {
		r := req.(*clientReq)
		rows, err := d.tx.QueryContext(ctx, r.Query, r.Args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			err = next(ctx, rows)
			if err == ErrBreakNext {
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

type dbTxx struct {
	txx  *sqlx.Tx
	name string
}

func (d dbTxx) Tx() *sql.Tx   { return d.txx.Tx }
func (d dbTxx) Txx() *sqlx.Tx { return d.txx }
func (d dbTxx) Unsafe() Txx   { return dbTxx{txx: d.txx.Unsafe(), name: d.name} }

func (d dbTxx) Find(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		Dest: dest,
	}
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "Find")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		return d.txx.SelectContext(ctx, sp.Dest, r.Query, r.Args...)
	})
	return err
}
func (d dbTxx) FindOne(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		Dest: dest,
	}
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "FindOne")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		err := d.txx.GetContext(ctx, sp.Dest, r.Query, r.Args...)
		if err == ErrNoRows {
			sp.IsNoRows = true
			return nil
		}
		return err
	})
	if rsp.IsNoRows {
		return ErrNoRows
	}
	return err
}
func (d dbTxx) FindColumn(ctx context.Context, dest []interface{}, query string, args ...interface{}) error {
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{
		DestList: dest,
	}
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "FindColumn")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, rsp interface{}) error {
		r := req.(*clientReq)
		sp := rsp.(*clientRsp)
		row := d.txx.Tx.QueryRowContext(ctx, r.Query, r.Args...)
		err := row.Scan(sp.DestList...)
		if err == ErrNoRows {
			sp.IsNoRows = true
			return nil
		}
		return err
	})
	if rsp.IsNoRows {
		return ErrNoRows
	}
	return err
}

func (d dbTxx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	req := &clientReq{
		Query: query,
		Args:  args,
	}

	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "Exec")
	rsp, err := chain.Handle(ctx, req, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		r := req.(*clientReq)
		result, err := d.txx.ExecContext(ctx, r.Query, r.Args...)
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

func (d dbTxx) Query(ctx context.Context, next NextFunc, query string, args ...interface{}) error {
	req := &clientReq{
		Query: query,
		Args:  args,
	}
	rsp := &clientRsp{}

	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), d.name, "Query")
	err := chain.HandleInject(ctx, req, rsp, func(ctx context.Context, req, _ interface{}) error {
		r := req.(*clientReq)
		rows, err := d.txx.Tx.QueryContext(ctx, r.Query, r.Args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			err = next(ctx, rows)
			if err == ErrBreakNext {
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
