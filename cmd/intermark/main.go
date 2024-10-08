package main

import (
	"fmt"
	"intermark/internal/app"
	"intermark/internal/database"
	"intermark/internal/utils"
)

func startup() {
	if !utils.Config.Load() {
		fmt.Println("Generated default config file")
		return
	}
	if utils.ArgPresent("-v") {
		fmt.Println("Intermark:", app.Version)
		return
	}
	utils.InitLogger()
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
