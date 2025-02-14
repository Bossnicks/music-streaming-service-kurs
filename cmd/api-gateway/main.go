package main

import (
	"github.com/Bossnicks/music-streaming-service-kurs/internal/gateway"
)

func main() {
	gw := gateway.NewGateway()
	gw.Start(":9999")
}
