package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var filePath string = "./files/miniJSON/"
var miniLatestPrices string = filePath + "miniLatestPrices.json"
var miniMafiaPrices string = filePath + "miniMafiaPrices.json"
var miniNewMarketAll string = filePath + "miniNewMarketAll.json"
var miniNewMarketID string = filePath + "miniNewMarketID.json"
var itemsJSON string = filePath + "items.json"

type MarketTrans struct {
	TransID int     `json:"trans"`
	ItemID  int     `json:"itemid"`
	Volume  int     `json:"vol"`
	Price   float32 `json:"price"`
	Time    int64   `json:"time"`
}

type MarketPrice struct {
	ItemID int `json:"itemid"`
	Price  int `json:"Price"`
}

type MafiaPrice struct {
	ItemID int   `json:"itemid"`
	Time   int64 `json:"time"`
	Price  int   `json:"price"`
}

type Item struct {
	Name string `json:"name"`
	ID   int    `json:"itemid"`
}

func main() {
	err := godotenv.Load("db.env")
	if err != nil {
		fmt.Printf("error loading env %v\n", err)
		return
	}

	var db *sql.DB
	db, err = dbConnectInit()
	if err != nil {
		fmt.Printf("error dbConnectInit() %v\n", err)
		return
	}

	// Beginning of mass loading items table.
	byteSlice, err := openReadFile(itemsJSON)
	if err != nil {
		fmt.Printf("error openReadFile %v\n", err)
		return
	}

	var items []Item
	err = json.Unmarshal(byteSlice, &items)
	if err != nil {
		fmt.Printf("error unmarshalling %v\n", err)
		return
	}

	//Waitgroup is used so all goroutines and finish running.
	var wg sync.WaitGroup
	for i := 0; i < len(items); i++ {
		wg.Add(1)
		//Go routines used so I can concurrently insert things.
		//Speed diff of 27s->112ms when using MaxOpenConns=100
		go insertItems(&wg, db, items, i)
		/*
			BIG NOTE: error returns for insertItems and the like are not handled.
			Check out "handling goroutine errors" / "goroutines with error returns"
		*/

	}

	wg.Wait()
	fmt.Println("Done init items table.")
	// End of mass loading items table.

	// Beginning of mass loading prices table.
	byteSlice, err = openReadFile(miniMafiaPrices)
	if err != nil {
		fmt.Printf("error openReadFile %v\n", err)
		return
	}

	var itemsMafiaPrices []MafiaPrice
	err = json.Unmarshal(byteSlice, &itemsMafiaPrices)
	if err != nil {
		fmt.Printf("error unmarshalling %v\n", err)
		return
	}

	for i := 0; i < len(itemsMafiaPrices); i++ {
		wg.Add(1)
		go insertMafiaPrices(&wg, db, itemsMafiaPrices, i)
	}

	wg.Wait()
	fmt.Println("Done init prices table.")
	// End of mass loading prices table.

	// Beginning of mass loading transactions table.
	byteSlice, err = openReadFile(miniNewMarketAll)
	if err != nil {
		fmt.Printf("error openReadFile %v\n", err)
		return
	}

	var itemMarketTrans []MarketTrans
	err = json.Unmarshal(byteSlice, &itemMarketTrans)
	if err != nil {
		fmt.Printf("error unmarshalling %v\n", err)
		return
	}

	for i := 0; i < len(itemMarketTrans); i++ {
		wg.Add(1)
		go insertMarketTrans(&wg, db, itemMarketTrans, i)
	}

	wg.Wait()
	fmt.Println("Done init transactions table.")
	// End of mass loading transactions table.

	fmt.Println("Done with all init.")

}

// Connects & initializes database handle using .env and sets max open connections.
func dbConnectInit() (*sql.DB, error) {
	// Capture connection properties.
	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    os.Getenv("NET"),
		Addr:   os.Getenv("ADDRESS"),
		DBName: os.Getenv("DBNAME"),
	}

	// Get a database handle.
	var db *sql.DB
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("error failed opening db %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error failed ping db %w", err)
	}
	fmt.Println("Connected!")
	// Set so I can use goroutines without hitting the "Too many connections" error
	db.SetMaxOpenConns(25)
	return db, nil
}

// Opens JSON file and returns it in byte slice for unmarshalling.
func openReadFile(fileName string) ([]byte, error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	fmt.Printf("Opened %s.\n", fileName)
	defer jsonFile.Close()

	byteSlice, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return byteSlice, nil

}

// Function (used with goroutines) init populates item table.
func insertItems(wg *sync.WaitGroup, db *sql.DB, items []Item, index int) error {
	// wg.Done() is used because I'm using goroutines and want to finish all runs first.
	defer wg.Done()
	// db var passed in so I don't have to use global variables.
	// Using REPLACE over INSERT because Mr. A has 1 old 1 new value.
	_, err := db.Exec("REPLACE INTO item (itemID, itemName) VALUES (?, ?)", items[index].ID, items[index].Name)
	if err != nil {
		// Should I fatal/exit if this fails? Don't want to continue if bad insert. tx?
		return fmt.Errorf("error func(insertItems) db.Exec() %w\n", err)
	}
	return nil
}

// Function (used with goroutines) init populates prices table.
func insertMafiaPrices(wg *sync.WaitGroup, db *sql.DB, itemsMafiaPrices []MafiaPrice, index int) error {
	defer wg.Done()
	_, err := db.Exec("INSERT INTO prices (itemID, cost, epochTime) VALUES (?, ?, ?)", itemsMafiaPrices[index].ItemID, itemsMafiaPrices[index].Price, itemsMafiaPrices[index].Time)
	if err != nil {
		return fmt.Errorf("error func(insertMafiaPrices) db.Exec() %w\n", err)
	}
	return nil
}

// Function (used with goroutines) init populates transactions table.
func insertMarketTrans(wg *sync.WaitGroup, db *sql.DB, itemMarketTrans []MarketTrans, index int) error {
	defer wg.Done()
	// NOTE: itemMarketTrans[index].TransID exists but is not inserted. Using AUTO_INCREMENT pk instead.
	_, err := db.Exec("INSERT INTO transactions (itemID, volume, cost, epochTime) VALUES (?, ?, ?, ?)", itemMarketTrans[index].ItemID, itemMarketTrans[index].Volume, itemMarketTrans[index].Price, itemMarketTrans[index].Time)
	if err != nil {
		return fmt.Errorf("error func(insertMarketTrans) db.Exec() %w\n", err)
	}
	return nil
}

// //pass in i because closure and scoping. if no pass, when func runs, will just use curr i not i when it was called
// go func(i int) {
// 	defer wg.Done()
// 	_, err := db.Exec("REPLACE INTO item (itemID, itemName) VALUES (?, ?)", items[i].ID, items[i].Name)
// 	//time.Sleep(time.Millisecond)
// 	if err != nil {
// 		// fmt.Println(items[i])
// 		fmt.Printf("error db exec %v\n", err)
// 		return
// 	}
// }(i)
