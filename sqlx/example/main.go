package main

import (
	"context"

	"github.com/didi/gendry/builder"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/pkg/utils"

	"github.com/zly-app/component/sqlx"
)

/*
create table test
(
    id int unsigned auto_increment
        primary key,
    a  int unsigned default 0  not null,
    b  varchar(32)  default '' not null
);
*/

type Model struct {
	ID int    `db:"id"`
	A  int    `db:"a"`
	B  string `db:"b"`
}

func main() {
	app := zapp.NewApp("zapp.test.sqlx")
	defer app.Exit()

	const table = "test.test"

	ctx, span := utils.Otel.StartSpan(context.Background(), "start")
	defer utils.Otel.EndSpan(span)

	// 执行
	// insertSql = INSERT INTO test.test (a,b) VALUES (?,?),(?,?),(?,?),(?,?),(?,?)
	// vals = [1 v1 2 v2 3 v3 4 v4 5 v5]
	insertSql, vals, _ := builder.BuildInsert(table, []map[string]interface{}{
		{"a": 1, "b": "v1"},
		{"a": 2, "b": "v2"},
		{"a": 3, "b": "v3"},
		{"a": 4, "b": "v4"},
		{"a": 5, "b": "v5"},
	})
	_, _ = sqlx.GetDefClient().Exec(ctx, insertSql, vals...)

	// 事务
	_ = sqlx.GetDefClient().Transaction(ctx, func(ctx context.Context, tx sqlx.Tx) error {
		// queryRowCond = SELECT a,b FROM test.test WHERE (id=?)
		// vals = [1]
		queryRowCond, vals, _ := builder.BuildSelect(table, map[string]interface{}{"id": 1}, []string{"a", "b"})

		var a int
		var b string
		_ = tx.FindColumn(ctx, []interface{}{&a, &b}, queryRowCond, vals...)
		return nil
	})

	// 事务
	_ = sqlx.GetDefClient().TransactionX(ctx, func(ctx context.Context, txx sqlx.Txx) error {
		m := Model{}
		_ = txx.FindOne(ctx, &m, `select * from test.test where id = ?`, 1)
		var id int
		_ = txx.FindOne(ctx, &id, `select count(1) from test.test`)

		m3 := []Model{}
		_ = txx.Find(ctx, &m3, `select * from test.test where id = 1 limit 1`)
		m4 := []int{}
		_ = txx.Find(ctx, &m4, `select id from test.test`)
		return nil
	})

}
