package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Header struct {
	*tview.TextView

	app          *App
	errorColor   tcell.Color
	errorTimeout int
	previousText string
	titleColor   tcell.Color
	labelText    string
}

func NewHeader(app *App, initialText, labelText string, color tcell.Color, errorTimeout int) *Header {
	h := Header{
		TextView: tview.NewTextView(),

		app:          app,
		errorColor:   tcell.ColorOrangeRed,
		errorTimeout: errorTimeout,
		previousText: initialText,
		labelText:    labelText,
		titleColor:   color,
	}
	h.setTextWithLabel(initialText)
	h.SetBorderPadding(0, 0, 1, 0)
	h.SetTextColor(h.titleColor)

	return &h
}

func (h *Header) ShowText(text string) {
	go h.app.QueueUpdateDraw(func() {
		h.previousText = text
		h.SetTextColor(h.titleColor)
		h.setTextWithLabel(text)
	})
}

func (h *Header) ShowError(text string) {
	go h.app.QueueUpdateDraw(func() {
		h.SetTextColor(h.errorColor)
		h.SetText(fmt.Sprintf("ERROR: %s", text))
	})
	go func() {
		time.Sleep(time.Duration(h.errorTimeout) * time.Second)
		h.ShowText(h.previousText)
	}()
}

func (h *Header) setTextWithLabel(text string) {
	if h.labelText != "" {
		if text == "" {
			text = "-"
		}
		h.SetText(fmt.Sprintf("%s: %s", h.labelText, text))
	} else {
		h.SetText(text)
	}
}
