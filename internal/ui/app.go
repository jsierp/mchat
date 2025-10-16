package ui

import (
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
	messages := data.GetData()
	a.list.Clear()
	for _, msg := range messages {
		a.list.AddItem(msg.Subject, "", 'a', func() {
			a.textView.SetText(msg.Content)
		})
	}
}

func NewApp() *App {
	app := tview.NewApplication()
	list := tview.NewList()

	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetWrap(true).
		SetBorder(true).
		SetTitle(" Chat ")
	textView.SetText("initial text")

	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})
	list.ShowSecondaryText(false)

	flex := tview.NewFlex().
		AddItem(list, 0, 1, true).
		AddItem(textView, 0, 2, false)

	a := &App{
		app:      app,
		flex:     flex,
		list:     list,
		textView: textView,
	}
	list.AddItem("Fetch", "Get messages", 'm', func() {
		a.GetMessages()
	})
	return a
}

func (a *App) Run() {
	if err := a.app.SetRoot(a.flex, true).SetFocus(a.list).Run(); err != nil {
		panic(err)
	}
}
