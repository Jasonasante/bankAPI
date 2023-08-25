package main

import (
	"log"
)

func main() {
	store, err := NewDB()
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}
	server := NewAPIServer(":3500", store)
	server.Run()
}
