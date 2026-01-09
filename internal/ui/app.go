package ui

import (
	"log"
	"os"

	"mchat/internal/data"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	app  *tview.Application
	flex *tview.Flex
	list *tview.List
	chat *tview.Flex
}

func (a *App) RenderChat(chat *data.Chat) {
	msgs := chat.Messages
	a.chat.Clear()
	el := tview.NewTextView()
	el.SetDynamicColors(true).
		SetWrap(true).
		SetBorder(true).
		SetTitle("Chat with " + chat.Contact)
	el.ScrollToEnd()

	text := ""
	for _, msg := range msgs {
		text += msg.Date + "\n\n" + msg.Content + "\n\n----------------\n\n"
	}

	el.SetText(text)

	input := tview.NewTextArea()
	input.SetBorder(true).SetTitle("Send Message")

	a.chat.AddItem(el, 0, 1, false)
	a.chat.AddItem(input, 5, 1, true)
}

func (a *App) GetMessages() {
	chats := data.GetChats()
	a.list.Clear()
	for _, chat := range chats {
		msg := chat.Messages[0]
		a.list.AddItem(msg.From.Name, msg.From.Address, 0, func() {
			a.RenderChat(chat)
			a.app.SetFocus(a.chat)
		})
	}
	a.initList()
}

func (a *App) initList() {
	a.list.AddItem("Fetch", "Get messages", 'm', func() {
		a.GetMessages()
	}).AddItem("Quit", "Press to exit", 'q', func() {
		a.app.Stop()
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
	chat := tview.NewFlex()
	chat.SetDirection(tview.FlexRow)
	chat.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyESC || e.Key() == tcell.KeyLeft {
			app.SetFocus(list)
		}
		return e
	})

	flex := tview.NewFlex().
		AddItem(list, 0, 1, true).
		AddItem(chat, 0, 2, false)

	a := &App{
		app:  app,
		flex: flex,
		list: list,
		chat: chat,
	}
	a.initList()
	return a
}

func (a *App) Run() {
	if err := a.app.SetRoot(a.flex, true).SetFocus(a.list).Run(); err != nil {
		panic(err)
	}
}
