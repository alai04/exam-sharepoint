package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/koltyakov/gosip"
	strategy "github.com/koltyakov/gosip-sandbox/strategies/azurecert"
	"github.com/koltyakov/gosip/api"
	finance "github.com/piquette/finance-go"
	"github.com/piquette/finance-go/equity"
)

type Event struct {
	Title       string `json:"Title"`
	EventDate   string `json:"EventDate"`
	EndDate     string `json:"EndDate"`
	Description string `json:"Description"`
	AllDay      bool   `json:"fAllDayEvent"`
}

var (
	tickers []string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file ", err)
	}

	ticker_list := os.Getenv("TICKER_LIST")
	tickers = strings.Split(ticker_list, ",")
	fmt.Println(len(tickers), "Tickers:", tickers)
}

func main() {

	// authCnfg := &strategy.AuthCnfg{
	// 	SiteURL:  os.Getenv("SPAUTH_SITEURL"),
	// 	TenantID: os.Getenv("AZURE_TENANT_ID"),
	// 	ClientID: os.Getenv("AZURE_CLIENT_ID"),
	// 	CertPath: os.Getenv("AZURE_CERTIFICATE_PATH"),
	// 	CertPass: os.Getenv("AZURE_CERTIFICATE_PASSWORD"),
	// }
	// or using `private.json` creds source

	authCnfg := &strategy.AuthCnfg{}
	configPath := "./config/private.json"
	if err := authCnfg.ReadConfig(configPath); err != nil {
		log.Fatalf("unable to get config: %v", err)
	}

	client := &gosip.SPClient{AuthCnfg: authCnfg}
	sp := api.NewSP(client)

	res, err := sp.Web().Select("Title").Get()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Site title: %s\n", res.Data().Title)

	list := sp.Web().GetList("Lists/Earnings Calendar")
	itemsResp, err := list.Items().OrderBy("Id", true).Get()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(string(itemsResp))
	fmt.Printf("There are %d items in Earnings Calendar\n", len(itemsResp.Data()))
	// m := make(map[string]interface{})
	// err = json.Unmarshal(itemsResp, &m)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for k, v := range m {
	// 	fmt.Printf("%v: %v\n", k, v)
	// }
	// for _, item := range itemsResp.Data() {
	// 	itemData := item.Data()
	// 	fmt.Printf("ID: %d, Title: %s\n", itemData.ID, itemData.Title)
	// }

	// fmt.Println("Add a event to calendar...")
	// event := &Event{
	// 	Title:       "MSFT",
	// 	EventDate:   "2023-01-25T00:00:00Z",
	// 	EndDate:     "2023-01-25T23:59:59Z",
	// 	Description: "For test 999",
	// 	AllDay:      true,
	// }
	for _, ticker := range tickers {
		ev, err := getEarningsDate(ticker)
		if err != nil {
			log.Println("getEarningsDate() return error:", err)
			break
		}
		fmt.Println(ev)

		itemPayload, _ := json.Marshal(ev)
		itemAddRes, err := list.Items().Add(itemPayload)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("Raw response: %s\n", itemAddRes)
		fmt.Printf("Added item's ID: %d\n", itemAddRes.Data().ID)
	}
}

func getEarningsDate(ticker string) (ev Event, err error) {
	var q *finance.Equity
	q, err = equity.Get(ticker)
	if err != nil {
		return
	}

	ev.Title = q.Symbol
	ev.Description = q.LongName
	ev.EventDate = timestamp2string(q.EarningsTimestampStart)
	ev.EndDate = timestamp2string(q.EarningsTimestampEnd)
	ev.AllDay = true
	return
}

func timestamp2string(ts int) string {
	tm := time.Unix(int64(ts), 0)
	return tm.Format("2006-01-02T15:04:05Z")
}
