package ui

import (
	"log"
	"mchat/internal/models"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type contactItem struct {
	title, description string
	selected           bool
}

func (i contactItem) Title() string {
	if i.selected {
		return "→ " + i.title
	}
	return i.title
}
func (i contactItem) Description() string {
	if i.selected {
		return "  " + i.description
	}
	return i.description
}
func (i contactItem) FilterValue() string { return i.title }

type sendMessageResult struct {
	msg *models.Message
	err error
}

type chatsModel struct {
	chats []*models.Chat

	contactsList     list.Model
	messagesViewport viewport.Model
	textInput        textinput.Model
}

var (
	chatStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).Padding(0, 0, 0, 1).
			BorderForeground(colMuted)
	inputStyle = lipgloss.NewStyle().
			BorderForeground(colPrimary).
			Padding(1)
	inMsgStyle  = getInMsgStyle()
	outMsgStyle = getOutMsgStyle()
)

var chatFocusedStyle = chatStyle.
	Border(lipgloss.ThickBorder(), false, false, false, true).Padding(0, 0, 0, 1).
	BorderForeground(colPrimaryMuted)

func getInMsgStyle() lipgloss.Style {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = "┴"

	return lipgloss.NewStyle().
		BorderForeground(colMuted).
		Border(border, true, true, true, true).
		Padding(0, 1)
}

func getOutMsgStyle() lipgloss.Style {
	outBorder := lipgloss.RoundedBorder()
	outBorder.BottomRight = "┴"

	return lipgloss.NewStyle().
		BorderForeground(colSuccess).
		Border(outBorder, true, true, true, true).
		Padding(0, 1)
}

func listDelegate(selected bool) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	var col, colDesc lipgloss.AdaptiveColor
	if selected {
		col = colSuccess
		colDesc = colSuccessMuted
	} else {
		col = colPrimary
		colDesc = colPrimaryMuted
	}
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(col).BorderForeground(col)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(colDesc).BorderForeground(col)
	return d
}

func initChatsModel() chatsModel {
	contacts := list.New([]list.Item{}, listDelegate(false), 40, 40)
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

func isAtBottom(v viewport.Model) bool {
	return v.YOffset >= v.TotalLineCount()-v.Height
}

func (m model) viewScroll() string {
	if isAtBottom(m.chats.messagesViewport) {
		return ""
	}
	w := m.chats.messagesViewport.Width
	return lipgloss.NewStyle().Foreground(colMuted).PaddingLeft(w/2 - 9).Render("↓ new messages ↓")
}

func (m model) viewChat() string {
	var view string
	if m.focus == focusChat {
		view = chatFocusedStyle.Render(m.chats.messagesViewport.View())
	} else {
		view = chatStyle.Render(m.chats.messagesViewport.View())
	}

	view = lipgloss.JoinVertical(lipgloss.Left, view, m.viewScroll())

	var input string
	if m.focus == focusMessageInput {
		input = inputStyle.Border(lipgloss.RoundedBorder()).Padding(0).Render(m.chats.textInput.View())
	} else {
		input = inputStyle.Render(m.chats.textInput.View())
	}
	view = lipgloss.JoinVertical(lipgloss.Left, view, input)
	return view
}

func (m model) viewChats() string {
	list := m.chats.contactsList.View()
	width := lipgloss.Width(list)
	list = lipgloss.NewStyle().PaddingRight(m.chats.contactsList.Width() - width).Render(list)
	content := lipgloss.JoinHorizontal(lipgloss.Top, list, m.viewChat())
	content += m.viewHelpBar("Press ? for help")
	return content
}

func messageStatusBar(m *models.Message) string {
	dateText := m.Date.Format("Mon, 15:04")
	bar := lipgloss.NewStyle().Foreground(colMuted).Render(dateText)
	if m.ChatAddress != m.From {
		switch m.Status {
		case models.MsgStatusSuccess:
			bar += " " + lipgloss.NewStyle().Foreground(colSuccess).Render("✓ ")
		case models.MsgStatusSending:
			bar += " " + lipgloss.NewStyle().Foreground(colWarning).Render("🗘 ")
		case models.MsgStatusError:
			bar += " " + lipgloss.NewStyle().Foreground(colDanger).Render("‼ ")
		}
	}
	return bar
}

func (m model) updateMessages(chat *models.Chat) model {
	content := ""
	for _, msg := range chat.Messages {
		text := msg.Content
		var msgBubble string
		msgWidth := min(lipgloss.Width(text)+2, m.chats.messagesViewport.Width/10*9)

		if msg.ChatAddress != msg.From {
			msgBubble = outMsgStyle.Width(msgWidth).Render(text)
			msgBubble = lipgloss.JoinVertical(lipgloss.Right, msgBubble, messageStatusBar(msg))
			msgBubble = lipgloss.NewStyle().Width(m.chats.messagesViewport.Width - 2).Align(lipgloss.Right).Render(msgBubble)
		} else {
			msgBubble = inMsgStyle.Width(msgWidth).Render(text)
			msgBubble = lipgloss.JoinVertical(lipgloss.Left, msgBubble, messageStatusBar(msg))
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
		m.chats.contactsList.SetWidth(w / 4)
		m.chats.messagesViewport.Width = w - w/4
		m.chats.textInput.Width = w - w/4 - 5

		m.chats.contactsList.SetHeight(msg.Height - 10)
		m.chats.messagesViewport.Height = msg.Height - 8

		index := m.chats.contactsList.Index()
		if len(m.chats.chats) > 0 {
			m = m.updateMessages(m.chats.chats[index])
		}

		return m, nil

	case sendMessageResult:
		if msg.err != nil {
			log.Println("error while sending the message", msg.err)
			msg.msg.Status = models.MsgStatusError
		} else {
			msg.msg.Status = models.MsgStatusSuccess
		}
		index := m.chats.contactsList.Index()
		if len(m.chats.chats) > 0 && m.chats.chats[index].Address == msg.msg.ChatAddress {
			m = m.updateMessages(m.chats.chats[index])
		}
		return m, nil

	case tea.KeyMsg:
		switch m.focus {

		case focusChats:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "?":
				m.view = viewHelp
				return m, nil
			case "c":
				m.view = viewConfig
				return m, nil
			case "enter", " ", "l", "right", "tab":
				if len(m.chats.contactsList.Items()) > 0 {
					m.focus = focusChat
					index := m.chats.contactsList.Index()

					items := m.chats.contactsList.Items()
					if item, ok := items[index].(contactItem); ok {
						item.selected = true
						items[index] = item
					}
					m.chats.contactsList.SetDelegate(listDelegate(true))
					m.chats.contactsList.SetItems(items)

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
			case "?":
				m.view = viewHelp
				return m, nil
			case "h", "left", "esc", "shift+tab":
				m.focus = focusChats

				index := m.chats.contactsList.Index()

				items := m.chats.contactsList.Items()
				if item, ok := items[index].(contactItem); ok {
					item.selected = false
					items[index] = item
				}
				m.chats.contactsList.SetDelegate(listDelegate(false))
				m.chats.contactsList.SetItems(items)

				return m, nil
			case "g":
				m.chats.messagesViewport.GotoTop()
				return m, nil
			case "G":
				m.chats.messagesViewport.GotoBottom()
				return m, nil
			case "c":
				m.view = viewConfig
				return m, nil
			case "enter", "tab":
				m.focus = focusMessageInput
				m.chats.textInput.Focus()
				return m, nil
			}
			m.chats.messagesViewport, cmd = m.chats.messagesViewport.Update(msg)
			return m, cmd

		case focusMessageInput:
			switch msg.String() {
			case "esc", "shift+tab":
				m.chats.textInput.Blur()
				m.focus = focusChat
				return m, nil
			case "enter":
				content := m.chats.textInput.Value()
				content = strings.TrimSpace(content)
				if content == "" {
					return m, nil
				}
				index := m.chats.contactsList.Index()
				msg := prepareMessage(m.chats.chats[index], content)

				cmdSend := func() tea.Msg {
					err := m.svc.SendMessage(msg)
					return sendMessageResult{msg: msg, err: err}
				}
				m = m.newMessage(msg)
				m.chats.messagesViewport.GotoBottom()
				m.focus = focusChat
				m.chats.textInput.SetValue("")
				m.chats.textInput.Blur()

				return m, cmdSend
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
	items = append(items, contactItem{
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

func prepareMessage(c *models.Chat, s string) *models.Message {
	return &models.Message{
		To:          c.Address,
		Contact:     c.Name,
		ChatAddress: c.Address,
		Content:     s,
		Date:        time.Now(),
		Status:      models.MsgStatusSending,
	}
}
