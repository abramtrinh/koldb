package structs

// JSON struct is imported as pkg because it is used by multiple other pkgs.
// Don't want accidental circular dependencies

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

type Items struct {
	Name string `json:"name"`
	ID   int    `json:"itemid"`
}
