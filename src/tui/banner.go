package tui

import "github.com/charmbracelet/lipgloss"

const banner = `
   ██████╗███████╗██████╗ ███████╗██████╗ ██████╗  ██████╗
  ██╔════╝██╔════╝██╔══██╗██╔════╝██╔══██╗██╔══██╗██╔═══██╗
  ██║     █████╗  ██████╔╝█████╗  ██████╔╝██████╔╝██║   ██║
  ██║     ██╔══╝  ██╔══██╗██╔══╝  ██╔══██╗██╔══██╗██║   ██║
  ╚██████╗███████╗██║  ██║███████╗██████╔╝██║  ██║╚██████╔╝
   ╚═════╝╚══════╝╚═╝  ╚═╝╚══════╝╚═════╝ ╚═╝  ╚═╝ ╚═════╝`

var bannerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#7C3AED")).
	Bold(true)

func RenderBanner() string {
	return bannerStyle.Render(banner)
}
