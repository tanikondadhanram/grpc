package main

import (
	"fmt"

	"log"

	flag "github.com/spf13/pflag"

	lua "github.com/yuin/gopher-lua"

	sqlite "crawshaw.io/sqlite"

	sqlitex "crawshaw.io/sqlite/sqlitex"
)

var dbpool *sqlitex.Pool

var conn *sqlite.Conn

func main() {
	var err error
	var luaFilePath *string = flag.String("luapath", "./my-cli/main.lua", "execute lua")
	dbpool, err = sqlitex.Open("file:memory:?mode=memory", 0, 10)
	if err != nil {
		log.Fatal(err)
	}
	flag.Parse()
	getAndRunLua(*luaFilePath)
}

func getAndRunLua(luaFilePath string) {
	L := lua.NewState()
	defer L.Close()
	L.Register("query", query)
	conn = dbpool.Get(nil)
	if conn == nil {
		return
	}
	defer dbpool.Put(conn)
	if err := L.DoFile(luaFilePath); err != nil {
		panic(err)
	}
}

func query(L *lua.LState) int {
	queryParam := L.Get(-1)
	stmt := conn.Prep(queryParam.String())
	for {
		if hasRow, err := stmt.Step(); err != nil {
			panic(err)
		} else if !hasRow {
			break
		}
		for i := 0; i < stmt.ColumnCount(); i++ {
			colName := stmt.ColumnName(i)
			fmt.Println(stmt.GetText(colName))
		}
	}
	return 1
}
