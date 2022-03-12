package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Header struct {
	*tview.TextView

	app           *App
	errorColor    tcell.Color
	errorTimeout  int
	previousTitle string
	titleColor    tcell.Color
}

func NewHeader(app *App, initialText string, errorTimeout int) *Header {
	h := Header{
		TextView: tview.NewTextView(),

		app:           app,
		errorColor:    tcell.ColorOrangeRed,
		errorTimeout:  errorTimeout,
		previousTitle: initialText,
		titleColor:    tcell.ColorWhiteSmoke,
	}
	h.SetText(initialText)
	h.SetBorderPadding(0, 0, 1, 0)

	return &h
}

func (h *Header) ShowTitle(text string) {
	go h.app.QueueUpdateDraw(func() {
		h.previousTitle = text
		h.SetTextColor(h.titleColor)
		h.SetText(text)
	})
}

func (h *Header) ShowError(text string) {
	go h.app.QueueUpdateDraw(func() {
		h.SetTextColor(h.errorColor)
		h.SetText(fmt.Sprintf("ERROR: %s", text))
	})
	go func() {
		time.Sleep(time.Duration(h.errorTimeout) * time.Second)
		h.ShowTitle(h.previousTitle)
	}()
}
