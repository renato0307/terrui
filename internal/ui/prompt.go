package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	defaultPrompt = "%c> [::b]"
)

type PromptListener interface {
	Completed(text string)
	Canceled()
}

type Prompt struct {
	*tview.InputField

	app       *App
	listeners map[string]PromptListener
}

func NewPrompt(app *App, bgColor tcell.Color) *Prompt {
	p := Prompt{
		InputField: tview.NewInputField(),

		app:       app,
		listeners: map[string]PromptListener{},
	}

	p.SetInputCapture(p.keyboard)
	p.SetLabel(fmt.Sprintf(defaultPrompt, '🚀'))
	p.SetFieldBackgroundColor(bgColor)
	return &p
}

func (p *Prompt) AddListener(k string, l PromptListener) {
	p.listeners[k] = l
}

func (p *Prompt) Reset() {
	p.Blur()
	p.SetText("")
}

func (p *Prompt) SetApp(app *App) {
	p.app = app
}

func (p *Prompt) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	// nolint:exhaustive
	switch evt.Key() {
	case tcell.KeyBackspace2, tcell.KeyBackspace, tcell.KeyDelete:
		return evt
	case tcell.KeyRune:
		return evt
	case tcell.KeyEscape:
		p.Reset()
		for _, l := range p.listeners {
			l.Canceled()
		}
	case tcell.KeyEnter, tcell.KeyCtrlE:
		txt := p.GetText()
		p.Reset()
		for _, l := range p.listeners {
			l.Completed(txt)
		}
	case tcell.KeyCtrlW, tcell.KeyCtrlU:
		p.Reset()
	case tcell.KeyUp:
	case tcell.KeyDown:
	case tcell.KeyTab, tcell.KeyRight, tcell.KeyCtrlF:
	}

	return nil
}
