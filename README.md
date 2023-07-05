# Simple Invest
### Description
Web app for asset management, more details on [page](https://anteater.dev/simple-invest/)

### How start
To start, you need to move an empty database to work in the desired directory on the disk, to start the application, you need to specify which database it will work with. As an example, consider that the database will be located in the directory c:\temp with the name default.db
First you need to run first_start.go
```
go run first_start.go -db c:\temp\default.db
```
You need to fill in three parameters

+ Working directory - this directory should contain the gui directory. For example, if you move the gui directory c:\temp, then the setting will be c:\temp\
+ The port on which the application will run - any free and convenient port for you
+ Movement of commissions in a deal - if set to 1, then when creating transactions for a deal, the commission will be associated with each deal. With a value of 0, it is assumed that the entire commission is deducted in one transaction per day

After updating the settings, you can run the application
```
go run app.go -db c:\temp\default.db
```
### Utils
In folder utils added some additional utilities
+ cbr_rates.py - load currencies rates from CBR 
+ moex_quotes.py - load assets quotes from MOEX

## Plans
- [ ] Load deal from Open Broker XML file
- [ ] Add futures/options
- [ ] Add accint calendar
