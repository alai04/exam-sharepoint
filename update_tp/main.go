package main

import (
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/koltyakov/gosip"
	strategy "github.com/koltyakov/gosip-sandbox/strategies/azurecert"
	"github.com/koltyakov/gosip/api"
	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/equity"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
)

type Ticker struct {
	Ticker     string `json:"ticker"`
	Model      string `json:"model"`
	Sheet      string `json:"sheet"`
	TPLocation string `json:"tpLocation"`
	name       string
	tp         float64
	cp         float64
}

type TickersConfig struct {
	Tickers []Ticker `json:"tickers"`
}

var (
	tickersConfig TickersConfig
	client        *gosip.SPClient
	sp            *api.SP
)

func init() {
	viper.AddConfigPath("../config")
	viper.SetConfigName("tickers")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error loading tickers config file ", err)
	}

	viper.Unmarshal(&tickersConfig)
	fmt.Println(tickersConfig)

	authCnfg := &strategy.AuthCnfg{}
	configPath := "../config/private.json"
	if err := authCnfg.ReadConfig(configPath); err != nil {
		log.Fatalf("unable to get auth config: %v", err)
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
	for i, ticker := range tickersConfig.Tickers {
		tp, err := getTargetPrice(ticker)
		if err != nil {
			log.Println(err)
			continue
		}

		cp, name, err := getCurrentPriceAndName(ticker.Ticker)
		if err != nil {
			log.Println(err)
			continue
		}

		tickersConfig.Tickers[i].tp = tp
		tickersConfig.Tickers[i].cp = cp
		tickersConfig.Tickers[i].name = name
		fmt.Printf("Get price of %s: $%.2f, $%.2f\n", ticker.Ticker, tp, cp)
	}

	updateSPList(tickersConfig.Tickers)
}

func updateSPList(tickers []Ticker) {
	listName := "Lists/Tracked companies"
	list := sp.Web().GetList(listName)
	itemsResp, err := list.Items().OrderBy("Id", true).Get()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Before update, there are %d items in list\n", len(itemsResp.Data()))

	for _, ticker := range tickers {
		if err = update1Ticker(list, ticker); err != nil {
			log.Println(err)
		}
	}

	itemsResp, err = list.Items().OrderBy("Id", true).Get()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("After update, there are %d items in list\n", len(itemsResp.Data()))
}

func update1Ticker(list *api.List, ticker Ticker) error {
	if len(ticker.name) == 0 {
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
	`, ticker.name)

	itemsResp, err := list.Items().GetByCAML(caml)
	if err != nil {
		return err
	}

	itemPayload := []byte(fmt.Sprintf(`{
		"Title": "%s",
		"field_2": %f,
		"field_3": %f,
		"field_4": %f
	}`, ticker.name, ticker.cp, ticker.tp, ticker.tp/ticker.cp-1.0))

	if len(itemsResp.Data()) == 0 {
		fmt.Printf("Adding Title: %s\n", ticker.name)
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

func getTargetPrice(ticker Ticker) (tp float64, err error) {
	var fileReader io.ReadCloser
	fileReader, err = sp.Web().GetFile(ticker.Model).GetReader()
	if err != nil {
		return
	}
	defer fileReader.Close()

	var f *excelize.File
	f, err = excelize.OpenReader(fileReader)
	if err != nil {
		return
	}
	defer f.Close()

	var cell string
	cell, err = f.GetCellValue(ticker.Sheet, ticker.TPLocation)
	if err != nil {
		return
	}

	tp, err = strconv.ParseFloat(cell, 32)
	return
}

func getCurrentPriceAndName(ticker string) (cp float64, name string, err error) {
	var q *finance.Equity
	q, err = equity.Get(ticker)
	if err != nil {
		return
	}

	// fmt.Printf("%+v\n", q)
	return q.RegularMarketPrice, q.ShortName, nil
}
