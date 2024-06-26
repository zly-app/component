package sqlx

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"

	_ "github.com/ClickHouse/clickhouse-go" // clickhosue
	_ "github.com/denisenkom/go-mssqldb"    // mssql
	_ "github.com/go-sql-driver/mysql"      // mysql
	_ "github.com/lib/pq"                   // postgres
	_ "github.com/mattn/go-sqlite3"         // sqlite
)

type Sqlx struct {
	app  core.IApp
	conn *conn.Conn
}

type ISqlx interface {
	// 获取
	GetSqlx(name ...string) Client
	// 获取
	GetDefSqlx() Client
	// 释放
	Close()
}
type instance struct {
	client Client
}

func (i *instance) Close() {
	_ = i.client.GetDB().Close()
}

func NewSqlx(app core.IApp) ISqlx {
	s := &Sqlx{
		app:  app,
		conn: conn.NewConn(),
	}
	return s
}

func (s *Sqlx) GetSqlx(name ...string) Client {
	ins, err := s.conn.GetConn(s.makeClient, name...)
	if err != nil {
		return newErrClient(err)
	}
	return ins.(*instance).client
}

func (s *Sqlx) GetDefSqlx() Client {
	return s.GetSqlx(consts.DefaultComponentName)
}

func (s *Sqlx) makeClient(name string) (conn.IInstance, error) {
	conf := newConfig()
	err := s.app.GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		return nil, fmt.Errorf("sqlx的配置错误: %v", err)
	}

	db, err := sqlx.Open(conf.Driver, conf.Source)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(conf.MaxIdleConns)
	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Millisecond)
	client := dbClient{
		db:   db,
		name: name,
	}
	return &instance{client}, nil
}

func (s *Sqlx) Close() {
	s.conn.CloseAll()
}
