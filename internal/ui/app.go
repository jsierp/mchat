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
	SendMessage(chat *models.Chat, msg string) (*models.Message, error)
}

var (
	// TODO - light palette needs rework
	colPrimary      = lipgloss.AdaptiveColor{Light: "#81c8be", Dark: "#81c8be"}
	colPrimaryMuted = lipgloss.AdaptiveColor{Light: "#528a82", Dark: "#528a82"}
	colSuccess      = lipgloss.AdaptiveColor{Light: "#a3be8c", Dark: "#a3be8c"}
	colWarning      = lipgloss.AdaptiveColor{Light: "#ebcb8b", Dark: "#ebcb8b"}
	colDanger       = lipgloss.AdaptiveColor{Light: "#bf616a", Dark: "#bf616a"}
	colMuted        = lipgloss.AdaptiveColor{Light: "#777777", Dark: "#777777"}
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(colSuccess).
			MarginBottom(1).
			MarginLeft(1).
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
