package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ListPageSource interface {
	ActionSelectWorkspace(table *tview.Table, currentItem int) func(ek *tcell.EventKey) *tcell.EventKey

	Crumb() []string
	Name() string
	NameList() string

	RenderHeader(table *tview.Table)
	RenderRows(table *tview.Table)

	SupportsSearch() bool
	Search(searchText string, pageNumber int) error

	Empty() bool
	CurrentPage() int
	TotalCount() int
	TotalPages() int
}

type ListPage struct {
	app    *App
	source ListPageSource

	*tview.Flex
	table *tview.Table

	pagination  *tview.TextView
	searchInput *tview.InputField
	searching   bool

	currentItem int
}

func NewListPage(app *App, source ListPageSource) Page {
	l := ListPage{
		app:    app,
		source: source,

		Flex:        tview.NewFlex(),
		table:       tview.NewTable(),
		pagination:  tview.NewTextView(),
		searchInput: tview.NewInputField(),

		currentItem: 1,
	}

	headerFlex := tview.NewFlex()
	headerFlex.AddItem(l.searchInput, 0, 1, false)
	headerFlex.AddItem(l.pagination, 0, 1, false)

	l.searchInput.SetFieldBackgroundColor(l.GetBackgroundColor())
	l.pagination.SetTextAlign(tview.AlignRight)

	l.AddItem(headerFlex, 2, 0, false).
		SetDirection(tview.FlexRow).
		AddItem(l.table, 0, 1, true)

	return &l
}

func (l *ListPage) Load() error {
	return l.source.Search("", -1)
}

func (l *ListPage) View() string {
	l.table.Clear()
	l.table.SetSelectable(true, false)

	l.source.RenderHeader(l.table)
	if l.source.Empty() {
		return fmt.Sprintf("no %s found", l.source.NameList())
	}

	l.source.RenderRows(l.table)

	l.table.SetSelectionChangedFunc(func(row, column int) {
		l.currentItem = row
	})

	l.pagination.SetText(
		fmt.Sprintf("page %d of %d, total %s: %d",
			l.source.CurrentPage(),
			l.source.TotalPages(),
			l.source.NameList(),
			l.source.TotalCount()))

	return "workspaces loaded"
}

func (l *ListPage) BindKeys() KeyActions {
	aa := KeyActions{
		tcell.KeyEnter: NewKeyAction(fmt.Sprintf("select %s", l.source.Name()), l.actionSelectWorkspace, true),
		tcell.KeyCtrlJ: NewKeyAction("next page", l.actionPaginationNextPage, true),
		tcell.KeyCtrlK: NewKeyAction("previous page", l.actionPaginationPrevPage, true),
	}

	if l.source.SupportsSearch() {
		aa.Add(KeyActions{
			KeySlash:     NewKeyAction(fmt.Sprintf("search %s", l.source.NameList()), l.actionSearch, true),
			tcell.KeyEsc: NewKeyAction("cancel search if search results are active", l.actionCancelSearch, true),
		})
	}

	return aa
}

func (l *ListPage) Crumb() []string {
	return l.source.Crumb()
}

func (l *ListPage) Name() string {
	return l.source.NameList()
}

func (l *ListPage) Footer() string {
	return ""
}

func (l *ListPage) actionSelectWorkspace(ek *tcell.EventKey) *tcell.EventKey {
	if l.searching {
		return ek
	}

	return l.source.ActionSelectWorkspace(l.table, l.currentItem)(ek)
}

func (l *ListPage) actionPaginationNextPage(ek *tcell.EventKey) *tcell.EventKey {
	go l.app.ExecPageWithLoadFunc(l, l.loadNextPageFunc(), false)
	return nil
}

func (l *ListPage) actionPaginationPrevPage(ek *tcell.EventKey) *tcell.EventKey {
	go l.app.ExecPageWithLoadFunc(l, l.loadPrevPageFunc(), false)
	return nil
}

func (l *ListPage) actionSearch(ek *tcell.EventKey) *tcell.EventKey {
	l.searching = true

	l.searchInput.SetFieldBackgroundColor(tview.Styles.ContrastBackgroundColor)
	l.searchInput.SetLabel("🔎 ")
	l.searchInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			l.searchInput.SetFieldBackgroundColor(l.GetBackgroundColor())
			l.searchInput.SetLabel("")
			l.searchInput.SetText("")
		case tcell.KeyEnter:
			go l.app.ExecPageWithLoadFunc(l, l.loadSearchFunc(), false)
		}
		l.searching = false
		l.app.SetFocus(l.table)
	})
	l.app.SetFocus(l.searchInput)
	return nil
}

func (l *ListPage) actionCancelSearch(ek *tcell.EventKey) *tcell.EventKey {
	if l.searchInput.GetText() == "" {
		return ek
	}

	l.searchInput.SetFieldBackgroundColor(l.GetBackgroundColor())
	l.searchInput.SetLabel("")
	l.searchInput.SetText("")
	go l.app.ExecPageWithLoadFunc(l, l.loadSearchFunc(), false)

	return nil
}

func (l *ListPage) loadNextPageFunc() func() error {
	return func() error {
		pageToLoad := l.source.CurrentPage() + 1
		if pageToLoad > l.source.TotalPages() {
			pageToLoad = 1
		}
		return l.source.Search(l.searchInput.GetText(), pageToLoad)
	}
}

func (l *ListPage) loadPrevPageFunc() func() error {
	return func() error {
		pageToLoad := l.source.CurrentPage() - 1
		if pageToLoad < 1 {
			pageToLoad = l.source.TotalPages()
		}
		return l.source.Search(l.searchInput.GetText(), pageToLoad)
	}
}

func (l *ListPage) loadSearchFunc() func() error {
	return func() error {
		return l.source.Search(l.searchInput.GetText(), -1)
	}
}
