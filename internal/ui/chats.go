package ui

import (
	"log"
	"mchat/internal/models"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type chatsModel struct {
	chats []*models.Chat

	contactsList     list.Model
	messagesViewport viewport.Model
	textInput        textinput.Model
}

var (
	chatStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("62")).
			PaddingBottom(2).
			PaddingRight(2).
			MarginLeft(1)
	inputStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("62")).
			Padding(1)
)

var chatFocusedStyle = chatStyle.
	Border(lipgloss.NormalBorder(), false, true, false, false)

func initChatsModel() chatsModel {
	contacts := list.New([]list.Item{}, list.NewDefaultDelegate(), 40, 40)
	contacts.SetShowStatusBar(false)
	contacts.SetFilteringEnabled(false)
	contacts.SetShowTitle(false)
	contacts.SetShowHelp(false)

	messages := viewport.New(40, 40)
	messages.SetContent("Select a chat")

	input := textinput.New()
	input.Placeholder = "Send a message..."

	return chatsModel{
		contactsList:     contacts,
		messagesViewport: messages,
		textInput:        input,
	}
}

func (m model) viewChat() string {
	var view string
	if m.focus == focusChat {
		view = chatFocusedStyle.Render(m.chats.messagesViewport.View())
		input := inputStyle.Render(m.chats.textInput.View())
		view = lipgloss.JoinVertical(lipgloss.Left, view, input)
	} else {
		view = chatStyle.MarginBottom(3).Render(m.chats.messagesViewport.View())
	}
	if m.focus == focusMessageInput {
		input := inputStyle.Border(lipgloss.RoundedBorder()).Padding(0).Render(m.chats.textInput.View())
		view = lipgloss.JoinVertical(lipgloss.Left, view, input)
	}
	return view
}

func (m model) viewChats() string {
	content := lipgloss.JoinHorizontal(lipgloss.Top, m.chats.contactsList.View(), m.viewChat())
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

func (m model) updateMessages(chat *models.Chat) model {
	content := ""
	for _, msg := range chat.Messages {
		content += msg.Date + "\n\n" + msg.Content + "\n\n----------------\n\n"
	}
	m.chats.messagesViewport.SetContent(content)

	return m
}

func (m model) updateChats(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Println(msg)
		w := msg.Width
		m.chats.contactsList.SetWidth(w / 3)
		m.chats.messagesViewport.Width = w - w/3 // TODO: it doesn't fully work
		m.chats.textInput.Width = w - w/3 - 2

		m.chats.contactsList.SetHeight(msg.Height - 20)
		m.chats.messagesViewport.Height = msg.Height - 20

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
				if len(m.chats.contactsList.Items()) > 0 {
					m.focus = focusChat
					index := m.chats.contactsList.Index()
					m = m.updateMessages(m.chats.chats[index])
					m.chats.messagesViewport.GotoBottom()
					return m, nil
				}
			}
			m.chats.contactsList, cmd = m.chats.contactsList.Update(msg)
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
			case "enter":
				m.focus = focusMessageInput
				m.chats.textInput.Focus()
				return m, nil
			}
			m.chats.messagesViewport, cmd = m.chats.messagesViewport.Update(msg)
			return m, cmd

		case focusMessageInput:
			switch msg.String() {
			case "esc":
				m.chats.textInput.Blur()
				m.focus = focusChat
				return m, nil
			case "enter":
				index := m.chats.contactsList.Index()
				err := m.svc.SendMessage(m.chats.chats[index], m.chats.textInput.Value())
				if err != nil {
					// TODO - handle error
					log.Println(err)
				}
				// TODO - handle with data storage, very temporary testing now
				m.chats.chats[index].Messages = append(m.chats.chats[index].Messages, &models.Message{
					Content: m.chats.textInput.Value(),
					Date:    "Now",
				})
				m = m.updateMessages(m.chats.chats[index])
				m.chats.messagesViewport.GotoBottom()
				m.focus = focusChat
				m.chats.textInput.SetValue("")
				m.chats.textInput.Blur()

				return m, nil
			}
			m.chats.textInput, cmd = m.chats.textInput.Update(msg)
			return m, cmd
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
	m.chats.contactsList.SetItems(items)
	m.chats.chats = chats
	return m
}
