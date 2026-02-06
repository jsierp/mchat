package ui

import (
	"mchat/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/oauth2"
)

type DataService interface {
	SaveBasicConfig(user, pass string)
	SaveGoogleConfig(user string, token *oauth2.Token)
	IsConfigured() bool
	SendMessage(chat *models.Chat, msg string) (*models.Message, error)
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

	chats chatsModel
	cfg   configModel
}

// View implements [tea.Model].
type listItem struct {
	title, description string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.title }

func InitialModel(svc DataService) model {
	return model{
		focus: focusChats,
		view:  viewChats,
		svc:   svc,
		chats: initChatsModel(),
		cfg:   initConfigModel(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if msg, ok := msg.(*models.Message); ok {
		return m.newMessage(msg), nil
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
