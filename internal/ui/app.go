package ui

import (
	"mchat/internal/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/oauth2"
)

type DataService interface {
	SaveBasicConfig(user, pass string)
	SaveGoogleConfig(user string, token *oauth2.Token)
	IsConfigured() bool
	GetChats() []*models.Chat
}

type App struct {
	app   *tview.Application
	pages *tview.Pages
	flex  *tview.Flex
	list  *tview.List
	chat  *tview.Flex
	svc   DataService
}

func (a *App) RenderChat(chat *models.Chat) {
	msgs := chat.Messages
	a.chat.Clear()
	el := tview.NewTextView()
	el.SetDynamicColors(true).
		SetWrap(true).
		SetBorder(true).
		SetTitle("Chat with " + chat.Contact.Name)
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
	chats := a.svc.GetChats()
	a.list.Clear()
	for _, chat := range chats {
		a.list.AddItem(chat.Contact.Name, chat.Contact.Address, 0, func() {
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

func NewApp(svc DataService) *App {
	pages := tview.NewPages()
	app := tview.NewApplication()

	a := &App{app: app, pages: pages, svc: svc}

	pages.AddPage("config", a.initConfigPage(), true, true).
		AddPage("inbox", a.initInboxPage(), true, false)

	if svc.IsConfigured() {
		pages.SwitchToPage("inbox")
	}

	a.app.SetRoot(pages, true)

	return a
}

func (a *App) Run() error {
	if err := a.app.Run(); err != nil {
		return err
	}
	return nil
}
