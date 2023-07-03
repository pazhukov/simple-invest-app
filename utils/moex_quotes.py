import sys
import sqlite3
import os
from datetime import datetime
import requests

# search url https://iss.moex.com/iss/securities.json?q=ru000a0jv4l2
search_url = "https://iss.moex.com/iss/securities.json?q={0}"

# shares https://iss.moex.com/iss/history/engines/stock/markets/shares/boards/TQTF/securities/RCMM.json?from=2023-06-01&till=2023-06-01
share_url = "https://iss.moex.com/iss/history/engines/stock/markets/shares/boards/{0}/securities/{1}.json?from={2}&till={3}"

# bonds https://iss.moex.com/iss/history/engines/stock/markets/bonds/boards/TQOB/securities/SU29006RMFS2.xml?from=2023-06-01&till=2023-06-01
bond_url = "https://iss.moex.com/iss/history/engines/stock/markets/bonds/boards/{0}/securities/{1}.json?from={2}&till={3}"


def load_bond(con, cur, asset, date_quote, ticker, update):
    url = search_url.format(ticker)
    response = requests.request("GET", url)
    moex_sec = response.json()    
    if len(moex_sec['securities']['data']) == 0:
        return -1
    
    board_id = ""
    sec_id = ""
    for sec in moex_sec['securities']['data']:
        item = sec
        if item[5] != ticker:
            continue
        else:
            board_id = item[14]
            sec_id = item[1]

    if board_id == "":
        return -1
    
    url = bond_url.format(board_id, sec_id, date_quote, date_quote)
    response = requests.request("GET", url)
    moex_q = response.json()
    quotes = moex_q['history']['data']
    if len(quotes) == 0:
        return -1
    
    for quote in quotes:
        item = quote
        open = item[13]
        max = item[7]
        min = item[6]
        close = item[8]
        accint = item[11]
        currency = item[31]
        if currency == "SUR":
            currency = "RUB"

        if update == -1:
            cur.execute("insert into asset_quotes(date_quote, asset, currency, open, min, max, close, accint) values (?, ?, ?, ?, ?, ?, ?, ?)", (date_quote, asset, currency, open, min, max, close, accint))
        else:
            cur.execute("update asset_quotes set date_quote = ?, asset = ?, currency = ?, open = ?, min = ?, max = ?, close = ?, accint = ? where id = ?", (date_quote, asset, currency, open, min, max, close, accint, update))
        con.commit() 

    return 0


def load_share(con, cur, asset, date_quote, ticker, update):
    url = search_url.format(ticker)
    response = requests.request("GET", url)
    moex_sec = response.json()    
    if len(moex_sec['securities']['data']) == 0:
        return -1
    
    board_id = ""
    for sec in moex_sec['securities']['data']:
        item = sec
        if item[1] != ticker:
            continue
        else:
            board_id = item[14]
    
    if board_id == "":
        return -1
    
    url = share_url.format(board_id, ticker, date_quote, date_quote)
    response = requests.request("GET", url)
    moex_q = response.json()
    quotes = moex_q['history']['data']
    if len(quotes) == 0:
        return -1
    
    for quote in quotes:
        item = quote
        open = item[6]
        max = item[8]
        min = item[7]
        close = item[11]
        accint = 0
        currency = item[21]
        if currency == "SUR":
            currency = "RUB"

        if update == -1:
            cur.execute("insert into asset_quotes(date_quote, asset, currency, open, min, max, close, accint) values (?, ?, ?, ?, ?, ?, ?, ?)", (date_quote, asset, currency, open, min, max, close, accint))
        else:
            cur.execute("update asset_quotes set date_quote = ?, asset = ?, currency = ?, open = ?, min = ?, max = ?, close = ?, accint = ? where id = ?", (date_quote, asset, currency, open, min, max, close, accint, update))
        con.commit() 

    return 0


def run_load(path_db, date_quote):
    if os.path.isfile(path_db) == False:
        print("File DB is not exist")
        return
    format_date = "%Y-%m-%d"
    try:
        res = bool(datetime.strptime(date_quote, format_date))
    except ValueError:
        res = False
    if res == False:
        print("Date is not valid")
        return
    
    con = sqlite3.connect(path_db)
    cur = con.cursor() 

    result = cur.execute("select a.id, a.ticker, a.type, ifnull(q.id, -1) as quote_id from assets as a left join asset_quotes as q on a.id = q.asset and q.date_quote = ?", (date_quote,))
    rows = result.fetchall()  
    for row in rows:
        asset = row[0]
        ticker = row[1]
        type = row[2]
        quote_id = row[3]
        if type == 1:
            res = load_bond(con, cur, asset, date_quote, ticker, quote_id)
        else:
            res = load_share(con, cur, asset, date_quote, ticker, quote_id)
        if res == -1:
            print("Can't get quote for " + ticker)

    con.close()
    print("Quotes loaded")


if __name__ == "__main__":
    path = ""
    date = ""
    if len(sys.argv) == 1 or len(sys.argv) > 3:
        print("Usage: python moex_quotes.py path_to_db quotes_date")
        print("quotes_date as YYYY-MM-DD")
        sys.exit()
    if len(sys.argv) == 2:
        print("Using DB in " + sys.argv[1])
        path = sys.argv[1]
    if len(sys.argv) == 3:
        print("Using DB in " + sys.argv[1])
        print("Using DATE as " + sys.argv[2])
        path = sys.argv[1]
        date = sys.argv[2]
    print("")
    run_load(path, date)