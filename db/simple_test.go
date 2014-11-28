package database

import (
	"database/sql"
	"testing"
	"fmt"
)

var db *sql.DB

func init(){
     db, _ =  CreateTable("./sample.db")
}

func TestCreateTable(t *testing.T) {

	_, err := CreateTable("./sample.db")

	if err != nil {
		t.Log("table create failed")
	}
}

func TestInsertTable(t *testing.T) {

	err := InsertTable(db, 1, "1.jpg", false)
	if err != nil {
		t.Log("table insert failed")
	}
}

func TestUpdateTable(t *testing.T) {

	err := UpdateTable(db, "1.jpg", true)
	if err != nil {
		t.Log("table update failed")
	}
}

func TestQueryTable(t *testing.T) {

	rows, err := QueryTable(db, "1.jpg")
	if err != nil {
		t.Log("table update failed")
	}

	for rows.Next() {
		var id int
		var fn string
		var complete int
		rows.Scan(&id, &fn, &complete)
		fmt.Println(id, fn, complete)
	}

}
