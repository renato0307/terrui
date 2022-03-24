package ui

import (
	"github.com/gdamore/tcell/v2"
)

type (
	// ActionHandler handles a keyboard command.
	ActionHandler func(*tcell.EventKey) *tcell.EventKey

	// KeyAction represents a keyboard action.
	KeyAction struct {
		Description string
		Action      ActionHandler
		Visible     bool
		Shared      bool
	}

	// KeyActions tracks mappings between keystrokes and actions.
	KeyActions map[tcell.Key]KeyAction
)

// NewKeyAction returns a new keyboard action.
func NewKeyAction(d string, a ActionHandler, display bool) KeyAction {
	return KeyAction{Description: d, Action: a, Visible: display}
}

// Add sets up keyboard action listener.
func (a KeyActions) Add(aa KeyActions) {
	for k, v := range aa {
		a[k] = v
	}
}

// Clear remove all actions.
func (a KeyActions) Clear() {
	for k := range a {
		delete(a, k)
	}
}

// Set replace actions with new ones.
func (a KeyActions) Set(aa KeyActions) {
	for k, v := range aa {
		a[k] = v
	}
}

// Delete deletes actions by the given keys.
func (a KeyActions) Delete(kk ...tcell.Key) {
	for _, k := range kk {
		delete(a, k)
	}
}
