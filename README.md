# Stock market core
## What is this?
This is an attempt to create a high performance API (core) for an imaginary stock market with as many users as possible.
## Why?
1. I wanted to create something to learn more about [Go](https://golang.org/) and challenge my knowledge.
2. I wanted to add a well-documented and well-tested project to my resume, and I hope that this would be it!.
3. [Iran's Stock market](http://www.tsetmc.com) has suffered so many outages and is still performing poorly, so I wonder if I can create something better.

## Stack
+ [Go Programming Language](https://golang.org/) is the language of choice for this project
+ [PostgreSQL](https://www.postgresql.org) is used as the RDBMS 

## Getting started
1. Follow [this](https://golang.org/doc/install) guide to install Go.
2. install `sql-migrate`: `$ go get -v github.com/rubenv/sql-migrate/...`
3. set required environment variables for the database connection. These variables are: 
`POSTGRES_PORT`, `POSTGRES_USER`,`POSTGRES_PASSWORD`,`DATABASE_NAME`
4. make sure you have [docker](https://www.docker.com) installed and run: `$ make db`. this will create the postgres database.
5. `$ make` this will apply the migrations and try to start the server on port `PORT` if it is set, otherwise on port `8080`

+ use `$ make destroy` to delete the database
+ use `$ make clean` to rollback all the migrations
