package database

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Global unexported db connector
var db *sql.DB

// Sets up the db connector pool
func DBConnectInit() error {
	// db.env should be in root aka ./koldb
	err := godotenv.Load("db.env")
	if err != nil {
		return fmt.Errorf("error loading env %w\n", err)
	}

	// Capture connection properties.
	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    os.Getenv("NET"),
		Addr:   os.Getenv("ADDRESS"),
		DBName: os.Getenv("DBNAME"),
	}

	// Get a database handle.
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("error failed opening db %w", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("error failed ping db %w", err)
	}
	fmt.Println("Connected!")
	// Set so I can use goroutines without hitting the "Too many connections" error
	db.SetMaxOpenConns(25)
	return nil
}

// Function (used with goroutines) init populates item table.
func InsertItems(wg *sync.WaitGroup, itemID int, itemName string) error {
	// wg.Done() is used because I'm using goroutines and want to finish all runs first.
	defer wg.Done()

	// Using REPLACE over INSERT because Mr. A has 1 old 1 new value.
	// REPLACE deletes old entries which causes other tables to cascade. So...
	// Instead now using INSERT ... ON DUPLICATE KEY UPDATE so above doesn't occur.
	_, err := db.Exec("INSERT INTO item (itemID, itemName) VALUES (?, ?) ON DUPLICATE KEY UPDATE itemName=?", itemID, itemName, itemName)
	if err != nil {
		//temp need to remove either goroutines or use channels to return errors
		fmt.Printf("error InsertItems %v\n", err)
		// Should I fatal/exit if this fails? Don't want to continue if bad insert. tx?
		return fmt.Errorf("error func(insertItems) db.Exec() %w\n", err)
	}
	return nil
}

// Function (used with goroutines) init populates prices table.
func InsertMafiaPrices(wg *sync.WaitGroup, itemID int, cost int, epochTime int64) error {
	defer wg.Done()

	var tempItemID int
	// Check if item that has a price is not in the "currently tradeable" item list.
	// Stops foreign key constraint fails.
	row := db.QueryRow("SELECT * FROM item WHERE itemID=?", itemID)
	if err := row.Scan(&tempItemID); err != nil {
		if err == sql.ErrNoRows {
			// Reason we don't throw error is because I'm trying to ignore old items that used to be tradeable.
			return nil
		}
	}

	// INSERT ... ON DUPLICATE KEY UPUDATE is good here because update prices regularly.
	_, err := db.Exec("INSERT INTO prices (itemID, cost, epochTime) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE cost=?, epochTime=?",
		itemID, cost, epochTime, cost, epochTime)
	if err != nil {
		//temp need to remove either goroutines or use channels to return errors
		fmt.Printf("error InsertMafiaPrices %v\n", err)
		return fmt.Errorf("error func(insertMafiaPrices) db.Exec() %w\n", err)
	}
	return nil
}

// Function (used with goroutines) init populates transactions table.
func InsertMarketTrans(wg *sync.WaitGroup, transID int, itemID int, volume int, cost float32, epochTime int64) error {
	defer wg.Done()
	// INSERT ... ON DUPLICATE KEY UPDATE is used instead of INSERT IGNORE because latter supresses errors.
	// transID=transID is used for the UPDATE because MySQL doesn't actually do the update.
	_, err := db.Exec("INSERT INTO transactions (transID, itemID, volume, cost, epochTime) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE transID=transID",
		transID, itemID, volume, cost, epochTime)
	if err != nil {
		//temp need to remove either goroutines or use channels to return errors
		fmt.Printf("error InsertMarketTrans %v\n", err)
		return fmt.Errorf("error func(insertMarketTrans) db.Exec() %w\n", err)
	}
	return nil
}
