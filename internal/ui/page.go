package ui

import "github.com/rivo/tview"

type PageFactory func(*App) Page

type Page interface {
	tview.Primitive

	Load() error
	View() string
	BindKeys() KeyActions
	Crumb() []string
}
