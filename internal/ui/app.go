package ui

import (
	"log"
	"os"

	"mchat/internal/config"
	"mchat/internal/data"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	app   *tview.Application
	cfg   *config.Config
	pages *tview.Pages
	flex  *tview.Flex
	list  *tview.List
	chat  *tview.Flex
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
	chats := data.GetChats(a.cfg)
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

func setLogsFile() {
	file, err := os.OpenFile("mchat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func (a *App) initInboxPage() *tview.Flex {
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
			a.app.SetFocus(list)
		}
		return e
	})

	flex := tview.NewFlex().
		AddItem(list, 0, 1, true).
		AddItem(chat, 0, 2, false)

	a.flex = flex
	a.list = list
	a.chat = chat
	a.initList()

	return flex
}

func NewApp() *App {
	setLogsFile()
	cfg, err := config.LoadConfig()

	if err != nil {
		panic(err)
	}
	pages := tview.NewPages()
	app := tview.NewApplication()

	a := &App{app: app, pages: pages, cfg: cfg}

	pages.AddPage("config", a.initConfigPage(), true, true).
		AddPage("inbox", a.initInboxPage(), true, false)

	if cfg.Login != "" || cfg.Google {
		pages.SwitchToPage("inbox")
	}

	a.app.SetRoot(pages, true)

	return a
}

func (a *App) Run() {
	if err := a.app.Run(); err != nil {
		panic(err)
	}
}
