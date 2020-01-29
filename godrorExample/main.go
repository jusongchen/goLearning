package main

import (
	"fmt"

	//oracle db driver
	_ "github.com/godror/godror"
	// sql "github.com/jmoiron/sqlx"
	"database/sql"
)

func main() {

	connectionIdentifier := "adhoc-db1-2-crd.eng.sfdc.net/slob"
	user := "kite"
	passwd := "kite"

	err := openDB(connectionIdentifier, user, passwd)
	if err != nil {
		fmt.Printf("open DB failed:%v", err)
		return
	}
	fmt.Printf("made connection to DB %s as use %s\n", connectionIdentifier, user)

}

func openDB(connectionIdentifier, user, passwd string) (err error) {

	dataSource := fmt.Sprintf("%s/%s@%s", user, passwd, connectionIdentifier)

	db, err := sql.Open("godror", dataSource)
	if err != nil {
		return fmt.Errorf("Open DB %s as user %s failed:%w", connectionIdentifier, user, err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("Ping DB %s as user %s failed:%w", connectionIdentifier, user, err)
	}

	var dbid uint64
	err = db.QueryRow("select 1 dbid from dual").Scan(&dbid)

	if err != nil {
		return fmt.Errorf("user %s@%s does not have SELECT priviledge on DUAL:\n%w", user, connectionIdentifier, err)
	}
	return nil
}
