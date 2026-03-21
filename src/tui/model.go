package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	anthropicClient "github.com/BHalacsy/cerebro/src/external/anthropic"
)

// Styles
var (
	userStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	assistantStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	borderStyle    = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(0, 1)
)

// Messages
type responseMsg struct {
	content string
	err     error
}

// Model is the main TUI model.
type Model struct {
	viewport viewport.Model
	textarea textarea.Model
	agent    *anthropicClient.Agent
	messages []chatMessage
	width    int
	height   int
	waiting  bool
	ready    bool
}

type chatMessage struct {
	role    string
	content string
}

func NewModel(agent *anthropicClient.Agent) Model {
	ta := textarea.New()
	ta.Placeholder = "Ask cerebro something..."
	ta.Focus()
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.KeyMap.InsertNewline.SetEnabled(true)

	return Model{
		textarea: ta,
		agent:    agent,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.waiting {
				return m, nil
			}
			val := m.textarea.Value()
			// Backslash + Enter: strip the trailing `\` and insert a newline
			if strings.HasSuffix(val, "\\") {
				m.textarea.SetValue(strings.TrimSuffix(val, "\\"))
				m.textarea.InsertString("\n")
				return m, nil
			}
			// Plain Enter: send the message
			input := strings.TrimSpace(val)
			if input == "" {
				return m, nil
			}
			m.textarea.Reset()
			m.messages = append(m.messages, chatMessage{role: "user", content: input})
			m.waiting = true
			m.updateViewport()
			m.viewport.GotoBottom()
			return m, m.sendMessage(input)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := strings.Count(banner, "\n") + 3 // banner + padding
		inputHeight := 5                                // textarea + border
		footerHeight := 1                               // help text

		vpHeight := m.height - headerHeight - inputHeight - footerHeight
		if vpHeight < 1 {
			vpHeight = 1
		}

		if !m.ready {
			m.viewport = viewport.New(m.width-2, vpHeight)
			m.ready = true
		} else {
			m.viewport.Width = m.width - 2
			m.viewport.Height = vpHeight
		}
		m.textarea.SetWidth(m.width - 6)
		m.updateViewport()

	case responseMsg:
		m.waiting = false
		if msg.err != nil {
			m.messages = append(m.messages, chatMessage{role: "error", content: msg.err.Error()})
		} else {
			m.messages = append(m.messages, chatMessage{role: "assistant", content: msg.content})
		}
		m.updateViewport()
		m.viewport.GotoBottom()
	}

	if !m.waiting {
		var taCmd tea.Cmd
		m.textarea, taCmd = m.textarea.Update(msg)
		cmds = append(cmds, taCmd)
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	header := RenderBanner()
	input := borderStyle.Width(m.width - 4).Render(m.textarea.View())

	status := ""
	if m.waiting {
		status = dimStyle.Render("  thinking...")
	}

	help := dimStyle.Render("  enter: send • esc/ctrl+c: quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		status,
		input,
		help,
	)
}

func (m *Model) updateViewport() {
	var sb strings.Builder
	for _, msg := range m.messages {
		switch msg.role {
		case "user":
			sb.WriteString(userStyle.Render("❯ You") + "\n")
			sb.WriteString("  " + strings.ReplaceAll(msg.content, "\n", "\n  ") + "\n\n")
		case "assistant":
			sb.WriteString(assistantStyle.Render("❯ Cerebro") + "\n")
			sb.WriteString("  " + strings.ReplaceAll(msg.content, "\n", "\n  ") + "\n\n")
		case "error":
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("❯ Error") + "\n")
			sb.WriteString("  " + msg.content + "\n\n")
		}
	}
	m.viewport.SetContent(sb.String())
}

func (m *Model) sendMessage(input string) tea.Cmd {
	agent := m.agent
	return func() tea.Msg {
		resp, err := agent.Run(context.Background(), input)
		return responseMsg{content: resp, err: err}
	}
}
