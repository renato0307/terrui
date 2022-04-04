package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Footer struct {
	*tview.TextView

	app          *App
	errorColor   tcell.Color
	errorTimeout int
	previousText string
	titleColor   tcell.Color
}

func NewFooter(app *App, initialText string, color tcell.Color, errorTimeout int) *Footer {
	f := Footer{
		TextView: tview.NewTextView(),

		app:          app,
		errorColor:   tcell.ColorOrangeRed,
		errorTimeout: errorTimeout,
		previousText: initialText,
		titleColor:   color,
	}
	f.SetBorderPadding(0, 0, 1, 0)
	f.SetTextColor(f.titleColor)

	f.SetText(initialText)

	return &f
}

func (f *Footer) ShowText(text string) {
	go f.app.QueueUpdateDraw(func() {
		f.previousText = text
		f.SetTextColor(f.titleColor)
		f.SetText(text)
	})
}

func (f *Footer) ShowError(text string) {
	f.Show(fmt.Sprintf("ERROR: %s", text), f.errorColor)
}

func (f *Footer) Show(text string, color tcell.Color) {
	go f.app.QueueUpdateDraw(func() {
		f.SetTextColor(color)
		f.SetText(text)
	})
	go func() {
		time.Sleep(time.Duration(f.errorTimeout) * time.Second)
		f.ShowText(f.previousText)
	}()
}
