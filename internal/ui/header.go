package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

const logoColor string = "red"
const color string = "yellow"
const activeColor string = "navy"

type Header struct {
	*tview.TextView

	crumb     []string
	baseCrumb []string
}

func NewHeader() *Header {
	h := &Header{
		TextView:  tview.NewTextView(),
		crumb:     []string{},
		baseCrumb: []string{"terrUI"},
	}
	h.SetBorderPadding(0, 0, 1, 0)
	h.SetDynamicColors(true)

	h.draw()

	return h
}

func (h *Header) draw() {
	h.Clear()
	var bgColor string
	for i, s := range h.crumb {
		if i == 0 {
			bgColor = logoColor
		} else if i == len(h.crumb)-1 {
			bgColor = activeColor
		} else {
			bgColor = color
		}
		fmt.Fprintf(h, "[%s:%s:b] <%s> [-:%s:-] ",
			"white",
			bgColor,
			strings.Replace(strings.ToLower(s), " ", "", -1),
			"black")
	}
}

func (h *Header) SetCrumb(crumbs []string) *Header {
	h.crumb = append(h.baseCrumb, crumbs...)
	h.draw()

	return h
}
