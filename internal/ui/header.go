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

	breadcrumb []string
}

func NewHeader() *Header {
	h := &Header{
		TextView:   tview.NewTextView(),
		breadcrumb: []string{"terrUI"},
	}
	h.SetBorderPadding(0, 0, 1, 0)
	h.SetDynamicColors(true)

	h.draw()

	return h
}

func (h *Header) draw() {
	h.Clear()
	var bgColor string
	for i, s := range h.breadcrumb {
		if i == 0 {
			bgColor = logoColor
		} else if i == len(h.breadcrumb)-1 {
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

func (h *Header) GoForward(text string) {
	h.breadcrumb = append(h.breadcrumb, text)
	h.draw()
}

func (h *Header) GoBack() {
	h.breadcrumb = h.breadcrumb[:len(h.breadcrumb)-1]
	h.draw()

}
