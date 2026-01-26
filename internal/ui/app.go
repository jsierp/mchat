package ui

import (
	"mchat/internal/models"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/oauth2"
)

type DataService interface {
	SaveBasicConfig(user, pass string)
	SaveGoogleConfig(user string, token *oauth2.Token)
	IsConfigured() bool
	GetChats() []*models.Chat
	SendMessage(chat *models.Chat, msg string) error
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			MarginBottom(1).
			Bold(true)
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginLeft(2).
			MarginTop(1)
)

type view int

const (
	viewChats view = iota
	viewConfig
)

type focus int

const (
	focusChats focus = iota
	focusChat
	focusMessageInput
)

type model struct {
	svc DataService

	view  view
	focus focus

	width  int
	height int

	chats []*models.Chat

	chatsList    list.Model
	chatViewport viewport.Model

	cfg configModel
}

// View implements [tea.Model].
type listItem struct {
	title, description string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.title }

func InitialModel(svc DataService) model {
	chatsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 40, 40)
	chatsList.SetShowStatusBar(false)
	chatsList.SetFilteringEnabled(false)
	chatsList.SetShowTitle(false)

	chatViewport := viewport.New(40, 40)
	chatViewport.SetContent("Select a chat")

	return model{
		focus:        focusChats,
		view:         viewChats,
		svc:          svc,
		chatsList:    chatsList,
		chatViewport: chatViewport,
		cfg:          initConfigModel(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.view {
	case viewChats:
		return m.updateChats(msg)
	case viewConfig:
		return m.updateConfig(msg)
	}
	return m, nil
}

func (m model) View() string {
	title := titleStyle.Render("mChat - the only messaging app you need")
	var viewContent string

	switch m.view {
	case viewChats:
		viewContent = m.viewChats()
	case viewConfig:
		viewContent = m.viewConfig()
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, viewContent)
}
