package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// THOUGHTS: RIP Overloading
// I guess just write each to JSON then change later.

// http://dcdb.coldfront.net/collections/index.cgi?query_type=info&query_value=sets
// Validate Data + URL
// Testing
// Some Error checking needs to be done for the _, _ errors.

// Verify file type. Data type. ??? Fields? What else.
// Temp/Test files for testing and checking if things are working correctly.
// Think assert ^

//how to keep track of time last got make it so you limit times you do this
// Don't want to keep running get new file every time.

// Could reimplement this as interface I think?

const (
	//Possible issues using epoch time? Leap seconds?
	epochDay  int64 = 86400
	epochHour int64 = 3600
)

// struct is exported for XML handling
type Market struct {
	XMLName xml.Name       `xml:"marketplace"`
	Details []Transactions `xml:"trans"`
}

// NOTE: Changed Cost from int to float32 since there are decimals in the prices.
type Transactions struct {
	XMLName xml.Name `xml:"trans"`
	TransID int      `xml:"id,attr"`
	ItemID  int      `xml:"itemid"`
	Vol     int      `xml:"vol"`
	Cost    float32  `xml:"cost"`
	When    int64    `xml:"when"`
}

type MarketTrans struct {
	TransID int     `json:"trans"`
	ItemID  int     `json:"itemid"`
	Volume  int     `json:"vol"`
	Price   float32 `json:"price"`
	Time    int64   `json:"time"`
}

type MarketPrices struct {
	ItemID int `json:"itemid"`
	Price  int `json:"Price"`
}

type MafiaPrices struct {
	ItemID int   `json:"itemid"`
	Time   int64 `json:"time"`
	Price  int   `json:"price"`
}

func main() {
	// Time == Now
	endTime := time.Now().Unix()
	// Time == Exactly 1 Day ago
	startTime := endTime - epochDay

	// Transactions of 24 hours ago till now of just Mr. A
	mUTID := marketURLTransID(startTime, endTime, "194")
	// All transactions from 24 hours ago till now.
	mUTA := marketURLTransAll(startTime, endTime)
	intSlice := []int{
		194,
		895,
		896,
	}
	// Latest Prices of 194, 895, 896
	marUP, _ := marketURLPrices(intSlice)
	// Prices of each item's 5th listing.
	mafUP := mafiaURLPrices()

	fmt.Println("Starting")

	itemList1, _ := marketParseTrans(mUTID)
	marshalToJSON(itemList1, "testNewMarketID.json")
	fmt.Println("Done First")

	time.Sleep(time.Second * 10)
	fmt.Println("Start Second")
	itemList2, _ := marketParseTrans(mUTA)
	marshalToJSON(itemList2, "testNewMarketAll.json")
	fmt.Println("Done Second")

	time.Sleep(time.Second * 10)
	fmt.Println("Start Third")
	itemList3, _ := marketParsePrices(marUP)
	marshalToJSON(itemList3, "testLatestPrices.json")
	fmt.Println("Done Third")

	time.Sleep(time.Second * 10)
	fmt.Println("Start Last")
	itemList4, _ := mafiaParsePrices(mafUP)
	marshalToJSON(itemList4, "testMafiaPrices.json")
	fmt.Println("Done Last")
}

func marshalToJSON(itemList any, fileName string) error {
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

// Gets html page and locally stores it in working directory.
// func getURLToFile(URL string, fileName string) error {
// 	fmt.Printf("Beginning to get data from %s\n", URL)
// 	// Requesting HTML page
// 	resp, err := http.Get(URL)
// 	if err != nil {
// 		return fmt.Errorf("failed getting URL: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	outputFile, err := os.Create(fileName)
// 	if err != nil {
// 		return fmt.Errorf("failed creating file: %w", err)
// 	}
// 	defer outputFile.Close()

// 	if _, err := io.Copy(outputFile, resp.Body); err != nil {
// 		return fmt.Errorf("failed copying file content: %w", err)
// 	}
// 	fmt.Printf("Completed, data written to %s\n", fileName)
// 	return nil
// }

// Creates URL that returns all transactions for itemid occuring in specified time frame on ColdFront.
func marketURLTransID(start int64, end int64, itemid string) string {
	// The reason itemid is a string not int is because I need it to be "" sometimes.
	return fmt.Sprintf("http://kol.coldfront.net/newmarket/export.php?start=%d&end=%d&itemid=%s", start, end, itemid)
}

// Creates URL that returns all transactions occuring in specified time frame on ColdFront.
// NOTE: Simulates function overloading. Kinda feels weird doing this.
func marketURLTransAll(start int64, end int64) string {
	//return fmt.Sprintf("http://kol.coldfront.net/newmarket/export.php?start=%d&end=%d&itemid=", start, end)
	return marketURLTransID(start, end, "")
}

// Parses the ColdFront newmarket XML transaction data and returns to slice ready for json marshal.
func marketParseTrans(URL string) ([]MarketTrans, error) {
	// https://kol.coldfront.net/newmarket/export.php?start=1674968400&end=1674969465&itemid=
	// Data is in XML format.
	// Incoming data format: TransactionID: (ItemId Volume Cost Time)
	resp, err := http.Get(URL)
	if err != nil {
		return nil, fmt.Errorf("error getting URL: %w", err)
	}
	defer resp.Body.Close()

	// ReadAll is used to put data into []byte
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body to slice: %w", err)
	}

	// XML data is parsed and stored into Market struct based on the <tags>
	var market Market
	xml.Unmarshal(body, &market)

	// Now do stuff with market and create json from it.
	// itemList slice is used to store each item xml block
	var itemList []MarketTrans

	// Iterate through all the unmarshalled xml items.
	for i := 0; i < len(market.Details); i++ {
		// Create a new item from each market.Detail[i] item and creates a struct for it
		newItem := MarketTrans{
			TransID: market.Details[i].TransID,
			ItemID:  market.Details[i].ItemID,
			Volume:  market.Details[i].Vol,
			Price:   market.Details[i].Cost,
			Time:    market.Details[i].When,
		}
		// Append new item for storage.
		itemList = append(itemList, newItem)
	}

	return itemList, nil
}

// Creates URL that returns up to 10 itemid and its current price on ColdFront.
func marketURLPrices(itemIDs []int) (string, error) {
	length := len(itemIDs)
	// URL only allows max xof 10 items to be requested.
	if length > 10 {
		return "", fmt.Errorf("error, item slice len is %d. max 10 items\n", length)
	}

	URL := "https://kol.coldfront.net/newmarket/latestprice.php?"
	for index, value := range itemIDs {
		itemString := fmt.Sprintf("item%d=%d", index+1, value)
		// If not last item in URL, needs &
		if length-1 != index {
			itemString += "&"
		}
		URL += itemString
	}

	return URL, nil
}

// Parses the ColdFront newmarket lastest item prices into usable format.
func marketParsePrices(URL string) ([]MarketPrices, error) {
	// NOTE: Can reimplement using the net/html golang pkg instead.
	// https://kol.coldfront.net/newmarket/latestprice.php?
	// Data is just html.
	// Incoming data format: "itemid,latestprice"<br>
	resp, err := http.Get(URL)
	if err != nil {
		return nil, fmt.Errorf("failed getting URL: %w", err)
	}
	defer resp.Body.Close()

	var itemList []MarketPrices

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		// Trims <br> from each line
		line := strings.TrimRight(scanner.Text(), "<br>")
		// Splits into [itemid latestprice]
		lineSlice := strings.Split(line, ",")

		id, _ := strconv.Atoi(lineSlice[0])
		price, _ := strconv.Atoi(lineSlice[1])
		newItem := MarketPrices{
			ItemID: id,
			Price:  price,
		}

		itemList = append(itemList, newItem)
	}

	return itemList, nil
}

// Creates and returns a URL string to kolmafia's item:time:price list.
func mafiaURLPrices() string {
	return "https://kolmafia.us/scripts/updateprices.php?action=getmap"
}

// Parses the kolmafia's item:time:price data list into useable format.
func mafiaParsePrices(URL string) ([]MafiaPrices, error) {
	// https://kolmafia.us/scripts/updateprices.php?action=getmap
	// Data is a pure txt file.
	// Incoming data format: ItemId	TimeLastUpdated	Price(of the 5th item)
	resp, err := http.Get(URL)
	if err != nil {
		return nil, fmt.Errorf("failed getting URL: %w", err)
	}
	defer resp.Body.Close()

	var itemList []MafiaPrices
	// Scanner used so I can parse line by line.
	scanner := bufio.NewScanner(resp.Body)
	//Skipping the first line because it isn't needed
	scanner.Scan()
	for scanner.Scan() {
		lineSlice := strings.Fields(scanner.Text())

		// Need to convert string to int so I can use for json. Also ParseInt since int64
		id, _ := strconv.Atoi(lineSlice[0])
		time, _ := strconv.ParseInt(lineSlice[1], 10, 64)
		price, _ := strconv.Atoi(lineSlice[2])

		newItem := MafiaPrices{
			ItemID: id,
			Time:   time,
			Price:  price,
		}

		itemList = append(itemList, newItem)
	}

	return itemList, nil
}
