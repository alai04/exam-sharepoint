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
	client  *gosip.SPClient
	sp      *api.SP
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file ", err)
	}

	ticker_list := os.Getenv("TICKER_LIST")
	tickers = strings.Split(ticker_list, ",")
	fmt.Println(len(tickers), "Tickers:", tickers)

	authCnfg := &strategy.AuthCnfg{}
	configPath := "./config/private.json"
	if err := authCnfg.ReadConfig(configPath); err != nil {
		log.Fatalf("unable to get config: %v", err)
	}

	client = &gosip.SPClient{AuthCnfg: authCnfg}
	sp = api.NewSP(client)

	res, err := sp.Web().Select("Title").Get()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Site title: %s\n", res.Data().Title)
}

func main() {
	list := sp.Web().GetList("Lists/Earnings Calendar")
	itemsResp, err := list.Items().OrderBy("Id", true).Get()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("There are %d items in Earnings Calendar\n", len(itemsResp.Data()))
	for _, ticker := range tickers {
		ev, err := getEarningsDate(ticker)
		if err != nil {
			log.Println("getEarningsDate() return error:", err)
			break
		}
		fmt.Println(ev)

		if err = update1Ticker(list, ev); err != nil {
			log.Println(err)
		}
	}
}

func getEarningsDate(ticker string) (ev Event, err error) {
	var q *finance.Equity
	q, err = equity.Get(ticker)
	if err != nil {
		return
	}

	// fmt.Printf("%+v", q)

	ev.Title = q.Symbol
	ev.Description = q.LongName
	ev.EventDate = timestamp2string(q.EarningsTimestamp)
	ev.EndDate = timestamp2string(q.EarningsTimestamp)
	ev.AllDay = true
	return
}

func timestamp2string(ts int) string {
	tm := time.Unix(int64(ts), 0)
	return tm.Format("2006-01-02T15:04:05Z")
}

func update1Ticker(list *api.List, ev Event) error {
	if len(ev.Title) == 0 {
		return nil
	}

	caml := fmt.Sprintf(`
		<View>
			<Query>
				<Where>
					<Eq>
						<FieldRef Name='Title' />
						<Value Type='Text'>%s</Value>
					</Eq>
				</Where>
			</Query>
		</View>
	`, ev.Title)

	itemsResp, err := list.Items().GetByCAML(caml)
	if err != nil {
		return err
	}

	itemPayload, _ := json.Marshal(ev)

	if len(itemsResp.Data()) == 0 {
		fmt.Printf("Adding Title: %s\n", ev.Title)
		_, err := list.Items().Add(itemPayload)
		return err
	}

	for _, item := range itemsResp.Data() {
		itemData := item.Data()
		fmt.Printf("Updating ID: %d, Title: %s\n", itemData.ID, itemData.Title)

		_, err := list.Items().GetByID(itemData.ID).Update(itemPayload)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	return nil
}
