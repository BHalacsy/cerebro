package main

import (
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	tea "github.com/charmbracelet/bubbletea"

	anthropicClient "github.com/BHalacsy/cerebro/src/external/anthropic"
	"github.com/BHalacsy/cerebro/src/tui"
)

func main() {
	agent := anthropicClient.NewAgent(
		anthropic.ModelClaudeSonnet4_6,
		"You are Cerebro, a helpful AI assistant. Be concise and clear.",
		nil,
	)

	m := tui.NewModel(agent)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
