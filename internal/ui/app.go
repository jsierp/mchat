package ui

import (
	"mchat/internal/models"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/oauth2"
)

type DataService interface {
	SaveBasicConfig(user, pass string)
	SaveGoogleConfig(user string, token *oauth2.Token)
	IsConfigured() bool
	GetChats() []*models.Chat
	SendMessage(chat *models.Chat, msg string) error
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
	input.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEnter {
			text := el.GetText(true)
			msg := input.GetText()

			text += time.Now().Format("2006-01-02 15:04") + "\n\nYou: " + msg + "\n\n----------------\n\n"
			a.svc.SendMessage(chat, msg)
			el.SetText(text)
			input.SetText("", true)
			return nil
		}
		return e
	})

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

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			MarginBottom(1).
			Bold(true)
	controlsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginLeft(2).
			MarginTop(1)
	chatStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("12")).
			PaddingLeft(2)
)

type mode int

const (
	modeConfig mode = iota
	modeList
	modeChat
	modeMessage
)

type model struct {
	mode      mode
	svc       DataService
	chatsList list.Model
	chatView  viewport.Model
	chats     []*models.Chat
	width     int
	height    int
}

// View implements [tea.Model].
type chatItem struct {
	name     string
	address  string
	selected bool
}

func (c chatItem) FilterValue() string { return c.name + c.address }

func (c chatItem) Title() string { return c.name }

func (c chatItem) Description() string { return c.address }

func InitialModel(svc DataService) model {
	chatsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 40, 40)
	chatsList.SetShowStatusBar(false)
	chatsList.SetFilteringEnabled(false)
	chatsList.SetShowTitle(false)

	chatView := viewport.New(40, 40)
	chatView.SetContent("Select a chat")

	return model{
		mode:      modeList,
		svc:       svc,
		chatsList: chatsList,
		chatView:  chatView,
	}

}

func (m model) viewControls() string {
	s := controlsStyle.Render("\nControls:")
	s += controlsStyle.Render("• r: refresh")
	s += controlsStyle.Render("• a: add a chat")
	s += controlsStyle.Render("• c: enter config")
	s += controlsStyle.Render("• q: quit")
	return s
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) refresh() model {
	chats := m.svc.GetChats()
	var items []list.Item
	for _, chat := range chats {
		items = append(items, chatItem{
			name:    chat.Contact.Name,
			address: chat.Contact.Address,
		})
	}
	m.chatsList.SetItems(items)
	m.chats = chats
	return m
}

func (m model) renderChat(chat *models.Chat) model {
	content := ""
	for _, msg := range chat.Messages {
		content += msg.Date + "\n\n" + msg.Content + "\n\n----------------\n\n"
	}
	m.chatView.SetContent(content)

	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.chatView.Width = 80
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m = m.refresh()
			return m, nil
		case "l":
			if m.mode == modeList {
				m.mode = modeChat
			}
			return m, nil
		case "h":
			if m.mode == modeChat {
				m.mode = modeList
			}
			return m, nil
		case "enter", " ":
			if len(m.chatsList.Items()) > 0 {
				index := m.chatsList.Index()
				m = m.renderChat(m.chats[index])
			}
		}
	}

	var cmd tea.Cmd

	switch m.mode {
	case modeChat:
		m.chatView, cmd = m.chatView.Update(msg)
	case modeList:
		m.chatsList, cmd = m.chatsList.Update(msg)
	}

	return m, cmd
}

func (m model) viewChat() string {
	if m.mode == modeChat {
		return chatStyle.
			Border(lipgloss.NormalBorder(), false, false, false, true).
			Render(m.chatView.View())
	}
	return chatStyle.Render(m.chatView.View())
}

func (m model) View() string {
	title := titleStyle.Render("mChat - the only messaging app you need")

	body := lipgloss.JoinHorizontal(lipgloss.Top, m.chatsList.View(), m.viewChat())
	view := lipgloss.JoinVertical(lipgloss.Left, title, body)

	view += m.viewControls()
	return view
}
