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
	colSuccessMuted = lipgloss.AdaptiveColor{Light: "#7a8c6d", Dark: "#7a8c6d"}
	colWarning      = lipgloss.AdaptiveColor{Light: "#ebcb8b", Dark: "#ebcb8b"}
	colDanger       = lipgloss.AdaptiveColor{Light: "#bf616a", Dark: "#bf616a"}
	colMuted        = lipgloss.AdaptiveColor{Light: "#777777", Dark: "#777777"}
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(colSuccess).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colSuccess).
			PaddingLeft(1).
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

	chats chatsModel
	cfg   configModel
}

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
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
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
	title := titleStyle.Width(m.width).Render("mChat - the only messaging app you need")
	var viewContent string

	switch m.view {
	case viewChats:
		viewContent = m.viewChats()
	case viewConfig:
		viewContent = m.viewConfig()
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, viewContent)
}
