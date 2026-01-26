package ui

import (
	"mchat/internal/auth_google"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/oauth2"
)

var cfgListStyle = lipgloss.NewStyle().
	MarginTop(2)

const (
	mchatTitle     = "mChat"
	googleTitle    = "Google"
	microsoftTitle = "Microsoft"
)

type configItem struct {
	title, description string
	selected           bool
}

type configFocus int

const (
	cfgFocusEmailInput configFocus = iota
	cfgFocusList
	cfgFocusUrl
)

type configModel struct {
	focus configFocus

	list       list.Model
	emailInput textinput.Model

	googleUrl string
}

func (i configItem) Title() string {
	if i.selected {
		return "✓ " + i.title
	}
	return i.title
}
func (i configItem) Description() string { return i.description }
func (i configItem) FilterValue() string { return i.title }

func initConfigModel() configModel {
	cfglist := list.New([]list.Item{
		configItem{
			title:       mchatTitle,
			description: "Use MChat Account",
		},
		configItem{
			title:       googleTitle,
			description: "Use Gmail Account",
		},
		configItem{
			title:       microsoftTitle,
			description: "Use Outlook Account",
		},
	}, list.NewDefaultDelegate(), 40, 10)
	cfglist.SetShowStatusBar(false)
	cfglist.SetFilteringEnabled(false)
	cfglist.SetShowTitle(false)
	cfglist.SetShowHelp(false)

	ei := textinput.New()
	ei.Placeholder = "Enter your e-mail address"
	ei.Focus()
	ei.Width = 50

	return configModel{
		list:       cfglist,
		emailInput: ei,
	}
}

func viewURL(url string) string {
	url = lipgloss.NewStyle().Underline(true).Render(url)
	urlText := lipgloss.NewStyle().Width(80).Render("Visit: " + url)
	return urlText
}

func (m model) viewConfig() string {
	email := m.cfg.emailInput.View()
	l := cfgListStyle.Render(m.cfg.list.View())
	view := lipgloss.JoinVertical(lipgloss.Top, email, l)

	if m.cfg.focus == cfgFocusUrl {
		i := m.cfg.list.SelectedItem()
		url := " Option not implemented yet! Use Google provider."
		if i, ok := i.(configItem); ok && (i.title == googleTitle) {
			url = viewURL(m.cfg.googleUrl)
		}
		view = lipgloss.JoinVertical(lipgloss.Top, view, url)
	}
	view += m.helpConfig()
	return view
}

func (m model) helpConfig() string {
	s := helpStyle.Render("\nControls:")
	s += helpStyle.Render("• esc: back to chats list")
	s += helpStyle.Render("• q: quit")
	return s
}

func (m model) leaveConfig() (model, tea.Cmd) {
	m.cfg = initConfigModel()
	m.view = viewChats
	return m, nil
}

type authResult struct {
	token *oauth2.Token
	err   error
}

func awaitAuthCode(svc *auth_google.GoogleAuthService) func() tea.Msg {
	return func() tea.Msg {
		code := svc.WaitForAuthCode()
		token, err := svc.ExchangeCode(code)
		return authResult{token: token, err: err}
	}
}

func (m model) updateConfig(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil
	case tea.KeyMsg:
		switch m.cfg.focus {

		case cfgFocusEmailInput:
			switch msg.String() {
			case "esc":
				return m.leaveConfig()
			case "enter":
				m.cfg.focus = cfgFocusList
				m.cfg.emailInput.Blur()
				return m, nil
			}
			m.cfg.emailInput, cmd = m.cfg.emailInput.Update(msg)
			return m, cmd

		case cfgFocusList:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "esc":
				return m.leaveConfig()
			case "enter":
				index := m.cfg.list.Index()
				items := m.cfg.list.Items()
				if item, ok := items[index].(configItem); ok {
					item.selected = true
					items[index] = item
					m.cfg.focus = cfgFocusUrl
				}
				m.cfg.list.SetItems(items)
				svc := auth_google.NewGoogleAuthService()
				m.cfg.googleUrl = svc.GetGoogleUrl()
				return m, awaitAuthCode(svc)
			}
			m.cfg.list, cmd = m.cfg.list.Update(msg)
			return m, cmd

		case cfgFocusUrl:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "esc":
				return m.leaveConfig()
			}
		}
	case authResult:
		if msg.err != nil {
			m.cfg.googleUrl = "An error occured! Try again by restarting configuration!"
			return m, nil
		} else {
			m.svc.SaveGoogleConfig(m.cfg.emailInput.Value(), msg.token)
			m.cfg.googleUrl = "Success! Go back to the chats list with ESC."
		}
	}

	return m, nil
}
