package zaplipgloss

import "github.com/charmbracelet/lipgloss"

const (
	colorDebug   = lipgloss.Color("#c678dd")
	colorInfo    = lipgloss.Color("#61afef")
	colorWarn    = lipgloss.Color("#d19a66")
	colorErr     = lipgloss.Color("#e06c75")
	colorDPanic  = lipgloss.Color("#7E1A53")
	colorPanic   = lipgloss.Color("#be5046")
	colorFatal   = lipgloss.Color("#282c34")
	colorUnknown = lipgloss.Color("#abb2bf")
	colorInvalid = lipgloss.Color("#ff0000")

	colorTextWhite = lipgloss.Color("#abb2bf")
	colorTextBlack = lipgloss.Color("#282c34")
)

var rainbow = []lipgloss.Color{
	lipgloss.Color("#e06c75"),
	lipgloss.Color("#d19a66"),
	lipgloss.Color("#e6c07b"),
	lipgloss.Color("#98c379"),
	lipgloss.Color("#56b6c2"),
	lipgloss.Color("#61aeee"),
	lipgloss.Color("#c678dd"),
}

func fieldcolor(i int) lipgloss.Color {
	return rainbow[i%len(rainbow)]
}

func fieldstyle(i int) lipgloss.Style {
	return lipgloss.NewStyle().Background(fieldcolor(i)).Foreground(colorTextBlack)
}

const (
	iconDebug    = "🔧"
	iconInfo     = " ✏️"
	iconWarn     = " ⚠️"
	iconErr      = " ❌"
	iconDPanic   = "🔨"
	iconPanic    = " ⛔"
	iconFatal    = "💀"
	iconUnknown  = "❔"
	invalidLevel = "☢️"

	pline = ""
)

var (
	debugStyle  = lipgloss.NewStyle().Background(colorDebug).Foreground(colorTextBlack)
	infoStyle   = lipgloss.NewStyle().Background(colorInfo).Foreground(colorTextBlack)
	warnStyle   = lipgloss.NewStyle().Background(colorWarn).Foreground(colorTextBlack)
	errStyle    = lipgloss.NewStyle().Background(colorErr).Foreground(colorTextBlack)
	dPanicStyle = lipgloss.NewStyle().Background(colorDPanic).Foreground(colorTextWhite)
	panicStyle  = lipgloss.NewStyle().Background(colorPanic).Foreground(colorTextWhite)
	fatalStyle  = lipgloss.NewStyle().Background(colorFatal).Foreground(colorTextWhite)

	unknownStyle = lipgloss.NewStyle().Background(colorUnknown).Foreground(colorTextBlack)
	invalidStyle = lipgloss.NewStyle().Background(colorInvalid).Foreground(colorTextBlack)
)

func invert(s lipgloss.Style) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(s.GetBackground()).Background(s.GetForeground())
}
