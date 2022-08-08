package main

import (
	"log"
	"os"

	"microservice-template/cmd/root"
	"microservice-template/cmd/serve"
	"microservice-template/internal"
)

func main() {
	app, err := internal.NewApplication()

	if err != nil {
		log.Println("An error occurred", err)
		os.Exit(1)
	}

	rootCmd := root.Cmd(app)
	rootCmd.AddCommand(serve.Cmd(app))

	if err := rootCmd.Execute(); err != nil {
		log.Println("An error occurred", err)
		os.Exit(1)
	}
}
