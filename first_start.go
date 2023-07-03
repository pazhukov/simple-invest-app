package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	var pathToConfig string
	var config_name = "default.db"
	//
	var workdir string
	var port string
	var broker_comm_in_deal string
	var new_workdir string
	var new_port string
	var new_broker_comm_in_deal string

	flag.StringVar(&pathToConfig, "db", config_name, "path to db")
	flag.Parse()

	file, err := os.Open(pathToConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	db, err := sql.Open("sqlite3", pathToConfig)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	row := db.QueryRow("select workdir, port, comm_in_deal from settings where id = 0")

	err_select := row.Scan(&workdir, &port, &broker_comm_in_deal)

	if err_select != nil {
		log.Fatal(err_select)
	}

	fmt.Println("Enter work dir path(current value in db is", workdir, ")")
	fmt.Scan(&new_workdir)
	fmt.Println("Enter port (current value in db is", port, ")")
	fmt.Scan(&new_port)
	fmt.Println("Enter calc broker commision in dela (current value in db is", broker_comm_in_deal, ")")
	fmt.Scan(&new_broker_comm_in_deal)

	_, err_update := db.Exec("update settings set workdir = ?, port = ?, comm_in_deal = ? where id = 0", new_workdir, new_port, new_broker_comm_in_deal)

	if err_update != nil {
		log.Fatal(err_update)
	}

	fmt.Println("Settings updated")

}
