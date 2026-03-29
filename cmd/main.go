package main

import (
	"log"
	"os"

	"github.com/osamikoyo/math-angel/internal/app"
)

func main() {
	path := ""
	for i, arg := range os.Args {
		if arg == "--config" {
			path = os.Args[i+1]
		}
	}

	app, err := app.SetupApp(path)
	if err != nil {
		log.Println(err)

		return
	}

	err = app.Run()
	if err != nil {
		log.Println(err)

		return
	}
}
