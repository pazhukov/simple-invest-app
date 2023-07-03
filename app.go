package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

var config_name = "empty.db"
var pathToConfig string
var database *sql.DB
var exPath string
var comm_broker_tx int

var version = "0.01"

type DB struct {
	name string
	path string
}

type Settings struct {
	WorkDir          string
	Port             string
	BrokerCommInDeal int
	ErrorMsg         string
}

type CustomsList struct {
	Currencies []Currency
	Portfolios []Portfolio
	Assets     []Asset
}

type Currency struct {
	Code     string
	Name     string
	IsMetal  int
	ErrorMsg string
}

type CurrencyList struct {
	ErrorMsg string
	List     []Currency
}

type CurrencyRate struct {
	Id         int
	DateRate   string
	Currency   string
	Rate       float64
	ErrorMsg   string
	Currencies []Currency
}

type Rates struct {
	Currency string
	ErrorMsg string
	List     []CurrencyRate
}

type Portfolio struct {
	Id       int
	Name     string
	Broker   string
	ErrorMsg string
}

type PortfolioList struct {
	ErrorMsg string
	List     []Portfolio
}

type Asset struct {
	Id         int
	Name       string
	Grn_code   string
	ISIN       string
	Type       int
	TypeName   string
	Ticker     string
	Currency   string
	ErrorMsg   string
	Currencies []Currency
}

type AssetList struct {
	ErrorMsg string
	List     []Asset
}

type AssetQuote struct {
	Id         int
	DateQuotes string
	Asset      int
	AssetName  string
	Currency   string
	Open       float64
	Max        float64
	Min        float64
	Close      float64
	AccInt     float64
	ErrorMsg   string
	Assets     []Asset
	Currencies []Currency
}

type Quotes struct {
	Asset     int
	AssetName string
	ErrorMsg  string
	List      []AssetQuote
}

type Deal struct {
	Id                   int
	Number               string
	Date                 string
	Date_ex              string
	Asset                int
	AssetName            string
	Direction            string
	Qty                  float64
	Price                float64
	Price_Currency       string
	Amount               float64
	AccInt               float64
	Broker_Comm          float64
	Broker_Comm_Currency string
	Portfolio            int
	Portfolio_Name       string
	ErrorMsg             string
	Currencies           []Currency
	Portfolios           []Portfolio
	Assets               []Asset
}

type DealList struct {
	ErrorMsg string
	List     []Deal
}

type AssetTx struct {
	Id            int64
	DateTx        string
	Asset         int
	AssetName     string
	Portfolio     int
	PortfolioName string
	Qty           float64
	DealId        int
	Comment       string
	ErrorMsg      string
	Portfolios    []Portfolio
	Assets        []Asset
}

type MoneyTx struct {
	Id            int64
	DateTx        string
	Portfolio     int
	PortfolioName string
	Currency      string
	Amount        float64
	DealId        int
	Comment       string
	ErrorMsg      string
	Currencies    []Currency
	Portfolios    []Portfolio
}

type AssetTxList struct {
	ErrorMsg string
	List     []AssetTx
}

type MoneyTxList struct {
	ErrorMsg string
	List     []MoneyTx
}

type Tx struct {
	Id       int
	Asset    []AssetTx
	Money    []MoneyTx
	ErrorMsg string
}

type AssetBalance struct {
	PortfolioName string
	AssetName     string
	Qty           float64
	Currency      string
	Amount        float64
	MarketValue   float64
	PnL           float64
}

type MoneyBalance struct {
	PortfolioName string
	Currency      string
	Amount        float64
	MarketValue   float64
	PnL           float64
}

type ReportOnDate struct {
	OnDate string
	Asset  []AssetBalance
	Money  []MoneyBalance
}

// funcs

func getCustomsList() CustomsList {

	lists := CustomsList{}

	lists.Currencies = getCurrencyArray()
	lists.Portfolios = getPortfolioArray()
	lists.Assets = getAssetArray()

	return lists
}

// DB funcs - currency

func getCurrencyArray() []Currency {
	rows, err := database.Query("select code, name, metal from currencies")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	list := []Currency{}
	for rows.Next() {
		element := Currency{}
		err := rows.Scan(&element.Code, &element.Name, &element.IsMetal)
		if err != nil {
			log.Println(err)
			continue
		}
		list = append(list, element)
	}

	return list
}

func getCurrencyList() CurrencyList {

	out := CurrencyList{}
	out.List = getCurrencyArray()

	return out
}

func getCurrency(id string) (Currency, int) {

	row := database.QueryRow("select code, name, metal from currencies where code = ?", id)
	element := Currency{}
	err := row.Scan(&element.Code, &element.Name, &element.IsMetal)

	if err != nil {
		return element, 300
	} else {
		return element, 200
	}
}

func addEditCurrency(isUpdate int, code string, name string, metal int) error {

	var error_info error

	if isUpdate == 1 {
		_, err := database.Exec("update currencies set code = ?, name = ?, metal = ? where code = ?", code, name, metal, code)

		if err != nil {
			error_info = err
		}
	} else {
		_, err := database.Exec("insert into currencies (code, name, metal) values (?, ?, ?)", code, name, metal)

		if err != nil {
			error_info = err
		}
	}

	return error_info
}

func deleteCurrency(code string) error {
	var error_info error

	_, err := database.Exec("delete from currencies where code = ?", code)
	if err != nil {
		error_info = err
	}
	return error_info
}

// DB funcs - rates

func getRateList(currency string) []CurrencyRate {

	rows, err := database.Query("select id, date(date_rate), currency, rate from currency_rates where currency = '" + currency + "' order by date_rate desc")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	list := []CurrencyRate{}
	for rows.Next() {
		element := CurrencyRate{}
		err := rows.Scan(&element.Id, &element.DateRate, &element.Currency, &element.Rate)
		if err != nil {
			log.Println(err)
			continue
		}
		list = append(list, element)
	}

	return list
}

func getRates(currency string) Rates {

	rates := Rates{}
	rates.Currency = currency
	rates.List = getRateList(currency)

	return rates
}

func getRate(id int) CurrencyRate {

	row := database.QueryRow("select id, date(date_rate), currency, rate from currency_rates where id = ?", id)
	element := CurrencyRate{}
	err := row.Scan(&element.Id, &element.DateRate, &element.Currency, &element.Rate)

	if err != nil {
		return CurrencyRate{}
	} else {
		return element
	}
}

func addEditRate(id int, rate_date string, curr string, rate float64) error {

	var error_info error

	if id == -1 {

		query := "select id from currency_rates where date_rate = ? and currency = ?"
		rows := database.QueryRow(query, rate_date, curr)
		err := rows.Scan(&id)
		if err != nil {
			_, err = database.Exec("insert into currency_rates (date_rate, currency, rate) values (?, ?, ?)", rate_date, curr, rate)
			if err != nil {
				error_info = err
			}
		}

	}

	if id != -1 {
		_, err := database.Exec("update currency_rates set date_rate = ?,  currency = ?, rate = ? where id = ?", rate_date, curr, rate, id)

		if err != nil {
			error_info = err
		}

	}

	return error_info
}

func deleteRate(id int) error {
	var error_info error

	_, err_del := database.Exec("delete from currency_rates where id = ?", id)
	if err_del != nil {
		error_info = err_del
	}

	return error_info
}

// DB funcs - portfolio

func getPortfolioArray() []Portfolio {

	rows, err := database.Query("select id, name, broker from portfolios")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	list := []Portfolio{}
	for rows.Next() {
		element := Portfolio{}
		err := rows.Scan(&element.Id, &element.Name, &element.Broker)
		if err != nil {
			log.Println(err)
			continue
		}
		list = append(list, element)
	}

	return list
}

func getPortfolioList() PortfolioList {

	out := PortfolioList{}
	out.List = getPortfolioArray()

	return out
}

func getPortfolio(id int) Portfolio {
	row := database.QueryRow("select id, name, broker from portfolios where id = ?", id)
	element := Portfolio{}
	_ = row.Scan(&element.Id, &element.Name, &element.Broker)

	return element
}

func addEditPortfolio(id int, name string, broker string) error {
	var error_info error

	if id == -1 {

		_, err := database.Exec("insert into portfolios (name, broker) values (?, ?)", name, broker)
		if err != nil {
			error_info = err
		}

	} else {

		_, err := database.Exec("update portfolios set name = ?, broker = ? where id = ?", name, broker, id)

		if err != nil {
			error_info = err
		}

	}

	return error_info
}

func deletePortfolio(id int) error {
	var error_info error

	_, err_del := database.Exec("delete from portfolios where id = ?", id)
	if err_del != nil {
		error_info = err_del
	}

	return error_info
}

// DB funcs - asset

func getAssetArray() []Asset {
	rows, err := database.Query("select id, name, grn_code, isin, type, case when type = 0 then 'share' when type = 1 then 'bond' when type = 2 then 'unit' when type = 3 then 'DR' end as type_name, ticker, curr from assets")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	list := []Asset{}
	for rows.Next() {
		element := Asset{}
		err := rows.Scan(&element.Id, &element.Name, &element.Grn_code, &element.ISIN, &element.Type, &element.TypeName, &element.Ticker, &element.Currency)
		if err != nil {
			log.Println(err)
			continue
		}
		list = append(list, element)
	}

	return list
}

func getAssetList() AssetList {
	out := AssetList{}
	out.List = getAssetArray()

	return out
}

func getAsset(id int) Asset {

	row := database.QueryRow("select id, name, grn_code, isin, type, case when type = 0 then 'share' when type = 1 then 'bond' when type = 2 then 'unit' when type = 3 then 'DR' end as type_name, ticker, curr from assets where id = ?", id)
	element := Asset{}
	_ = row.Scan(&element.Id, &element.Name, &element.Grn_code, &element.ISIN, &element.Type, &element.TypeName, &element.Ticker, &element.Currency)

	return element
}

func addEditAsset(id int, name string, grn_code string, isin string, type_asset int, ticker string, curr string) error {
	var error_info error

	if id == -1 {

		_, err := database.Exec("insert into assets (name, grn_code, isin, type, ticker, curr) values (?, ?, ?, ?, ?, ?)", name, grn_code, isin, type_asset, ticker, curr)
		if err != nil {
			error_info = err
		}

	} else {

		_, err := database.Exec("update assets set name = ?, grn_code = ?, isin = ?, type = ?, ticker = ?, curr = ? where id = ?", name, grn_code, isin, type_asset, ticker, curr, id)

		if err != nil {
			error_info = err
		}

	}

	return error_info
}

func deleteAsset(id int) error {
	var error_info error

	_, err_del := database.Exec("delete from assets where id = ?", id)
	if err_del != nil {
		error_info = err_del
	}

	return error_info
}

// DB funcs - quotes

func getAssetQuotes(asset_id int) Quotes {

	quotes := Quotes{}

	asset := getAsset(asset_id)
	quotes.Asset = asset.Id
	quotes.AssetName = asset.Name
	quotes.List = getQuotes(asset_id)

	return quotes
}

func getQuotes(asset_id int) []AssetQuote {

	query := "select id, date(date_quote), asset, currency, open, max, min, close, accint from asset_quotes where asset = " + strconv.Itoa(asset_id) + " order by date_quote desc"

	rows, err := database.Query(query)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	list := []AssetQuote{}
	for rows.Next() {
		element := AssetQuote{}
		err := rows.Scan(&element.Id, &element.DateQuotes, &element.Asset, &element.Currency, &element.Open, &element.Max, &element.Min, &element.Close, &element.AccInt)
		if err != nil {
			log.Println(err)
			continue
		}
		list = append(list, element)
	}

	return list
}

func getQuote(id int) AssetQuote {

	query := "select q.id, date(q.date_quote), q.asset, a.name, q.currency, q.open, q.max, q.min, q.close, q.accint from asset_quotes q left join assets as a on a.id = q.asset where q.id = " + strconv.Itoa(id)

	rows, err := database.Query(query)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	element := AssetQuote{}
	for rows.Next() {
		err := rows.Scan(&element.Id, &element.DateQuotes, &element.Asset, &element.AssetName, &element.Currency, &element.Open, &element.Max, &element.Min, &element.Close, &element.AccInt)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	return element
}

func addEditQuote(id int, asset int, quote_date string, curr string, open float64, max float64, min float64, close float64, accint float64) error {
	var error_info error

	if id == -1 {

		query := "select id from asset_quotes where date_quote = ? and asset = ? and currency = ?"
		rows := database.QueryRow(query, quote_date, asset, curr)
		err := rows.Scan(&id)
		if err != nil {
			_, err = database.Exec("insert into asset_quotes (date_quote, asset, currency, open, max, min, close, accint) values (?, ?, ?, ?, ?, ?, ?, ?)", quote_date, asset, curr, open, max, min, close, accint)
			if err != nil {
				error_info = err
			}
		}

	}

	if id != -1 {
		_, err := database.Exec("update asset_quotes set date_quote = ?, asset = ?, currency = ?, open = ?, max = ?, min = ?, close = ?, accint = ? where id = ?", quote_date, asset, curr, open, max, min, close, accint, id)

		if err != nil {
			error_info = err
		}

	}

	return error_info
}

func deleteQuote(id int) error {
	var error_info error
	_, err_del := database.Exec("delete from asset_quotes where id = ?", id)
	if err_del != nil {
		error_info = err_del
	}

	return error_info
}

// DB funcs - deals

func getDealArray() []Deal {
	query := "select d.id, d.number, date(d.date_deal), date(d.date_exec), a.name as asset, case when d.direction = 'B' then 'BUY' when d.direction='S' then 'SELL' end as durection, d.qty, d.price_curr, d.amount, d.broker_comm_curr, d.broker_comm, p.name, accint from deals as d" +
		" left join assets as a on d.asset = a.id" +
		" left join portfolios as p on d.portfolio = p.id"

	rows, err := database.Query(query)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	list := []Deal{}
	for rows.Next() {
		element := Deal{}
		err := rows.Scan(&element.Id, &element.Number, &element.Date, &element.Date_ex, &element.AssetName, &element.Direction, &element.Qty, &element.Price_Currency, &element.Amount, &element.Broker_Comm_Currency, &element.Broker_Comm, &element.Portfolio_Name, &element.AccInt)
		if err != nil {
			log.Println(err)
			continue
		}
		list = append(list, element)
	}

	return list
}

func getDealsList() DealList {
	out := DealList{}
	out.ErrorMsg = ""
	out.List = getDealArray()

	return out
}

func getDeal(id int) (Deal, error) {

	var error_info error

	row := database.QueryRow("select id, number, date(date_deal), date(date_exec), asset, direction, qty, price_curr, price, amount, broker_comm_curr, broker_comm, portfolio, accint from deals where id = ?", id)
	element := Deal{}
	error_info = row.Scan(&element.Id, &element.Number, &element.Date, &element.Date_ex, &element.Asset, &element.Direction, &element.Qty, &element.Price_Currency, &element.Price, &element.Amount, &element.Broker_Comm_Currency, &element.Broker_Comm, &element.Portfolio, &element.AccInt)

	lists := getCustomsList()
	element.Currencies = lists.Currencies
	element.Portfolios = lists.Portfolios
	element.Assets = lists.Assets

	return element, error_info
}

func addEditDeal(id int, number string, date_deal string, date_exec string, asset int, direction string, qty float64, price_curr string, price float64, amount float64, broker_comm_curr string, broker_comm float64, portfolio int, accint float64) error {
	var error_info error

	if id == -1 {
		_, err := database.Exec("insert into deals (number, date_deal, date_exec, asset, direction, qty, price_curr, price, amount, broker_comm_curr, broker_comm, portfolio, accint) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", number, date_deal, date_exec, asset, direction, qty, price_curr, price, amount, broker_comm_curr, broker_comm, portfolio, accint)
		if err != nil {
			error_info = err
		}
	} else {
		_, err := database.Exec("update deals set number = ?, date_deal = ?, date_exec = ?, asset = ?, direction = ?, qty = ?, price_curr = ?, price = ?, amount = ?, broker_comm_curr = ?, broker_comm = ?, portfolio = ?, accint = ? where id = ?", number, date_deal, date_exec, asset, direction, qty, price_curr, price, amount, broker_comm_curr, broker_comm, portfolio, accint, id)

		if err != nil {
			error_info = err
		}
	}
	return error_info
}

func deleteDeal(id int) error {
	var error_info error

	_, err_del := database.Exec("delete from deals where id = ?", id)
	if err_del != nil {
		error_info = err_del
	} else {
		error_info = deleteTxDeal(id)
	}
	return error_info
}

// DB funcs - deal tx

func deleteTxDeal(id int) error {
	var error_info error

	_, err_del := database.Exec("delete from asset_transactions where deal_id = ?", id)
	if err_del != nil {
		error_info = err_del
	}

	_, err_del_money := database.Exec("delete from money_transactions where deal_id = ?", id)
	if err_del_money != nil {
		error_info = err_del_money
	}

	return error_info
}

func createTxDeal(id int) error {

	var error_info error

	deal, error_info := getDeal(id)
	if error_info != nil {
		return error_info
	}

	direction := 1
	if deal.Direction == "S" {
		direction = -1
	}

	_, err := database.Exec("insert into asset_transactions (date_tx, asset, portfolio, qty, deal_id, comment) values (?, ?, ?, ?, ?, ?)", deal.Date_ex, deal.Asset, deal.Portfolio, deal.Qty*float64(direction), id, "Asset move by Deal")
	if err != nil {
		error_info = err
	}

	_, err = database.Exec("insert into money_transactions (date_tx, curr, portfolio, amount, deal_id, comment) values (?, ?, ?, ?, ?, ?)", deal.Date_ex, deal.Price_Currency, deal.Portfolio, deal.Amount*float64(direction)*-1, id, "Money move by Deal")
	if err != nil {
		error_info = err
	}

	if error_info == nil && deal.Broker_Comm > 0 && comm_broker_tx == 1 {
		_, err = database.Exec("insert into money_transactions (date_tx, curr, portfolio, amount, deal_id, comment) values (?, ?, ?, ?, ?, ?)", deal.Date_ex, deal.Price_Currency, deal.Portfolio, deal.Broker_Comm*-1, id, "Money move by broker commission")
		if err != nil {
			error_info = err
		}
	}

	if error_info == nil && deal.AccInt > 0 {
		_, err = database.Exec("insert into money_transactions (date_tx, curr, portfolio, amount, deal_id, comment) values (?, ?, ?, ?, ?, ?)", deal.Date_ex, deal.Price_Currency, deal.Portfolio, deal.AccInt*float64(direction)*-1, id, "Money move by acc int")
		if err != nil {
			error_info = err
		}
	}

	return error_info
}

func getTxDeal(id int) (Tx, error) {

	var error_info error

	deal_tx := Tx{}
	deal_tx.Id = id

	asset := []AssetTx{}
	money := []MoneyTx{}

	query_asset := "select d.id, date(d.date_tx), d.asset, a.name as asset_name, d.portfolio, p.name, d.qty, d.deal_id, d.comment from asset_transactions as d" +
		" left join assets as a on d.asset = a.id" +
		" left join portfolios as p on d.portfolio = p.id" +
		" where d.deal_id = " + strconv.Itoa(id)

	rows_a, err_a := database.Query(query_asset)
	if err_a != nil {
		error_info = err_a
	}
	defer rows_a.Close()

	for rows_a.Next() {
		element := AssetTx{}
		err := rows_a.Scan(&element.Id, &element.DateTx, &element.Asset, &element.AssetName, &element.Portfolio, &element.PortfolioName, &element.Qty, &element.DealId, &element.Comment)
		if err != nil {
			error_info = err
			continue
		}
		asset = append(asset, element)
	}

	query_money := "select d.id, date(d.date_tx), d.portfolio, p.name, d.curr, d.amount, d.deal_id, d.comment from money_transactions as d" +
		" left join portfolios as p on d.portfolio = p.id" +
		" where d.deal_id = " + strconv.Itoa(id)

	rows_m, err_m := database.Query(query_money)
	if err_m != nil {
		error_info = err_m
	}
	defer rows_m.Close()

	for rows_m.Next() {
		element := MoneyTx{}
		err := rows_m.Scan(&element.Id, &element.DateTx, &element.Portfolio, &element.PortfolioName, &element.Currency, &element.Amount, &element.DealId, &element.Comment)
		if err != nil {
			error_info = err
			continue
		}
		money = append(money, element)
	}

	deal_tx.Asset = asset
	deal_tx.Money = money

	return deal_tx, error_info
}

// DB funcs - non trade

func getMoneyInOutArray() []MoneyTx {
	money := []MoneyTx{}

	query_money := "select d.id, date(d.date_tx), d.portfolio, p.name, d.curr, d.amount, d.comment from money_transactions as d" +
		" left join portfolios as p on d.portfolio = p.id" +
		" where d.nontrade = 1" +
		" order by d.date_tx desc"

	rows_m, err_m := database.Query(query_money)
	if err_m != nil {
		log.Println(err_m)
	}
	defer rows_m.Close()

	for rows_m.Next() {
		element := MoneyTx{}
		err := rows_m.Scan(&element.Id, &element.DateTx, &element.Portfolio, &element.PortfolioName, &element.Currency, &element.Amount, &element.Comment)
		if err != nil {
			log.Println(err)
			continue
		}
		money = append(money, element)
	}

	return money
}

func getAssetInOutArray() []AssetTx {
	asset := []AssetTx{}

	query_asset := "select d.id, date(d.date_tx), d.asset, a.name as asset_name, d.portfolio, p.name, d.qty, d.comment from asset_transactions as d" +
		" left join assets as a on d.asset = a.id" +
		" left join portfolios as p on d.portfolio = p.id" +
		" where d.nontrade = 1" +
		" order by d.date_tx desc"

	rows_a, err_a := database.Query(query_asset)
	if err_a != nil {
		log.Println(err_a)
	}
	defer rows_a.Close()

	for rows_a.Next() {
		element := AssetTx{}
		err := rows_a.Scan(&element.Id, &element.DateTx, &element.Asset, &element.AssetName, &element.Portfolio, &element.PortfolioName, &element.Qty, &element.Comment)
		if err != nil {
			log.Println(err)
			continue
		}
		asset = append(asset, element)
	}

	return asset
}

func getMoneyInOutList() MoneyTxList {
	out := MoneyTxList{}
	out.ErrorMsg = ""
	out.List = getMoneyInOutArray()

	return out
}

func getAssetInOutList() AssetTxList {
	out := AssetTxList{}
	out.ErrorMsg = ""
	out.List = getAssetInOutArray()

	return out
}

func addEditMoneyInOut(id int64, date_exec string, amount float64, curr string, portfolio int, comment string) error {
	var error_info error

	if id == -1 {

		_, err := database.Exec("insert into money_transactions(date_tx, portfolio, amount, curr, nontrade, comment) values (?, ?, ?, ?, ?, ?)", date_exec, portfolio, amount, curr, 1, comment)
		if err != nil {
			error_info = err
		}

	} else {

		_, err := database.Exec("update money_transactions set date_tx = ?, portfolio = ?, amount = ?, curr = ?, comment = ? where id = ?", date_exec, portfolio, amount, curr, comment, id)

		if err != nil {
			error_info = err
		}
	}

	return error_info
}

func addEditAssetInOut(id int64, date_exec string, qty float64, asset int, portfolio int, comment string) error {
	var error_info error

	if id == -1 {

		_, err := database.Exec("insert into asset_transactions(date_tx, portfolio, qty, asset, nontrade, comment) values (?, ?, ?, ?, ?, ?)", date_exec, portfolio, qty, asset, 1, comment)
		if err != nil {
			error_info = err
		}

	} else {

		_, err := database.Exec("update asset_transactions set date_tx = ?, portfolio = ?, qty = ?, asset = ?, comment = ? where id = ?", date_exec, portfolio, qty, asset, comment, id)

		if err != nil {
			error_info = err
		}
	}

	return error_info
}

func getMoneyInOut(id int64) (MoneyTx, error) {
	var error_info error

	row := database.QueryRow("select d.id, date(d.date_tx), d.portfolio, d.curr, d.amount, d.comment from money_transactions as d where d.nontrade = 1 and d.id = ?", id)
	element := MoneyTx{}
	error_info = row.Scan(&element.Id, &element.DateTx, &element.Portfolio, &element.Currency, &element.Amount, &element.Comment)

	return element, error_info
}

func getAssetInOut(id int64) (AssetTx, error) {
	var error_info error

	row := database.QueryRow("select d.id, date(d.date_tx), d.portfolio, d.asset, d.qty, d.comment from asset_transactions as d where d.nontrade = 1 and d.id = ?", id)
	element := AssetTx{}
	error_info = row.Scan(&element.Id, &element.DateTx, &element.Portfolio, &element.Asset, &element.Qty, &element.Comment)

	return element, error_info
}

func deleteMoneyInOut(id int64) error {
	var error_info error

	_, err_del := database.Exec("delete from money_transactions where id = ?", id)
	if err_del != nil {
		error_info = err_del
	}

	return error_info
}

func deleteAssetInOut(id int64) error {
	var error_info error

	_, err_del := database.Exec("delete from asset_transactions where id = ?", id)
	if err_del != nil {
		error_info = err_del
	}

	return error_info
}

// DB funcs - reports

func getAssetBalance(on_date string) []AssetBalance {

	balance := []AssetBalance{}

	query := "select p.name, a.name, sum(tx.qty), ifnull(money.curr, a.curr), ifnull(sum(money.amount) *-1, 0), IFNULL(quotes.close, 0) *sum(tx.qty) as market_value, ifnull(round(IFNULL(quotes.close, 0) *sum(tx.qty) + sum(money.amount),2), 0) as pnl from asset_transactions as tx  " +
		" left join assets as a on tx.asset = a.id" +
		" left join portfolios as p on tx.portfolio = p.id" +
		" left join (select deal_id, curr, sum(amount) as amount from money_transactions group by deal_id, curr) as money on tx.deal_id = money.deal_id" +
		" left join (select max(date_quote) as d_rate, asset a_rate from asset_quotes where date_quote <= '" + on_date + "' group by asset) as max_rates on tx.asset = max_rates.a_rate" +
		" left join asset_quotes as quotes on max_rates.a_rate = quotes.asset and max_rates.d_rate = quotes.date_quote" +
		" where tx.date_tx <=  '" + on_date + "'" +
		" group by p.name, a.name, money.curr"

	rows_a, err_a := database.Query(query)
	if err_a != nil {
		log.Println(err_a)
	}
	defer rows_a.Close()

	for rows_a.Next() {
		element := AssetBalance{}
		err := rows_a.Scan(&element.PortfolioName, &element.AssetName, &element.Qty, &element.Currency, &element.Amount, &element.MarketValue, &element.PnL)
		if err != nil {
			log.Println(err)
			continue
		}
		balance = append(balance, element)
	}

	return balance
}

func getMoneyBalance(on_date string) []MoneyBalance {

	balance := []MoneyBalance{}

	query := "select p.name, sum(tx.amount), tx.curr, sum(tx.amount)*ifnull(last_rates.rate, 1) as market_value, sum(tx.amount)*ifnull(last_rates.rate, 1)-sum(tx.amount) as pnl from money_transactions as tx  " +
		" left join portfolios as p on tx.portfolio = p.id" +
		" left join (select max(date_rate) as date_rate, currency from currency_rates where date_rate <= '" + on_date + "' group by currency) as rates on tx.curr = rates.currency" +
		" left join currency_rates as last_rates on rates.currency = last_rates.currency and rates.date_rate = last_rates.date_rate" +
		" where tx.date_tx <=  '" + on_date + "'" +
		" group by p.name, tx.curr"

	rows_a, err_a := database.Query(query)
	if err_a != nil {
		log.Println(err_a)
	}
	defer rows_a.Close()

	for rows_a.Next() {
		element := MoneyBalance{}
		err := rows_a.Scan(&element.PortfolioName, &element.Amount, &element.Currency, &element.MarketValue, &element.PnL)
		if err != nil {
			log.Println(err)
			continue
		}
		balance = append(balance, element)
	}

	return balance
}

func getReportOnDate(on_date string) ReportOnDate {

	report := ReportOnDate{}
	report.OnDate = on_date
	report.Asset = getAssetBalance(on_date)
	report.Money = getMoneyBalance(on_date)

	return report
}

// DB funcs - settings

func getSettings() (Settings, error) {

	row := database.QueryRow("select workdir, port, comm_in_deal from settings where id = 0")
	element := Settings{}
	err := row.Scan(&element.WorkDir, &element.Port, &element.BrokerCommInDeal)

	if err != nil {
		return element, err
	} else {
		return element, nil
	}
}

func saveSettings(wokr_dir string, port string, comm_in_deal int) error {
	var error_info error

	_, err := database.Exec("update settings set workdir = ?, port = ?, comm_in_deal = ? where id = 0", wokr_dir, port, comm_in_deal)

	if err != nil {
		error_info = err
	}

	return error_info
}

// page funcs

func indexHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	var source_path = filepath.Join(exPath, "gui", "index.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, nil)
}

func catalogsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	var source_path = filepath.Join(exPath, "gui", "catalogs.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, nil)
}

// currency func

func currenciesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	list := getCurrencyList()

	source_path := filepath.Join(exPath, "gui", "catalogs_currencies.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, list)
}

func curAddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		error_msg := ""

		cur_code := r.FormValue("Cur_code")
		cur_name := r.FormValue("Cur_name")
		is_metal := r.FormValue("Is_metal")

		metal, err := strconv.Atoi(is_metal)
		if err != nil {
			metal = 0
			log.Println("Can't convert IS METAL parameter")
		}

		if cur_code == "" {
			error_msg = error_msg + "'Code' can't be empty;"
		}

		if cur_name == "" {
			error_msg = error_msg + "'Name' can't be empty;"
		}

		if error_msg != "" {
			error_msg = "ERROR: " + error_msg

			element := Currency{}
			element.Code = cur_code
			element.Name = cur_name
			element.IsMetal = metal
			element.ErrorMsg = error_msg

			var source_path = filepath.Join(exPath, "gui", "currency_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		} else {
			error_info := addEditCurrency(0, cur_code, cur_name, metal)

			if error_info == nil {
				http.Redirect(w, r, "/currencies", http.StatusMovedPermanently)
			} else {
				list := getCurrencyList()
				list.ErrorMsg = error_info.Error()

				source_path := filepath.Join(exPath, "gui", "catalogs_currencies.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, list)
			}
		}
	} else {
		var source_path = filepath.Join(exPath, "gui", "currency_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, nil)
	}
}

func curEditHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		error_msg := ""

		cur_code := r.FormValue("Cur_code")
		cur_name := r.FormValue("Cur_name")
		is_metal := r.FormValue("Is_metal")

		metal, err := strconv.Atoi(is_metal)
		if err != nil {
			metal = 0
			log.Println("Can't convert IS METAL parameter")
		}

		if cur_code == "" {
			error_msg = error_msg + "'Code' can't be empty;"
		}

		if cur_name == "" {
			error_msg = error_msg + "'Name' can't be empty;"
		}

		if error_msg != "" {
			error_msg = "ERROR: " + error_msg

			element := Currency{}
			element.Code = cur_code
			element.Name = cur_name
			element.IsMetal = metal
			element.ErrorMsg = error_msg

			var source_path = filepath.Join(exPath, "gui", "currency_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		} else {
			error_info := addEditCurrency(1, cur_code, cur_name, metal)

			if error_info == nil {
				http.Redirect(w, r, "/currencies", http.StatusMovedPermanently)
			} else {
				list := getCurrencyList()
				list.ErrorMsg = error_info.Error()

				source_path := filepath.Join(exPath, "gui", "catalogs_currencies.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, list)
			}
		}
	} else {

		vars := mux.Vars(r)
		id := vars["id"]

		element, err := getCurrency(id)
		if err == 0 {
			log.Println(err)
			http.Error(w, http.StatusText(404), http.StatusNotFound)
		} else {
			var source_path = filepath.Join(exPath, "gui", "currency_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}
	}
}

func curDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	error_info := deleteCurrency(p_id)
	if error_info == nil {
		http.Redirect(w, r, "/currencies", http.StatusMovedPermanently)
	} else {
		w.Header().Set("Cache-Control", "no-cache")

		list := getCurrencyList()
		list.ErrorMsg = error_info.Error()

		source_path := filepath.Join(exPath, "gui", "catalogs_currencies.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, list)
	}
}

// rates func

func ratesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	quotes := getRates(p_id)

	var source_path = filepath.Join(exPath, "gui", "rates.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, quotes)
}

func rateAddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		cur_code := r.FormValue("Rate_curr")
		date_rate := r.FormValue("Rate_date")
		p_value := r.FormValue("Rate_value")

		value, err := strconv.ParseFloat(p_value, 64)
		if err != nil {
			value = 0
			log.Println("Can't convert RATE VALUE parameter")
		}

		if cur_code == "" {
			error_msg = error_msg + "'Currency' can't be empty;"
		}

		if date_rate == "" {
			error_msg = error_msg + "'Date' can't be empty;"
		}

		if value == 0 {
			error_msg = error_msg + "'Rate' can't be empty;"
		}

		if error_msg == "" {

			error_info := addEditRate(-1, date_rate, cur_code, value)
			if error_info == nil {
				http.Redirect(w, r, "/rates/"+cur_code, http.StatusMovedPermanently)
			} else {

				rates := getRates(cur_code)
				rates.ErrorMsg = error_info.Error()

				var source_path = filepath.Join(exPath, "gui", "rates.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, rates)
			}
		} else {
			error_msg = "ERROR: " + error_msg

			element := CurrencyRate{}
			element.Currency = cur_code
			element.DateRate = date_rate
			element.Rate = value
			element.Currencies = getCurrencyArray()
			element.ErrorMsg = error_msg

			var source_path = filepath.Join(exPath, "gui", "rate_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}
	} else {

		vars := mux.Vars(r)
		p_id := vars["id"]

		rate := CurrencyRate{}
		rate.Currency = p_id
		rate.Currencies = getCurrencyArray()

		var source_path = filepath.Join(exPath, "gui", "rate_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, rate)

	}
}

func rateEditHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		p_id := r.FormValue("Rate_id")
		cur_code := r.FormValue("Rate_curr")
		date_rate := r.FormValue("Rate_date")
		p_value := r.FormValue("Rate_value")

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		value, err := strconv.ParseFloat(p_value, 64)
		if err != nil {
			value = 0
			log.Println("Can't convert RATE VALUE parameter")
		}

		if id == 0 {
			error_msg = error_msg + "'ID' can't be empty;"
		}

		if cur_code == "" {
			error_msg = error_msg + "'Currency' can't be empty;"
		}

		if date_rate == "" {
			error_msg = error_msg + "'Date' can't be empty;"
		}

		if value == 0 {
			error_msg = error_msg + "'Rate' can't be empty;"
		}

		if error_msg == "" {
			error_msg = "ERROR: " + error_msg

			error_info := addEditRate(id, date_rate, cur_code, value)
			if error_info == nil {
				http.Redirect(w, r, "/rates/"+cur_code, http.StatusMovedPermanently)
			} else {

				rates := getRates(cur_code)
				rates.ErrorMsg = error_info.Error()

				var source_path = filepath.Join(exPath, "gui", "rates.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, rates)
			}
		} else {
			element := CurrencyRate{}
			element.Id = id
			element.Currency = cur_code
			element.DateRate = date_rate
			element.Rate = value
			element.Currencies = getCurrencyArray()
			element.ErrorMsg = error_msg

			var source_path = filepath.Join(exPath, "gui", "rate_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}

	} else {

		vars := mux.Vars(r)
		p_id := vars["id"]

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		rate := getRate(id)
		rate.Currencies = getCurrencyArray()

		var source_path = filepath.Join(exPath, "gui", "rate_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, rate)

	}
}

func rateDeleteHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.Atoi(p_id)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	error_info := deleteRate(id)
	if error_info == nil {
		http.Redirect(w, r, "/currencies", http.StatusMovedPermanently)
	} else {
		w.Header().Set("Cache-Control", "no-cache")

		list := getCurrencyList()
		list.ErrorMsg = error_info.Error()

		source_path := filepath.Join(exPath, "gui", "catalogs_currencies.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, list)
	}
}

// portfolio func

func portfoliosHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	var list = getPortfolioList()

	var source_path = filepath.Join(exPath, "gui", "catalogs_portfolios.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, list)
}

func portfolioAddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		name := r.FormValue("Portfolio_name")
		broker := r.FormValue("Portfolio_broker")

		if name == "" {
			error_msg = error_msg + "'Name' can't be empty;"
		}

		if broker == "" {
			error_msg = error_msg + "'Broker' can't be empty;"
		}

		if error_msg == "" {

			error_info := addEditPortfolio(-1, name, broker)
			if error_info == nil {
				http.Redirect(w, r, "/portfolios", http.StatusMovedPermanently)
			} else {

				portfolios := getPortfolioList()
				portfolios.ErrorMsg = error_info.Error()

				var source_path = filepath.Join(exPath, "gui", "catalogs_portfolios.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, portfolios)
			}
		} else {
			error_msg = "ERROR: " + error_msg

			element := Portfolio{}
			element.Name = name
			element.Broker = broker
			element.ErrorMsg = error_msg

			var source_path = filepath.Join(exPath, "gui", "portfolio_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)

		}
	} else {
		element := Portfolio{}
		var source_path = filepath.Join(exPath, "gui", "portfolio_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, element)
	}
}

func portfolioEditHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		p_id := r.FormValue("Portfolio_id")
		name := r.FormValue("Portfolio_name")
		broker := r.FormValue("Portfolio_broker")

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert IS ID parameter")
		}

		if name == "" {
			error_msg = error_msg + "'Name' can't be empty;"
		}

		if broker == "" {
			error_msg = error_msg + "'Broker' can't be empty;"
		}

		if id == -1 {
			error_msg = error_msg + "'ID' can't be empty;"
		}

		if error_msg == "" {
			error_info := addEditPortfolio(id, name, broker)
			if error_info == nil {
				http.Redirect(w, r, "/portfolios", http.StatusMovedPermanently)
			} else {
				portfolios := getPortfolioList()
				portfolios.ErrorMsg = error_info.Error()

				var source_path = filepath.Join(exPath, "gui", "catalogs_portfolios.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, portfolios)
			}
		} else {
			error_msg = "ERROR: " + error_msg

			element := Portfolio{}
			element.Id = id
			element.Name = name
			element.Broker = broker
			element.ErrorMsg = error_msg

			source_path := filepath.Join(exPath, "gui", "portfolio_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}
	} else {

		vars := mux.Vars(r)
		p_id := vars["id"]

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert IS ID parameter")
		}

		element := getPortfolio(id)

		source_path := filepath.Join(exPath, "gui", "portfolio_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, element)
	}
}

func portfolioDeleteHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.Atoi(p_id)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	error_info := deletePortfolio(id)
	if error_info == nil {
		http.Redirect(w, r, "/portfolios", http.StatusMovedPermanently)
	} else {
		w.Header().Set("Cache-Control", "no-cache")

		list := getPortfolioList()
		list.ErrorMsg = error_info.Error()

		source_path := filepath.Join(exPath, "gui", "catalogs_portfolios.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, list)
	}
}

// assets

func assetsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	var list = getAssetList()

	var source_path = filepath.Join(exPath, "gui", "catalogs_assets.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, list)
}

func assetAddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		name := r.FormValue("Asset_name")
		grn_code := r.FormValue("Asset_grn_code")
		isin := r.FormValue("Asset_isin")
		p_type := r.FormValue("Asset_type")
		ticker := r.FormValue("Asset_ticker")
		curr := r.FormValue("Asset_currency")

		type_asset, err_type := strconv.Atoi(p_type)
		if err_type != nil {
			type_asset = -1
			log.Println("Can't convert IS TYPE ID parameter")
		}

		if name == "" {
			error_msg = error_msg + "'Name' can't be empty;"
		}

		if grn_code == "" {
			error_msg = error_msg + "'Gov Reg Number' can't be empty;"
		}

		if isin == "" {
			error_msg = error_msg + "'ISIN' can't be empty;"
		}

		if type_asset == -1 {
			error_msg = error_msg + "'Type' can't be empty;"
		}

		if ticker == "" {
			error_msg = error_msg + "'Ticker' can't be empty;"
		}

		if curr == "" {
			error_msg = error_msg + "'Currency' can't be empty;"
		}

		if error_msg == "" {
			error_info := addEditAsset(-1, name, grn_code, isin, type_asset, ticker, curr)
			if error_info == nil {
				http.Redirect(w, r, "/assets", http.StatusMovedPermanently)
			} else {

				assets := getAssetList()
				assets.ErrorMsg = error_info.Error()

				var source_path = filepath.Join(exPath, "gui", "catalogs_assets.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, assets)
			}
		} else {
			error_msg = "ERROR: " + error_msg

			element := Asset{}
			element.Currencies = getCurrencyArray()
			element.ErrorMsg = error_msg
			element.Name = name
			element.Grn_code = grn_code
			element.ISIN = isin
			element.Type = type_asset
			element.Ticker = ticker
			element.Currency = curr

			source_path := filepath.Join(exPath, "gui", "asset_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}
	} else {
		element := Asset{}
		element.Currencies = getCurrencyArray()

		source_path := filepath.Join(exPath, "gui", "asset_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, element)
	}
}

func assetEditHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		p_id := r.FormValue("Asset_id")
		name := r.FormValue("Asset_name")
		grn_code := r.FormValue("Asset_grn_code")
		isin := r.FormValue("Asset_isin")
		p_type := r.FormValue("Asset_type")
		ticker := r.FormValue("Asset_ticker")
		curr := r.FormValue("Asset_currency")

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert IS ID parameter")
		}

		type_asset, err_type := strconv.Atoi(p_type)
		if err_type != nil {
			type_asset = -1
			log.Println("Can't convert IS TYPE ID parameter")
		}

		if id == -1 {
			error_msg = error_msg + "'Id' can't be empty;"
		}

		if name == "" {
			error_msg = error_msg + "'Name' can't be empty;"
		}

		if grn_code == "" {
			error_msg = error_msg + "'Gov Reg Number' can't be empty;"
		}

		if isin == "" {
			error_msg = error_msg + "'ISIN' can't be empty;"
		}

		if type_asset == -1 {
			error_msg = error_msg + "'Type' can't be empty;"
		}

		if ticker == "" {
			error_msg = error_msg + "'Ticker' can't be empty;"
		}

		if curr == "" {
			error_msg = error_msg + "'Currency' can't be empty;"
		}

		if error_msg == "" {
			error_info := addEditAsset(id, name, grn_code, isin, type_asset, ticker, curr)
			if error_info == nil {
				http.Redirect(w, r, "/assets", http.StatusMovedPermanently)
			} else {

				assets := getAssetList()
				assets.ErrorMsg = error_info.Error()

				var source_path = filepath.Join(exPath, "gui", "catalogs_assets.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, assets)
			}
		} else {
			error_msg = "ERROR: " + error_msg

			element := Asset{}
			element.Id = id
			element.Currencies = getCurrencyArray()
			element.ErrorMsg = error_msg
			element.Name = name
			element.Grn_code = grn_code
			element.ISIN = isin
			element.Type = type_asset
			element.Ticker = ticker
			element.Currency = curr

			source_path := filepath.Join(exPath, "gui", "asset_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}
	} else {

		vars := mux.Vars(r)
		p_id := vars["id"]

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert IS ID parameter")
		}

		element := getAsset(id)
		element.ErrorMsg = ""
		element.Currencies = getCurrencyArray()

		source_path := filepath.Join(exPath, "gui", "asset_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, element)
	}
}

func assetDeleteHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.Atoi(p_id)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	error_info := deleteAsset(id)
	if error_info == nil {
		http.Redirect(w, r, "/assets", http.StatusMovedPermanently)
	} else {
		w.Header().Set("Cache-Control", "no-cache")

		list := getAssetList()
		list.ErrorMsg = error_info.Error()

		source_path := filepath.Join(exPath, "gui", "catalogs_assets.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, list)
	}
}

// quotes

func quotesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.Atoi(p_id)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	quotes := getAssetQuotes(id)

	var source_path = filepath.Join(exPath, "gui", "quotes.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, quotes)
}

func quoteAddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		quote_date := r.FormValue("Quote_date")
		p_asset := r.FormValue("Quote_asset")
		quote_curr := r.FormValue("Quote_curr")
		p_open := r.FormValue("Quote_open")
		p_max := r.FormValue("Quote_max")
		p_min := r.FormValue("Quote_min")
		p_close := r.FormValue("Quote_close")
		p_accint := r.FormValue("Quote_accint")

		asset, err := strconv.Atoi(p_asset)
		if err != nil {
			asset = -1
			log.Println("Can't convert ASSET ID parameter")
		}

		open, err := strconv.ParseFloat(p_open, 64)
		if err != nil {
			open = 0
			log.Println("Can't convert OPEN parameter")
		}

		max, err := strconv.ParseFloat(p_max, 64)
		if err != nil {
			max = 0
			log.Println("Can't convert MAX parameter")
		}

		min, err := strconv.ParseFloat(p_min, 64)
		if err != nil {
			min = 0
			log.Println("Can't convert MIN parameter")
		}

		close, err := strconv.ParseFloat(p_close, 64)
		if err != nil {
			close = 0
			log.Println("Can't convert CLOSE parameter")
		}

		accint, err := strconv.ParseFloat(p_accint, 64)
		if err != nil {
			accint = 0
			log.Println("Can't convert ACCINT parameter")
		}

		if asset == -1 {
			error_msg = error_msg + "'Asset' can't empty;"
		}

		if quote_date == "" {
			error_msg = error_msg + "'Date' can't empty;"
		}

		if quote_curr == "" {
			error_msg = error_msg + "'Currency' can't empty;"
		}

		if open == 0 && close == 0 {
			error_msg = error_msg + "'Open' and 'Close' can't empty;"
		}

		if error_msg == "" {
			error_info := addEditQuote(-1, asset, quote_date, quote_curr, open, max, min, close, accint)
			if error_info == nil {
				http.Redirect(w, r, "/quotes/"+p_asset, http.StatusMovedPermanently)
			} else {
				quotes := getAssetQuotes(asset)
				quotes.ErrorMsg = error_info.Error()

				source_path := filepath.Join(exPath, "gui", "quotes.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, quotes)
			}
		} else {

			element := AssetQuote{}
			element.Asset = asset
			element.Currency = quote_curr
			element.DateQuotes = quote_date
			element.Open = open
			element.Min = min
			element.Max = max
			element.Close = close
			element.AccInt = accint
			element.Currencies = getCurrencyArray()
			element.Assets = getAssetArray()
			element.ErrorMsg = error_msg

			source_path := filepath.Join(exPath, "gui", "quote_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}

	} else {

		vars := mux.Vars(r)
		p_id := vars["id"]

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		quote := AssetQuote{}

		asset := getAsset(id)
		quote.Asset = asset.Id
		quote.AssetName = asset.Name
		quote.Currencies = getCurrencyArray()
		quote.Assets = getAssetArray()

		var source_path = filepath.Join(exPath, "gui", "quote_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, quote)

	}
}

func quoteEditHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		p_id := r.FormValue("Quote_id")
		quote_date := r.FormValue("Quote_date")
		p_asset := r.FormValue("Quote_asset")
		quote_curr := r.FormValue("Quote_curr")
		p_open := r.FormValue("Quote_open")
		p_max := r.FormValue("Quote_max")
		p_min := r.FormValue("Quote_min")
		p_close := r.FormValue("Quote_close")
		p_accint := r.FormValue("Quote_accint")

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		asset, err := strconv.Atoi(p_asset)
		if err != nil {
			asset = -1
			log.Println("Can't convert ASSET ID parameter")
		}

		open, err := strconv.ParseFloat(p_open, 64)
		if err != nil {
			open = 0
			log.Println("Can't convert OPEN parameter")
		}

		max, err := strconv.ParseFloat(p_max, 64)
		if err != nil {
			max = 0
			log.Println("Can't convert MAX parameter")
		}

		min, err := strconv.ParseFloat(p_min, 64)
		if err != nil {
			min = 0
			log.Println("Can't convert MIN parameter")
		}

		close, err := strconv.ParseFloat(p_close, 64)
		if err != nil {
			close = 0
			log.Println("Can't convert CLOSE parameter")
		}

		accint, err := strconv.ParseFloat(p_accint, 64)
		if err != nil {
			accint = 0
			log.Println("Can't convert ACCINT parameter")
		}

		if id == -1 {
			error_msg = error_msg + "'ID' can't empty;"
		}

		if asset == -1 {
			error_msg = error_msg + "'Asset' can't empty;"
		}

		if quote_date == "" {
			error_msg = error_msg + "'Date' can't empty;"
		}

		if quote_curr == "" {
			error_msg = error_msg + "'Currency' can't empty;"
		}

		if open == 0 && close == 0 {
			error_msg = error_msg + "'Open' and 'Close' can't empty;"
		}

		if error_msg == "" {
			error_info := addEditQuote(id, asset, quote_date, quote_curr, open, max, min, close, accint)
			if error_info == nil {
				http.Redirect(w, r, "/quotes/"+p_asset, http.StatusMovedPermanently)
			} else {
				quotes := getAssetQuotes(asset)
				quotes.ErrorMsg = error_info.Error()

				source_path := filepath.Join(exPath, "gui", "quotes.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, quotes)
			}
		} else {
			element := AssetQuote{}
			element.Id = id
			element.Asset = asset
			element.Currency = quote_curr
			element.DateQuotes = quote_date
			element.Open = open
			element.Min = min
			element.Max = max
			element.Close = close
			element.AccInt = accint
			element.Currencies = getCurrencyArray()
			element.Assets = getAssetArray()
			element.ErrorMsg = error_msg

			source_path := filepath.Join(exPath, "gui", "quote_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}
	} else {

		vars := mux.Vars(r)
		p_id := vars["id"]

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		quote := getQuote(id)
		quote.Currencies = getCurrencyArray()
		quote.Assets = getAssetArray()

		var source_path = filepath.Join(exPath, "gui", "quote_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, quote)
	}
}

func quoteDeleteHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.Atoi(p_id)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	error_info := deleteQuote(id)
	if error_info == nil {
		http.Redirect(w, r, "/assets", http.StatusMovedPermanently)
	} else {
		w.Header().Set("Cache-Control", "no-cache")

		list := getAssetList()
		list.ErrorMsg = error_info.Error()

		source_path := filepath.Join(exPath, "gui", "catalogs_assets.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, list)
	}
}

// deals

func dealsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	list := getDealsList()

	source_path := filepath.Join(exPath, "gui", "deals.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, list)
}

func dealEditHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		p_id := r.FormValue("Trade_id")
		number := r.FormValue("Trade_number")
		date_deal := r.FormValue("Trade_date")
		date_exec := r.FormValue("Trade_date_exec")
		p_asset := r.FormValue("Trade_asset")
		direction := r.FormValue("Trade_direction")
		p_qty := r.FormValue("Trade_qty")
		price_curr := r.FormValue("Trade_curr")
		p_price := r.FormValue("Trade_price")
		p_amount := r.FormValue("Trade_amount")
		p_accint := r.FormValue("Trade_accint")
		broker_comm_curr := r.FormValue("Trade_brok_comm_curr")
		p_broker_comm := r.FormValue("Trade_brok_comm")
		p_portfolio := r.FormValue("Trade_portfolio")

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		asset, err := strconv.Atoi(p_asset)
		if err != nil {
			asset = -1
			log.Println("Can't convert ASSET ID parameter")
		}

		portfolio, err := strconv.Atoi(p_portfolio)
		if err != nil {
			portfolio = -1
			log.Println("Can't convert PORTFOLIO ID parameter")
		}

		qty, err := strconv.ParseFloat(p_qty, 64)
		if err != nil {
			qty = 0
			log.Println("Can't convert QTY parameter")
		}

		price, err := strconv.ParseFloat(p_price, 64)
		if err != nil {
			price = 0
			log.Println("Can't convert PRICE parameter")
		}

		amount, err := strconv.ParseFloat(p_amount, 64)
		if err != nil {
			amount = 0
			log.Println("Can't convert AMOUNT parameter")
		}

		accint, err := strconv.ParseFloat(p_accint, 64)
		if err != nil {
			accint = 0
			log.Println("Can't convert ACCINT parameter")
		}

		broker_comm, err := strconv.ParseFloat(p_broker_comm, 64)
		if err != nil {
			broker_comm = 0
			log.Println("Can't convert BROKER COMM parameter")
		}

		if id == -1 {
			error_msg = error_msg + "'ID' can't be empty;"
		}

		if number == "" {
			error_msg = error_msg + "'Number' can't be empty;"
		}

		if date_deal == "" {
			error_msg = error_msg + "'Deal Date' can't be empty;"
		}

		if date_exec == "" {
			error_msg = error_msg + "'Deal Date Exec' can't be empty;"
		}

		if asset == -1 {
			error_msg = error_msg + "'Asset' can't be empty;"
		}

		if direction == "" {
			error_msg = error_msg + "'Direction' can't be empty;"
		}

		if qty == 0 {
			error_msg = error_msg + "'Qty' can't be empty;"
		}

		if price_curr == "" {
			error_msg = error_msg + "'Price Currency' can't be empty;"
		}

		if price == 0 {
			error_msg = error_msg + "'Price' can't be empty;"
		}

		if amount == 0 {
			error_msg = error_msg + "'Amount' can't be empty;"
		}

		if portfolio == -1 {
			error_msg = error_msg + "'Portfolio' can't be empty;"
		}

		if error_msg == "" {

			error_info := addEditDeal(id, number, date_deal, date_exec, asset, direction, qty, price_curr, price, amount, broker_comm_curr, broker_comm, portfolio, accint)
			if error_info == nil {
				http.Redirect(w, r, "/deals", http.StatusMovedPermanently)
			} else {
				list := getDealsList()
				list.ErrorMsg = error_info.Error()

				source_path := filepath.Join(exPath, "gui", "deals.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, list)
			}

		} else {

			deal := Deal{}
			deal.Id = id
			deal.Number = number
			deal.Date = date_deal
			deal.Date_ex = date_exec
			deal.Asset = asset
			deal.Direction = direction
			deal.Qty = qty
			deal.Price_Currency = price_curr
			deal.Price = price
			deal.Amount = amount
			deal.Broker_Comm_Currency = broker_comm_curr
			deal.Broker_Comm = broker_comm
			deal.Portfolio = portfolio
			deal.AccInt = accint
			deal.Assets = getAssetArray()
			deal.Currencies = getCurrencyArray()
			deal.Portfolios = getPortfolioArray()
			deal.ErrorMsg = error_msg

			source_path := filepath.Join(exPath, "gui", "deal_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, deal)
		}
	} else {
		vars := mux.Vars(r)
		p_id := vars["id"]

		id, err := strconv.Atoi(p_id)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		element, err_get := getDeal(id)
		if err_get != nil {
			list := getDealsList()
			list.ErrorMsg = err_get.Error()

			source_path := filepath.Join(exPath, "gui", "deals.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, list)
		} else {
			element.Assets = getAssetArray()
			element.Currencies = getCurrencyArray()
			element.Portfolios = getPortfolioArray()

			var source_path = filepath.Join(exPath, "gui", "deal_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}
	}
}

func dealAddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		number := r.FormValue("Trade_number")
		date_deal := r.FormValue("Trade_date")
		date_exec := r.FormValue("Trade_date_exec")
		p_asset := r.FormValue("Trade_asset")
		direction := r.FormValue("Trade_direction")
		p_qty := r.FormValue("Trade_qty")
		price_curr := r.FormValue("Trade_curr")
		p_price := r.FormValue("Trade_price")
		p_amount := r.FormValue("Trade_amount")
		p_accint := r.FormValue("Trade_accint")
		broker_comm_curr := r.FormValue("Trade_brok_comm_curr")
		p_broker_comm := r.FormValue("Trade_brok_comm")
		p_portfolio := r.FormValue("Trade_portfolio")

		asset, err := strconv.Atoi(p_asset)
		if err != nil {
			asset = -1
			log.Println("Can't convert ASSET ID parameter")
		}

		portfolio, err := strconv.Atoi(p_portfolio)
		if err != nil {
			portfolio = -1
			log.Println("Can't convert PORTFOLIO ID parameter")
		}

		qty, err := strconv.ParseFloat(p_qty, 64)
		if err != nil {
			qty = 0
			log.Println("Can't convert QTY parameter")
		}

		price, err := strconv.ParseFloat(p_price, 64)
		if err != nil {
			price = 0
			log.Println("Can't convert PRICE parameter")
		}

		amount, err := strconv.ParseFloat(p_amount, 64)
		if err != nil {
			amount = 0
			log.Println("Can't convert AMOUNT parameter")
		}

		accint, err := strconv.ParseFloat(p_accint, 64)
		if err != nil {
			accint = 0
			log.Println("Can't convert ACCINT parameter")
		}

		broker_comm, err := strconv.ParseFloat(p_broker_comm, 64)
		if err != nil {
			broker_comm = 0
			log.Println("Can't convert BROKER COMM parameter")
		}

		if number == "" {
			error_msg = error_msg + "'Number' can't be empty;"
		}

		if date_deal == "" {
			error_msg = error_msg + "'Deal Date' can't be empty;"
		}

		if date_exec == "" {
			error_msg = error_msg + "'Deal Date Exec' can't be empty;"
		}

		if asset == -1 {
			error_msg = error_msg + "'Asset' can't be empty;"
		}

		if direction == "" {
			error_msg = error_msg + "'Direction' can't be empty;"
		}

		if qty == 0 {
			error_msg = error_msg + "'Qty' can't be empty;"
		}

		if price_curr == "" {
			error_msg = error_msg + "'Price Currency' can't be empty;"
		}

		if price == 0 {
			error_msg = error_msg + "'Price' can't be empty;"
		}

		if amount == 0 {
			error_msg = error_msg + "'Amount' can't be empty;"
		}

		if portfolio == -1 {
			error_msg = error_msg + "'Portfolio' can't be empty;"
		}

		if error_msg == "" {

			error_info := addEditDeal(-1, number, date_deal, date_exec, asset, direction, qty, price_curr, price, amount, broker_comm_curr, broker_comm, portfolio, accint)
			if error_info == nil {
				http.Redirect(w, r, "/deals", http.StatusMovedPermanently)
			} else {
				list := getDealsList()
				list.ErrorMsg = error_info.Error()

				source_path := filepath.Join(exPath, "gui", "deals.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, list)
			}

		} else {

			deal := Deal{}
			deal.Number = number
			deal.Date = date_deal
			deal.Date_ex = date_exec
			deal.Asset = asset
			deal.Direction = direction
			deal.Qty = qty
			deal.Price_Currency = price_curr
			deal.Price = price
			deal.Amount = amount
			deal.Broker_Comm_Currency = broker_comm_curr
			deal.Broker_Comm = broker_comm
			deal.Portfolio = portfolio
			deal.AccInt = accint
			deal.Assets = getAssetArray()
			deal.Currencies = getCurrencyArray()
			deal.Portfolios = getPortfolioArray()
			deal.ErrorMsg = error_msg

			var source_path = filepath.Join(exPath, "gui", "deal_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, deal)
		}

	} else {

		deal := Deal{}
		deal.Assets = getAssetArray()
		deal.Currencies = getCurrencyArray()
		deal.Portfolios = getPortfolioArray()

		source_path := filepath.Join(exPath, "gui", "deal_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, deal)
	}
}

func dealDeleteHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.Atoi(p_id)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	error_info := deleteDeal(id)
	if error_info == nil {
		http.Redirect(w, r, "/deals", http.StatusMovedPermanently)
	} else {
		list := getDealsList()
		list.ErrorMsg = error_info.Error()

		source_path := filepath.Join(exPath, "gui", "deals.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, list)
	}
}

func dealTxHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.Atoi(p_id)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	deal_tx, error_get := getTxDeal(id)
	if error_get != nil {
		deal_tx.ErrorMsg = error_get.Error()
	}

	source_path := filepath.Join(exPath, "gui", "deal_tx.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, deal_tx)
}

func dealTxAddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.Atoi(p_id)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	error_del_tx := deleteTxDeal(id)
	if error_del_tx == nil {
		error_create := createTxDeal(id)

		if error_create != nil {
			deal_tx := Tx{}
			deal_tx.Id = id
			deal_tx.ErrorMsg = error_create.Error()

			source_path := filepath.Join(exPath, "gui", "deal_tx.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, deal_tx)
		} else {
			http.Redirect(w, r, "/deal_tx/"+p_id, http.StatusMovedPermanently)
		}
	} else {
		deal_tx := Tx{}
		deal_tx.Id = id
		deal_tx.ErrorMsg = error_del_tx.Error()

		source_path := filepath.Join(exPath, "gui", "deal_tx.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, deal_tx)
	}
}

func dealTxRemoveHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.Atoi(p_id)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	error_del_tx := deleteTxDeal(id)
	if error_del_tx != nil {

		deal_tx := Tx{}
		deal_tx.Id = id
		deal_tx.ErrorMsg = error_del_tx.Error()

		source_path := filepath.Join(exPath, "gui", "deal_tx.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, deal_tx)

	} else {
		http.Redirect(w, r, "/deal_tx/"+p_id, http.StatusMovedPermanently)
	}
}

// non-trade tx

func nontradeHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	var source_path = filepath.Join(exPath, "gui", "nontrade.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, nil)
}

func moneyInOutHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	list := getMoneyInOutList()

	source_path := filepath.Join(exPath, "gui", "money_inout.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, list)
}

func moneyInOutAddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {
		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		date_exec := r.FormValue("InOut_date_exec")
		p_amount := r.FormValue("InOut_amount")
		curr := r.FormValue("InOut_curr")
		p_portfolio := r.FormValue("InOut_portfolio")
		comment := r.FormValue("InOut_comment")

		portfolio, err := strconv.Atoi(p_portfolio)
		if err != nil {
			portfolio = -1
			log.Println("Can't convert PORTFOLIO ID parameter")
		}

		amount, err := strconv.ParseFloat(p_amount, 64)
		if err != nil {
			amount = 0
			log.Println("Can't convert AMOUNT parameter")
		}

		if date_exec == "" {
			error_msg = error_msg + "'Date Exec' can't be empty;"
		}

		if curr == "" {
			error_msg = error_msg + "'Currency' can't be empty;"
		}

		if amount == 0 {
			error_msg = error_msg + "'Amount' can't be empty;"
		}

		if portfolio == -1 {
			error_msg = error_msg + "'Portfolio' can't be empty;"
		}

		if error_msg == "" {

			error_info := addEditMoneyInOut(-1, date_exec, amount, curr, portfolio, comment)
			if error_info == nil {
				http.Redirect(w, r, "/money_in_out", http.StatusMovedPermanently)
			} else {

				list := getMoneyInOutList()
				list.ErrorMsg = error_info.Error()

				source_path := filepath.Join(exPath, "gui", "money_inout.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, list)

			}

		} else {

			inout := MoneyTx{}
			inout.DateTx = date_exec
			inout.Amount = amount
			inout.Currency = curr
			inout.Portfolio = portfolio
			inout.Comment = comment
			inout.Currencies = getCurrencyArray()
			inout.Portfolios = getPortfolioArray()
			inout.ErrorMsg = error_msg

			var source_path = filepath.Join(exPath, "gui", "money_inout_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, inout)
		}

	} else {

		inout := MoneyTx{}
		inout.Currencies = getCurrencyArray()
		inout.Portfolios = getPortfolioArray()

		source_path := filepath.Join(exPath, "gui", "money_inout_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, inout)
	}
}

func moneyInOutEditHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {
		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		p_id := r.FormValue("InOut_id")
		date_exec := r.FormValue("InOut_date_exec")
		p_amount := r.FormValue("InOut_amount")
		curr := r.FormValue("InOut_curr")
		p_portfolio := r.FormValue("InOut_portfolio")
		comment := r.FormValue("InOut_comment")

		id, err := strconv.ParseInt(p_id, 10, 64)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		portfolio, err := strconv.Atoi(p_portfolio)
		if err != nil {
			portfolio = -1
			log.Println("Can't convert PORTFOLIO ID parameter")
		}

		amount, err := strconv.ParseFloat(p_amount, 64)
		if err != nil {
			amount = 0
			log.Println("Can't convert AMOUNT parameter")
		}

		if date_exec == "" {
			error_msg = error_msg + "'Date Exec' can't be empty;"
		}

		if curr == "" {
			error_msg = error_msg + "'Currency' can't be empty;"
		}

		if amount == 0 {
			error_msg = error_msg + "'Amount' can't be empty;"
		}

		if portfolio == -1 {
			error_msg = error_msg + "'Portfolio' can't be empty;"
		}

		if id == -1 {
			error_msg = error_msg + "'ID' can't be empty;"
		}

		if error_msg == "" {

			error_info := addEditMoneyInOut(id, date_exec, amount, curr, portfolio, comment)
			if error_info == nil {
				http.Redirect(w, r, "/money_in_out", http.StatusMovedPermanently)
			} else {

				list := getMoneyInOutList()
				list.ErrorMsg = error_info.Error()

				source_path := filepath.Join(exPath, "gui", "money_inout.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, list)

			}

		} else {

			inout := MoneyTx{}
			inout.Id = id
			inout.DateTx = date_exec
			inout.Amount = amount
			inout.Currency = curr
			inout.Portfolio = portfolio
			inout.Comment = comment
			inout.Currencies = getCurrencyArray()
			inout.Portfolios = getPortfolioArray()
			inout.ErrorMsg = error_msg

			var source_path = filepath.Join(exPath, "gui", "money_inout_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, inout)
		}

	} else {

		vars := mux.Vars(r)
		p_id := vars["id"]

		id, err := strconv.ParseInt(p_id, 10, 64)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		element, err_get := getMoneyInOut(id)
		if err_get != nil {
			list := getMoneyInOutList()
			list.ErrorMsg = err_get.Error()
			source_path := filepath.Join(exPath, "gui", "money_inout.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, list)

		} else {

			element.Currencies = getCurrencyArray()
			element.Portfolios = getPortfolioArray()

			var source_path = filepath.Join(exPath, "gui", "money_inout_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}

	}
}

func moneyIntOutDeleteHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.ParseInt(p_id, 10, 64)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	error_info := deleteMoneyInOut(id)
	if error_info == nil {
		http.Redirect(w, r, "/money_in_out", http.StatusMovedPermanently)
	} else {
		list := getMoneyInOutList()
		list.ErrorMsg = error_info.Error()

		source_path := filepath.Join(exPath, "gui", "money_inout.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, list)
	}
}

func assetInOutHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	list := getAssetInOutList()

	source_path := filepath.Join(exPath, "gui", "asset_inout.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, list)
}

func assetInOutAddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {
		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		date_exec := r.FormValue("InOut_date_exec")
		p_qty := r.FormValue("InOut_qty")
		p_asset := r.FormValue("InOut_asset")
		p_portfolio := r.FormValue("InOut_portfolio")
		comment := r.FormValue("InOut_comment")

		portfolio, err := strconv.Atoi(p_portfolio)
		if err != nil {
			portfolio = -1
			log.Println("Can't convert PORTFOLIO ID parameter")
		}

		qty, err := strconv.ParseFloat(p_qty, 64)
		if err != nil {
			qty = 0
			log.Println("Can't convert QTY parameter")
		}

		asset, err := strconv.Atoi(p_asset)
		if err != nil {
			asset = -1
			log.Println("Can't convert ASSET parameter")
		}

		if date_exec == "" {
			error_msg = error_msg + "'Date Exec' can't be empty;"
		}

		if portfolio == -1 {
			error_msg = error_msg + "'Portfolio' can't be empty;"
		}

		if asset == -1 {
			error_msg = error_msg + "'Asset' can't be empty;"
		}

		if qty == 0 {
			error_msg = error_msg + "'Qty' can't be empty;"
		}

		if error_msg == "" {

			error_info := addEditAssetInOut(-1, date_exec, qty, asset, portfolio, comment)
			if error_info == nil {
				http.Redirect(w, r, "/asset_in_out", http.StatusMovedPermanently)
			} else {

				list := getAssetInOutList()
				list.ErrorMsg = error_info.Error()

				source_path := filepath.Join(exPath, "gui", "asset_inout.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, list)

			}

		} else {

			inout := AssetTx{}
			inout.DateTx = date_exec
			inout.Qty = qty
			inout.Asset = asset
			inout.Portfolio = portfolio
			inout.Comment = comment
			inout.Assets = getAssetArray()
			inout.Portfolios = getPortfolioArray()
			inout.ErrorMsg = error_msg

			var source_path = filepath.Join(exPath, "gui", "asset_inout_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, inout)
		}

	} else {

		inout := AssetTx{}
		inout.Assets = getAssetArray()
		inout.Portfolios = getPortfolioArray()

		source_path := filepath.Join(exPath, "gui", "asset_inout_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, inout)
	}
}

func assetInOutEditHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {
		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		p_id := r.FormValue("InOut_id")
		date_exec := r.FormValue("InOut_date_exec")
		p_qty := r.FormValue("InOut_qty")
		p_asset := r.FormValue("InOut_asset")
		p_portfolio := r.FormValue("InOut_portfolio")
		comment := r.FormValue("InOut_comment")

		id, err := strconv.ParseInt(p_id, 10, 64)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		portfolio, err := strconv.Atoi(p_portfolio)
		if err != nil {
			portfolio = -1
			log.Println("Can't convert PORTFOLIO ID parameter")
		}

		qty, err := strconv.ParseFloat(p_qty, 64)
		if err != nil {
			qty = 0
			log.Println("Can't convert QTY parameter")
		}

		asset, err := strconv.Atoi(p_asset)
		if err != nil {
			asset = -1
			log.Println("Can't convert ASSET parameter")
		}

		if date_exec == "" {
			error_msg = error_msg + "'Date Exec' can't be empty;"
		}

		if asset == -1 {
			error_msg = error_msg + "'Asset' can't be empty;"
		}

		if qty == 0 {
			error_msg = error_msg + "'Qty' can't be empty;"
		}

		if portfolio == -1 {
			error_msg = error_msg + "'Portfolio' can't be empty;"
		}

		if id == -1 {
			error_msg = error_msg + "'ID' can't be empty;"
		}

		if error_msg == "" {

			error_info := addEditAssetInOut(id, date_exec, qty, asset, portfolio, comment)
			if error_info == nil {
				http.Redirect(w, r, "/asset_in_out", http.StatusMovedPermanently)
			} else {

				list := getAssetInOutList()
				list.ErrorMsg = error_info.Error()

				source_path := filepath.Join(exPath, "gui", "asset_inout.html")
				t, _ := template.ParseFiles(source_path)
				t.Execute(w, list)

			}

		} else {

			inout := AssetTx{}
			inout.Id = id
			inout.DateTx = date_exec
			inout.Qty = qty
			inout.Asset = asset
			inout.Portfolio = portfolio
			inout.Comment = comment
			inout.Assets = getAssetArray()
			inout.Portfolios = getPortfolioArray()
			inout.ErrorMsg = error_msg

			var source_path = filepath.Join(exPath, "gui", "asset_inout_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, inout)
		}

	} else {

		vars := mux.Vars(r)
		p_id := vars["id"]

		id, err := strconv.ParseInt(p_id, 10, 64)
		if err != nil {
			id = -1
			log.Println("Can't convert ID parameter")
		}

		element, err_get := getAssetInOut(id)
		if err_get != nil {
			list := getAssetInOutList()
			list.ErrorMsg = err_get.Error()
			source_path := filepath.Join(exPath, "gui", "asset_inout.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, list)

		} else {

			element.Assets = getAssetArray()
			element.Portfolios = getPortfolioArray()

			var source_path = filepath.Join(exPath, "gui", "asset_inout_form.html")
			t, _ := template.ParseFiles(source_path)
			t.Execute(w, element)
		}

	}
}

func assetIntOutDeleteHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	vars := mux.Vars(r)
	p_id := vars["id"]

	id, err := strconv.ParseInt(p_id, 10, 64)
	if err != nil {
		id = -1
		log.Println("Can't convert ID parameter")
	}

	error_info := deleteAssetInOut(id)
	if error_info == nil {
		http.Redirect(w, r, "/asset_in_out", http.StatusMovedPermanently)
	} else {
		list := getAssetInOutList()
		list.ErrorMsg = error_info.Error()

		source_path := filepath.Join(exPath, "gui", "asset_inout.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, list)
	}
}

// reports

func reportsHandler(w http.ResponseWriter, r *http.Request) {

	var source_path = filepath.Join(exPath, "gui", "reports.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, nil)
}

func reportsOnDateHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	balance := ReportOnDate{}

	if r.Method == "POST" {

		date_report := r.FormValue("date_report")
		balance = getReportOnDate(date_report)

	}

	var source_path = filepath.Join(exPath, "gui", "reports_ondate.html")
	t, _ := template.ParseFiles(source_path)
	t.Execute(w, balance)
}

// settings
func settingsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")

	if r.Method == "POST" {

		error_msg := ""

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		p_comm_in_deal := r.FormValue("comm_in_deal")
		workdir := r.FormValue("workdir")
		port := r.FormValue("port")

		comm_in_deal, err := strconv.Atoi(p_comm_in_deal)
		if err != nil {
			comm_in_deal = -1
			log.Println("Can't convert COMM IN DEAL parameter")
		}

		if workdir == "" {
			error_msg = error_msg + "'Work Directory' can't be empty"
		}

		if port == "" {
			error_msg = error_msg + "'Port' can't be empty"
		}

		error_info := saveSettings(workdir, port, comm_in_deal)

		settings := Settings{}
		settings.BrokerCommInDeal = comm_in_deal
		settings.Port = port
		settings.WorkDir = workdir

		if error_msg == "" && error_info == nil {
			settings.ErrorMsg = "Settings saved"
		} else if error_msg != "" {
			settings.ErrorMsg = error_msg
		} else if error_info != nil {
			settings.ErrorMsg = error_info.Error()
		}

		source_path := filepath.Join(exPath, "gui", "settings_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, settings)

	} else {

		config, err := getSettings()
		if err != nil {
			config.ErrorMsg = err.Error()
		}
		source_path := filepath.Join(exPath, "gui", "settings_form.html")
		t, _ := template.ParseFiles(source_path)
		t.Execute(w, config)
	}
}

// main
func main() {

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

	database = db

	app_settings, err := getSettings()
	if err != nil {
		log.Fatal(err)
	}

	if app_settings.WorkDir == "" {
		log.Fatal("Setting 'Work directory' is empty - can't start app. Use 'firts_start' for filling settings!")
	}

	if app_settings.Port == "" {
		log.Fatal("Setting 'Port' is empty - can't start app. Use 'firts_start' for filling settings!")
	}

	exPath = app_settings.WorkDir
	listen_host := ":" + app_settings.Port
	comm_broker_tx = app_settings.BrokerCommInDeal

	fmt.Println("")
	fmt.Println("███████ ██ ███    ███ ██████  ██      ███████     ██ ███    ██ ██    ██ ███████ ███████ ████████")
	fmt.Println("██      ██ ████  ████ ██   ██ ██      ██          ██ ████   ██ ██    ██ ██      ██         ██")
	fmt.Println("███████ ██ ██ ████ ██ ██████  ██      █████       ██ ██ ██  ██ ██    ██ █████   ███████    ██")
	fmt.Println("     ██ ██ ██  ██  ██ ██      ██      ██          ██ ██  ██ ██  ██  ██  ██           ██    ██")
	fmt.Println("███████ ██ ██      ██ ██      ███████ ███████     ██ ██   ████   ████   ███████ ███████    ██")
	fmt.Println("")
	fmt.Println("Version " + version)
	fmt.Println("")
	fmt.Println("SQLite DB:", pathToConfig)
	fmt.Println("Work directory:", exPath)
	fmt.Println("Web App URL:", "http://localhost"+listen_host)

	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/settings", settingsHandler)
	router.HandleFunc("/catalogs", catalogsHandler)
	// Currency func
	router.HandleFunc("/currencies", currenciesHandler)
	router.HandleFunc("/cur_add", curAddHandler)
	router.HandleFunc("/cur_edit/{id}", curEditHandler)
	router.HandleFunc("/cur_delete/{id}", curDeleteHandler)
	// Rates
	router.HandleFunc("/rates/{id}", ratesHandler)
	router.HandleFunc("/rate_add/{id}", rateAddHandler)
	router.HandleFunc("/rate_edit/{id}", rateEditHandler)
	router.HandleFunc("/rate_delete/{id}", rateDeleteHandler)
	// Portfolio func
	router.HandleFunc("/portfolios", portfoliosHandler)
	router.HandleFunc("/folio_add", portfolioAddHandler)
	router.HandleFunc("/folio_edit/{id}", portfolioEditHandler)
	router.HandleFunc("/folio_delete/{id}", portfolioDeleteHandler)
	// Asset func
	router.HandleFunc("/assets", assetsHandler)
	router.HandleFunc("/asset_add", assetAddHandler)
	router.HandleFunc("/asset_edit/{id}", assetEditHandler)
	router.HandleFunc("/asset_delete/{id}", assetDeleteHandler)
	// Quotes
	router.HandleFunc("/quotes/{id}", quotesHandler)
	router.HandleFunc("/quote_add/{id}", quoteAddHandler)
	router.HandleFunc("/quote_edit/{id}", quoteEditHandler)
	router.HandleFunc("/quote_delete/{id}", quoteDeleteHandler)
	// Deal func
	router.HandleFunc("/deals", dealsHandler)
	router.HandleFunc("/deal_add", dealAddHandler)
	router.HandleFunc("/deal_edit/{id}", dealEditHandler)
	router.HandleFunc("/deal_delete/{id}", dealDeleteHandler)
	router.HandleFunc("/deal_tx/{id}", dealTxHandler)
	router.HandleFunc("/tx_add/{id}", dealTxAddHandler)
	router.HandleFunc("/tx_remove/{id}", dealTxRemoveHandler)
	// Non-Trade Tx
	router.HandleFunc("/in_out", nontradeHandler)
	router.HandleFunc("/money_in_out", moneyInOutHandler)
	router.HandleFunc("/money_in_out_add", moneyInOutAddHandler)
	router.HandleFunc("/money_in_out_edit/{id}", moneyInOutEditHandler)
	router.HandleFunc("/money_in_out_delete/{id}", moneyIntOutDeleteHandler)
	router.HandleFunc("/asset_in_out", assetInOutHandler)
	router.HandleFunc("/asset_in_out_add", assetInOutAddHandler)
	router.HandleFunc("/asset_in_out_edit/{id}", assetInOutEditHandler)
	router.HandleFunc("/asset_in_out_delete/{id}", assetIntOutDeleteHandler)
	// Reports
	router.HandleFunc("/reports", reportsHandler)
	router.HandleFunc("/report_balance", reportsOnDateHandler)

	http.Handle("/", router)
	http.ListenAndServe(listen_host, nil)
}
