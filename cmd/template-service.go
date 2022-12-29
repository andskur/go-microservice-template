package main

import (
	"os"

	"github.com/Misnaged/annales/logger"

	"microservice-template/cmd/root"
	"microservice-template/cmd/serve"
	"microservice-template/internal"
)

// main is the entry point of the application
// It creates an instance of the internal application and adds
// the "serve" command to the root command. Then it executes
// the root command. If any error occurs, it logs the error and
// exits the application with status code 1.
func main() {
	app, err := internal.NewApplication()
	if err != nil {
		logger.Log().Infof("An error occurred: %s", err.Error())
		os.Exit(1)
	}

	rootCmd := root.Cmd(app)
	rootCmd.AddCommand(serve.Cmd(app))

	if err := rootCmd.Execute(); err != nil {
		logger.Log().Infof("An error occurred: %s", err.Error())
		os.Exit(1)
	}
}
