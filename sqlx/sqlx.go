package sqlx

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zly-app/zapp"

	_ "github.com/ClickHouse/clickhouse-go" // clickhosue
	_ "github.com/denisenkom/go-mssqldb"    // mssql
	_ "github.com/go-sql-driver/mysql"      // mysql
	_ "github.com/lib/pq"                   // postgres
	_ "github.com/mattn/go-sqlite3"         // sqlite
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
)

type SqlxCreator struct {
	conn *conn.Conn
}

type ISqlxCreator interface {
	// 获取
	GetClient(name string) Client
	// 获取
	GetDefClient() Client
}
type instance struct {
	client Client
}

func (i *instance) Close() {
	_ = i.client.GetDB().Close()
}

func GetSqlxCreator() ISqlxCreator {
	return defCreator
}

func (s *SqlxCreator) GetClient(name string) Client {
	ins, err := s.conn.GetConn(s.makeClient, name)
	if err != nil {
		return newErrClient(err)
	}
	return ins.(*instance).client
}

func (s *SqlxCreator) GetDefClient() Client {
	return s.GetClient(consts.DefaultComponentName)
}

func (s *SqlxCreator) makeClient(name string) (conn.IInstance, error) {
	conf := newConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
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

func (s *SqlxCreator) Close() {
	s.conn.CloseAll()
}
