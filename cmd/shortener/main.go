package main

import (
	"github.com/olya2209/yandex-hw/internal/server"
)

func main() {
	s := server.NewServer()
	s.Route()
	s.Run()
}
