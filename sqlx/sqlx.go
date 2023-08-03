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
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

type ISqlx interface {
	// 获取
	GetSqlx(name ...string) *sqlx.DB
	// 获取
	GetDefSqlx() *sqlx.DB
	// 释放
	Close()
}
type instance struct {
	*sqlx.DB
}

func (i *instance) Close() {
	_ = i.DB.Close()
}

func NewSqlx(app core.IApp, componentType ...core.ComponentType) ISqlx {
	s := &Sqlx{
		app:           app,
		conn:          conn.NewConn(),
		componentType: DefaultComponentType,
	}
	if len(componentType) > 0 {
		s.componentType = componentType[0]
	}
	return s
}

func (s *Sqlx) GetSqlx(name ...string) *sqlx.DB {
	return s.conn.GetInstance(s.makeClient, name...).(*instance).DB
}

func (s *Sqlx) GetDefSqlx() *sqlx.DB {
	return s.conn.GetInstance(s.makeClient, consts.DefaultComponentName).(*instance).DB
}

func (s *Sqlx) makeClient(name string) (conn.IInstance, error) {
	conf := newConfig()
	err := s.app.GetConfig().ParseComponentConfig(s.componentType, name, conf)
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
	return &instance{db}, nil
}

func (s *Sqlx) Close() {
	s.conn.CloseAll()
}
