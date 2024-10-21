package main

import (
	"fmt"
	"intermark/internal/app"
	"intermark/internal/database"
	"intermark/internal/utils"
)

func startup() {
	if utils.ArgPresent("-v") {
		fmt.Println("Intermark:", app.Version)
		return
	}

	if !utils.Config.Load() {
		fmt.Println("Generated default config file")
		return
	}

	utils.InitLogger()

	if !utils.GitInstalled() {
		fmt.Println("Issue with git installation, see logs for details. Make sure git is installed and in your PATH.")
		return
	}

	if !utils.TailwindInstalled() {
		fmt.Println("Issue with tailwind installation, see logs for details. Make sure node/npm is installed and in your PATH. Also don't forget to run 'npm install' in the project root directory.")
		return
	}

	utils.InitMarkdownConverter()

	database.Init()

	app.ServerInstance.Start()
}

func cleanup() {
	database.Close()
	utils.CleanupLogger()
}

func main() {
	startup()
	defer cleanup()
}
