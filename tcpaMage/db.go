package main

import (
	"fmt"
	"time"

	//just need in db
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	log "github.com/sirupsen/logrus"
)

var dbconnLoger *log.Entry

//Db  db
var Db *sqlx.DB
var (
	port    int    = 3306
	dbName  string = "oai_db"
	charset string = "utf8"
)

func connectMysql(ipAddrees string, userName string, password string) *sqlx.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", userName, password, ipAddrees, port, dbName, charset)
	Db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil
	}
	return Db
}

func disconnectMysql(Db *sqlx.DB) {
	Db.Close()
}

func ping(Db *sqlx.DB) {
	err := Db.Ping()
	if err != nil {
	} else {
	}
}

//InitDbAndAuc  init
func init() {

	/* mysql  */
	dbIP := GViperCfg.GetString("db.mysql_ip")
	userName := GViperCfg.GetString("db.db_user")
	userPass := GViperCfg.GetString("db.db_pass")

	for {
		Db = connectMysql(dbIP, userName, userPass)

		if Db == nil {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	ping(Db)
}

//QueryTcparByUeIP quert tcpa_flag
func QueryTcparByUeIP(ueIP string) bool {

	var tcpaFlag string
	row := Db.QueryRow("SELECT `tcpa_flag` FROM `user_pdn` WHERE `pdn_ipv4`=? LIMIT 1", ueIP)
	if err := row.Scan(&tcpaFlag); err != nil {
		return false
	}

	if tcpaFlag == "ON" {
		return true
	}

	return false
}
