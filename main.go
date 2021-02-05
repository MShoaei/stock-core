package main

import "github.com/MShoaei/stock-core/server"

func main() {
	s := server.NewServer()
	s.Run()
}
