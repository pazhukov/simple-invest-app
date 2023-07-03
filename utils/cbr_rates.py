import sys
import sqlite3
import os
from datetime import datetime
import requests

# get rates from json https://www.cbr-xml-daily.ru/archive/2023/06/11/daily_json.js
url_cbr = "https://www.cbr-xml-daily.ru/archive/{0}/daily_json.js"

def run_load(path_db, date_rates):
    if os.path.isfile(path_db) == False:
        print("File DB is not exist")
        return
    format_date = "%Y-%m-%d"
    try:
        res = bool(datetime.strptime(date_rates, format_date))
    except ValueError:
        res = False
    if res == False:
        print("Date is not valid")
        return
    
    date_rates = date_rates.replace("-", "/")
    url = url_cbr.format(date_rates)
    response = requests.request("GET", url)
    data_rates = response.json()

    try:
        rate = data_rates.get('Valute')
    except:
        print("Can't get rates from url")
        return
    
    con = sqlite3.connect(path_db)
    cur = con.cursor()

    result = cur.execute("SELECT cur.name, ifnull(rates.rate, -1) as rate FROM currencies as cur left join (select currency, rate from currency_rates where date_rate=?) as rates  on cur.name = rates.currency", (date_rates,))
    rows = result.fetchall()
    for row in rows:
        currrency = row[0]
        value = row[1]

        try:
            new_value = rate.get(currrency).get('Value')
        except:
            print("Can't get rate for " + currrency)
            continue

        if value == -1:
            cur.execute("insert into currency_rates(date_rate, currency, rate) values (?, ?, ?)", (date_rates, currrency, new_value))
        else:
            cur.execute("update currency_rates set rate = ? where date_rate = ? and currency = ?", (new_value, date_rates, currrency))
        con.commit()     

    con.close()
    print("Rates loaded")


if __name__ == "__main__":
    path = ""
    date = ""
    if len(sys.argv) == 1 or len(sys.argv) > 3:
        print("Usage: python cbr_rates.py path_to_db rates_date")
        print("rate_date as YYYY-MM-DD")
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