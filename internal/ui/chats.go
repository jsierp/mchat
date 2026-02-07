package ui

import (
	"log"
	"mchat/internal/models"
	"strings"

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
			BorderForeground(colPrimaryMuted).
			Padding(2)
	inputStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Margin(1)
)

var chatFocusedStyle = chatStyle.
	Border(lipgloss.ThickBorder(), false, false, false, true).Padding(1)

func initChatsModel() chatsModel {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(colPrimary).BorderForeground(colPrimary)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(colPrimaryMuted).BorderForeground(colPrimary)

	contacts := list.New([]list.Item{}, d, 40, 40)
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
		view = chatStyle.Render(m.chats.messagesViewport.View())
	}
	if m.focus == focusMessageInput {
		input := inputStyle.Border(lipgloss.RoundedBorder()).Padding(0).Render(m.chats.textInput.View())
		view = lipgloss.JoinVertical(lipgloss.Left, view, input)
	}
	return view
}

func (m model) viewChats() string {
	list := lipgloss.NewStyle().MarginRight(1).Render(m.chats.contactsList.View())
	content := lipgloss.JoinHorizontal(lipgloss.Top, list, m.viewChat())
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
	inBorder := lipgloss.RoundedBorder()
	inBorder.BottomLeft = "┴"
	outBorder := lipgloss.RoundedBorder()
	outBorder.BottomRight = "┴"

	inMsgStyle := lipgloss.NewStyle().
		BorderForeground(colMuted).
		Border(inBorder, true, true, true, true).
		Padding(0, 1).
		MarginTop(1)

	outMsgStyle := lipgloss.NewStyle().
		BorderForeground(colSuccess).
		Border(outBorder, true, true, true, true).
		Padding(0, 1).
		MarginTop(1)

	content := ""
	for _, msg := range chat.Messages {
		text := msg.Content
		var msgBubble string

		dateText := msg.Date.Format("Mon, 15:04")
		if strings.Contains(msg.Id, "mchat") { // TODO: tmp solution for outgoing msgs
			msgBubble = outMsgStyle.MarginLeft(60).Render(text)
			msgBubble = lipgloss.JoinVertical(lipgloss.Right, msgBubble, lipgloss.NewStyle().Foreground(colMuted).Render(dateText))
		} else {
			msgBubble = inMsgStyle.Render(text)
			msgBubble = lipgloss.JoinVertical(lipgloss.Left, msgBubble, lipgloss.NewStyle().Foreground(colMuted).Render(dateText))
		}
		content = lipgloss.JoinVertical(lipgloss.Left, content, msgBubble)
	}
	m.chats.messagesViewport.SetContent(content)

	return m
}

func (m model) updateChats(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
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
				msg, err := m.svc.SendMessage(m.chats.chats[index], m.chats.textInput.Value())
				if err != nil {
					// TODO - handle error
					log.Println(err)
				}
				m = m.newMessage(msg)
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

func (m model) newMessage(msg *models.Message) model {
	index := m.chats.contactsList.Index()
	for i, c := range m.chats.chats {
		if c.Address == msg.ChatAddress {
			c.Messages = appendIfNew(c.Messages, msg)
			if i == index {
				m = m.updateMessages(c)
				m.chats.messagesViewport.GotoBottom()
			}
			return m
		}
	}
	c := models.Chat{Address: msg.ChatAddress, Name: msg.Contact, Messages: []*models.Message{msg}}
	m.chats.chats = append(m.chats.chats, &c)

	items := m.chats.contactsList.Items()
	items = append(items, listItem{
		title:       c.Name,
		description: c.Address,
	})
	m.chats.contactsList.SetItems(items)

	return m
}

func appendIfNew(msgs []*models.Message, msg *models.Message) []*models.Message {
	for _, m := range msgs {
		if m.Id == msg.Id {
			return msgs
		}
	}
	return append(msgs, msg)
}
