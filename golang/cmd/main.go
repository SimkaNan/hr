package main

import (
	"taskWorker/service"
)

func main() {
	s := service.New()

	s.Start(10)
}
