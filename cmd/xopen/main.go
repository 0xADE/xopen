package main

import (
	"log"
	"os"

	"github.com/0xADE/ade-ctld/client/exe"

	"github.com/0xADE/xopen/ui"
)

func main() {
	c, err := exe.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	app := ui.New(c)
	if err := app.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}

	os.Exit(0)
}
