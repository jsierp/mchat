package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.view = viewChats
		case "c":
			m.view = viewConfig
		}
	}

	return m, nil
}

func (m model) viewHelp() string {
	help := "Controls:\n"
	help += "• ←,↑,→,↓ or h,k,j,l: list navigation, scrolling\n"
	help += "• enter or tab: confirm\n"
	help += "• esc or shift+tab: go back\n"
	help += "• c: enter config\n"
	help += "• r: refresh (not implemented yet)\n"
	help += "• a: add a chat (not implemented yet)\n"
	help += "• q: quit\n"
	help += "\n"
	help += "Instructions:\n"
	help += "Configure email account with Google or Microsoft providers.\n"
	help += "Mchat server is not implemented yet."

	w := lipgloss.Width(help)
	h := lipgloss.Height(help)
	style := lipgloss.NewStyle().PaddingLeft(m.width/2 - w/2).PaddingTop(m.height/2 - h/2).Height(m.height - 4)
	return style.Render(help) + m.viewHelpBar("Press ESC to go to chats")
}
