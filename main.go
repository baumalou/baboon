package main

import (
	"log"
	"time"
)

func main() {
	for {
		log.Println("operator started")
		time.Sleep(100 * time.Minute)
	}

}
