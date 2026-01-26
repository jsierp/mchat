package ui

import (
	"mchat/internal/models"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var chatStyle = lipgloss.NewStyle().
	BorderForeground(lipgloss.Color("12")).
	PaddingLeft(2)

var chatFocusedStyle = chatStyle.
	Border(lipgloss.NormalBorder(), false, false, false, true)

func (m model) viewChat() string {
	if m.focus == focusChat {
		return chatFocusedStyle.Render(m.chatViewport.View())
	}
	return chatStyle.Render(m.chatViewport.View())
}

func (m model) viewChats() string {
	content := lipgloss.JoinHorizontal(lipgloss.Top, m.chatsList.View(), m.viewChat())
	content += m.helpChats()
	return content
}

func (m model) helpChats() string {
	s := helpStyle.Render("\nControls:")
	s += helpStyle.Render("• r: refresh")
	s += helpStyle.Render("• a: add a chat")
	s += helpStyle.Render("• c: enter config")
	s += helpStyle.Render("• q: quit")
	return s
}

func (m model) updateChatViewport(chat *models.Chat) model {
	content := ""
	for _, msg := range chat.Messages {
		content += msg.Date + "\n\n" + msg.Content + "\n\n----------------\n\n"
	}
	m.chatViewport.SetContent(content)

	return m
}

func (m model) updateChats(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// set sizes
		return m, nil
	case tea.KeyMsg:
		switch m.focus {

		case focusChats:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "r":
				return m.refreshChats(), nil
			case "c":
				m.view = viewConfig
				return m, nil
			case "enter", " ", "l":
				if len(m.chatsList.Items()) > 0 {
					m.focus = focusChat
					index := m.chatsList.Index()
					m = m.updateChatViewport(m.chats[index])
					return m, nil
				}
			}
			m.chatsList, cmd = m.chatsList.Update(msg)
			return m, cmd

		case focusChat:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "h":
				m.focus = focusChats
				return m, nil
			case "r":
				return m.refreshChats(), nil
			case "c":
				m.view = viewConfig
				return m, nil
			case "esc":
				m.focus = focusChats
				return m, nil
			}
			m.chatViewport, cmd = m.chatViewport.Update(msg)
			return m, cmd

		case focusMessageInput:
			switch msg.String() {
			case "esc":
				m.focus = focusChat
				return m, nil
			}
			// update text input
			return m, nil
		}
	}

	return m, nil
}

func (m model) refreshChats() model {
	chats := m.svc.GetChats()
	var items []list.Item
	for _, chat := range chats {
		items = append(items, listItem{
			title:       chat.Contact.Name,
			description: chat.Contact.Address,
		})
	}
	m.chatsList.SetItems(items)
	m.chats = chats
	return m
}
