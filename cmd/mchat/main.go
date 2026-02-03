package main

import (
	"log"
	"mchat/internal/data"
	"mchat/internal/models"
	"mchat/internal/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	logFile, err := setupLogger("mchat.log")
	if err != nil {
		log.Fatalf("failed to setup logger: %v", err)
	}
	defer logFile.Close()

	msgChan := make(chan *models.Message, 100)
	svc, err := data.NewDataService(msgChan)

	if err != nil {
		log.Fatalf("failed to setup dataservice: %v", err)
	}
	app := tea.NewProgram(ui.InitialModel(svc), tea.WithAltScreen())

	go func() {
		for {
			m := <-msgChan
			app.Send(m)
		}
	}()

	if _, err := app.Run(); err != nil {
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
