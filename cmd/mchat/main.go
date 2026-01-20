package main

import (
	"log"
	"mchat/internal/data"
	"mchat/internal/ui"
	"os"
)

func main() {
	logFile, err := setupLogger("mchat.log")
	if err != nil {
		log.Fatalf("failed to setup logger: %v", err)
	}
	defer logFile.Close()

	svc, err := data.NewDataService()
	if err != nil {
		log.Fatalf("failed to setup dataservice: %v", err)
	}

	app := ui.NewApp(svc)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func setupLogger(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	log.SetOutput(file)
	return file, nil
}
