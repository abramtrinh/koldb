package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/abramtrinh/koldb/data"
	"github.com/abramtrinh/koldb/database"
	"github.com/abramtrinh/koldb/structs"
)

var filePath string = "./"
var itemsJSON string = filePath + "items.json"
var miniLatestPrices string = filePath + "testLatestPrices.json"
var miniMafiaPrices string = filePath + "testMafiaPrices.json"
var miniNewMarketAll string = filePath + "testNewMarketAll.json"
var miniNewMarketID string = filePath + "testNewMarketID.json"

func main() {
	//TempTestData()

	err := database.DBConnectInit()
	if err != nil {
		fmt.Printf("error DBConnectInit() %v\n", err)
		return
	}

	//TempTestInsertItems()
	//TempTestInsertMafiaPrices()
	//TempTestInsertMarketTrans()

}

// Takes in any data slice and marshals into given fileName
func MarshalToJSONFile(itemList any, fileName string) error {
	// I should check that the itemList is a valid struct from pkg.
	content, err := json.Marshal(itemList)
	if err != nil {
		return fmt.Errorf("error marshalling itemList: %w", err)
	}
	err = os.WriteFile(fileName, content, 0660)
	if err != nil {
		return fmt.Errorf("error writing to %v: %v", fileName, err)
	}
	fmt.Printf("Wrote item data to %s\n", fileName)
	return nil
}

// Opens JSON file and returns it in byte slice for unmarshalling.
func OpenReadJSONFile(fileName string) ([]byte, error) {
	// I should check that it is a JSON file.
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

// Here to test data package functionality. Should return 4 JSON files with data.
// Note: Files are written to ./koldb
func TempTestData() {
	// Time == Now
	endTime := time.Now().Unix()
	// Time == Exactly 1 Day ago
	startTime := endTime - data.EpochDay

	mUI := data.MarketURLItems()

	// Transactions of 24 hours ago till now of just Mr. A
	mUTID := data.MarketURLTransID(startTime, endTime, "194")
	// All transactions from 24 hours ago till now.
	mUTA := data.MarketURLTransAll(startTime, endTime)
	intSlice := []int{
		194,
		895,
		896,
	}
	// Latest Prices of 194, 895, 896
	marUP, _ := data.MarketURLPrices(intSlice)
	// Prices of each item's 5th listing.
	mafUP := data.MafiaURLPrices()

	fmt.Println("Starting")

	itemList0, err := data.MarketParseItems(mUI)
	if err != nil {
		fmt.Printf("error 0 MarketParseItems: %v\n", err)
	}

	if err := MarshalToJSONFile(itemList0, "items.json"); err != nil {
		fmt.Printf("error 0 MarshalToJSONFile: %v\n", err)
	}

	fmt.Println("Done Zero")
	time.Sleep(time.Second * 5)
	fmt.Println("Start First")

	itemList1, err := data.MarketParseTrans(mUTID)
	if err != nil {
		fmt.Printf("error 1 MarketParseTrans: %v\n", err)
	}

	if err := MarshalToJSONFile(itemList1, "testNewMarketID.json"); err != nil {
		fmt.Printf("error 1 MarshalToJSONFile: %v\n", err)
	}

	fmt.Println("Done First")
	time.Sleep(time.Second * 5)
	fmt.Println("Start Second")

	itemList2, err := data.MarketParseTrans(mUTA)
	if err != nil {
		fmt.Printf("error 2 MarketParseTrans: %v\n", err)
	}

	if err := MarshalToJSONFile(itemList2, "testNewMarketAll.json"); err != nil {
		fmt.Printf("error 2 MarshalToJSONFile: %v\n", err)
	}

	fmt.Println("Done Second")
	time.Sleep(time.Second * 5)
	fmt.Println("Start Third")

	itemList3, err := data.MarketParsePrices(marUP)
	if err != nil {
		fmt.Printf("error 3 MarketParsePrices: %v\n", err)
	}

	if err := MarshalToJSONFile(itemList3, "testLatestPrices.json"); err != nil {
		fmt.Printf("error 3 MarshalToJSONFile: %v\n", err)
	}

	fmt.Println("Done Third")
	time.Sleep(time.Second * 5)
	fmt.Println("Start Last")

	itemList4, err := data.MafiaParsePrices(mafUP)
	if err != nil {
		fmt.Printf("error 4 MafiaParsePrices: %v\n", err)
	}

	if err := MarshalToJSONFile(itemList4, "testMafiaPrices.json"); err != nil {
		fmt.Printf("error 4 MarshalToJSONFile: %v\n", err)
	}

	fmt.Println("Done Last")
}

// Note: Could run these batch file updates as a database transaction. Not sure if needed.
func TempTestInsertItems() {
	byteSlice, err := OpenReadJSONFile(itemsJSON)
	if err != nil {
		fmt.Printf("error openReadFile %v\n", err)
		return
	}

	var items []structs.Items
	err = json.Unmarshal(byteSlice, &items)
	if err != nil {
		fmt.Printf("error unmarshalling %v\n", err)
		return
	}

	start := time.Now()

	//Waitgroup is used so all goroutines and finish running.
	var wg sync.WaitGroup
	for i := 0; i < len(items); i++ {
		wg.Add(1)
		//Go routines used so I can concurrently insert things.
		//Speed diff of 27s->3s when using MaxOpenConns=25
		go database.InsertItems(&wg, items[i].ID, items[i].Name)
		/*
			BIG NOTE: error returns for insertItems and the like are not handled.
			Check out "handling goroutine errors" / "goroutines with error returns"
			Goroutine errors are not checked. Could not use goroutines and check errors normally.
			Use channels, error groups, or logging possibly?
		*/
	}

	wg.Wait()
	end := time.Now()
	fmt.Printf("took %v\n", end.Sub(start))
	fmt.Println("Done init items table.")
}

func TempTestInsertMafiaPrices() {
	byteSlice, err := OpenReadJSONFile(miniMafiaPrices)
	if err != nil {
		fmt.Printf("error openReadFile %v\n", err)
		return
	}

	var mafiaPrices []structs.MafiaPrices
	err = json.Unmarshal(byteSlice, &mafiaPrices)
	if err != nil {
		fmt.Printf("error unmarshalling %v\n", err)
		return
	}

	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < len(mafiaPrices); i++ {
		wg.Add(1)
		go database.InsertMafiaPrices(&wg, mafiaPrices[i].ItemID, mafiaPrices[i].Price, mafiaPrices[i].Time)
	}

	wg.Wait()
	end := time.Now()
	fmt.Printf("took %v\n", end.Sub(start))
	fmt.Println("Done init mafia prices table.")
}

func TempTestInsertMarketTrans() {
	byteSlice, err := OpenReadJSONFile(miniNewMarketAll)
	if err != nil {
		fmt.Printf("error openReadFile %v\n", err)
		return
	}

	var mafiaTrans []structs.MarketTrans
	err = json.Unmarshal(byteSlice, &mafiaTrans)
	if err != nil {
		fmt.Printf("error unmarshalling %v\n", err)
		return
	}

	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < len(mafiaTrans); i++ {
		wg.Add(1)
		go database.InsertMarketTrans(&wg, mafiaTrans[i].TransID, mafiaTrans[i].ItemID, mafiaTrans[i].Volume, mafiaTrans[i].Price, mafiaTrans[i].Time)
	}

	wg.Wait()
	end := time.Now()
	fmt.Printf("took %v\n", end.Sub(start))
	fmt.Println("Done init market trans table.")
}
