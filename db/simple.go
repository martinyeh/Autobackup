package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"math/rand"
	"os"
)

type DB struct {
	db *sql.DB // db object
	tx *sql.Tx // db transaction
}

func NewConn(dbfile string) *DB {
	db := new(DB)
	var err error
	db.db, err = CreateTable(dbfile)
	if err != nil {
		log.Println(err)
	}
	db.tx, err = db.db.Begin()
	if err != nil {
		log.Println(err)
	}

	return db
}

func CreateTable(dbfile string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	db, err = sql.Open("sqlite3", "file:"+dbfile+"?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	//defer db.Close()

	sqlStmt := `create table if not exists fileinfo (id integer not null primary key, fn text, complete numeric);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}

	return db, nil
}

func InsertTable(db *sql.DB, id int, fn string, complete bool) error {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
		return err
	}
	stmt, err := tx.Prepare("insert into fileinfo(id, fn, complete) values(?, ?, ?)")
	if err != nil {
		log.Fatal(stmt, err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(id, fn, complete)
	if err != nil {
		log.Fatal("insert table error", err)
		return err
	}

	tx.Commit()
	return nil
}

func (db DB) UpdateTable(fn string, complete bool) error {
	/*tx, err := db.Begin()
	if err != nil {
		log.Println("Begin update table", err)
		//return err
	}*/

	var value int
	if complete {
		value = 1
	} else {
		value = 0
	}

	stmt, err := db.tx.Prepare("update fileinfo set complete=? where fn =?")
	if err != nil {
		log.Fatal(stmt, err)
		return err
	}

	defer stmt.Close()

	res, err := stmt.Exec(value, fn)
	if err != nil {
		log.Fatal("update table error:", err)
		return err
	}

	affect, err := res.RowsAffected()
	fmt.Println(affect)

	db.tx.Commit()
	db.db.Close()
	return nil
}

func (db DB) QueryTable(fn string) (*sql.Rows, error) {
	rows, err := db.db.Query("select id, fn, complete from fileinfo where fn ='" + fn + "'")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	db.db.Close()
	return rows, nil
}

func main() {
	os.Remove("./sample.db")

	db, err := sql.Open("sqlite3", "./sample.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
    create table fileinfo (id integer not null primary key, fn text, size numeric);
    delete from fileinfo;
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into fileinfo(id, fn, size) values(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("圖片%d", i), 1000+rand.Intn(10000))
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()

	rows, err := db.Query("select id, fn, size from fileinfo")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var fn string
		var size int
		rows.Scan(&id, &fn, &size)
		fmt.Println(id, fn, size)
	}
	rows.Close()

	fmt.Println("end")

}
