package ui

import (
	"fmt"
	"log"
	"os"

	"github.com/rivo/tview"

	"mchat/internal/data"
)

type App struct {
	app      *tview.Application
	flex     *tview.Flex
	list     *tview.List
	textView *tview.TextView
}

func (a *App) GetMessages() {
	messages := data.GetChats()
	a.list.Clear()
	for _, msg := range messages {
		a.list.AddItem(msg.Subject, "", 0, func() {
			a.textView.SetText(msg.Content)
		})
	}
	a.initList()
}

func (a *App) initList() {
	a.list.AddItem("Fetch", "Get messages", 'm', func() {
		a.GetMessages()
	}).AddItem("Quit", "Press to exit", 'q', func() {
		a.app.Stop()
	}).AddItem("Logs", "Show logs", 'l', func() {
		a.app.Suspend(func() {
			fmt.Scanln()
		})
	})
	a.list.ShowSecondaryText(true)
}

func NewApp() *App {
	file, err := os.OpenFile("mchat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := tview.NewApplication()
	list := tview.NewList()

	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetWrap(true).
		SetBorder(true).
		SetTitle(" Chat ")
	textView.SetText("Welcome to mChat. Press 'm' to fetch the messages.")

	flex := tview.NewFlex().
		AddItem(list, 0, 1, true).
		AddItem(textView, 0, 2, false)

	a := &App{
		app:      app,
		flex:     flex,
		list:     list,
		textView: textView,
	}
	a.initList()
	return a
}

func (a *App) Run() {
	if err := a.app.SetRoot(a.flex, true).SetFocus(a.list).Run(); err != nil {
		panic(err)
	}
}
