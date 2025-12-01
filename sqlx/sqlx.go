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

type sqlxCreator struct {
	conn *conn.AnyConn[Client]
}

type Creator interface {
	// 获取客户端
	GetClient(name string) Client
	// 获取默认客户端
	GetDefClient() Client
}

func GetCreator() Creator {
	return defCreator
}

func (s *sqlxCreator) GetClient(name string) Client {
	ins, err := s.conn.GetConn(s.makeClient, name)
	if err != nil {
		return newErrClient(err)
	}
	return ins
}

func (s *sqlxCreator) GetDefClient() Client {
	return s.GetClient(consts.DefaultComponentName)
}

func (s *sqlxCreator) makeClient(name string) (Client, error) {
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
	return client, nil
}

func (s *sqlxCreator) Close() {
	s.conn.CloseAll()
}
