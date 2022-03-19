package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/renato0307/terrui/internal/client"
	"github.com/rivo/tview"

	"gopkg.in/yaml.v3"
)

type Workspace struct {
	*tview.Flex
	text *tview.TextView

	app       *App
	tfeClient *client.TFEClient
}

type workspaceBaseInfo struct {
	ID               string `yaml:"ID"`
	Name             string `yaml:"Name"`
	Description      string `yaml:"Description"`
	Resources        int    `yaml:"Resource Count"`
	TerraformVersion string `yaml:"Terraform Version"`
	Updated          string `yaml:"Updated"`
	Locked           bool   `yaml:"Locked"`
}

type workspaceLastRun struct {
	By               string `yaml:"By"`
	When             string `yaml:"When"`
	Status           string `yaml:"Status"`
	ResourcesAdded   int    `yaml:"Resources Added"`
	ResourcesUpdated int    `yaml:"Resources Updated"`
	ResourcesDeleted int    `yaml:"Resources Deleted"`
}

func NewWorkspace(app *App) (*Workspace, error) {
	tfeClient, err := client.NewTFEClient()
	if err != nil {
		app.message.ShowError("could not show the workspace")
		return nil, fmt.Errorf("error creating the TFE client: %w", err)
	}

	w := Workspace{
		Flex: tview.NewFlex(),
		text: tview.NewTextView(),

		app:       app,
		tfeClient: tfeClient,
	}

	return &w, nil
}

func (w *Workspace) Load() {

	w.app.QueueUpdateDraw(func() {
		w.app.message.ShowText("> workspace")
		w.text.SetText("loading workspace...").
			SetTextAlign(tview.AlignCenter).
			SetTextColor(tcell.ColorPaleVioletRed)

		w.app.supportedCmds.SetCommands(
			[]SupportedCommand{
				{
					ShortCut: "esc",
					Name:     "workspace list",
				},
				{
					ShortCut: "e",
					Name:     "edit details",
				},
				{
					ShortCut: "p",
					Name:     "start a new plan",
				},
				{
					ShortCut: "v",
					Name:     "manage variables",
				},
				{
					ShortCut: "s",
					Name:     "show state",
				},
			},
		)
	})

	workspace, err := w.tfeClient.ReadWorkspace(w.app.config.Organization, w.app.config.Workspace)
	if err != nil {
		w.app.message.ShowError("could not fetch workspace")

		w.app.QueueUpdateDraw(func() {
			w.text.SetText("😵 error loading workspace: " + err.Error()).
				SetTextAlign(tview.AlignCenter).
				SetTextColor(tcell.ColorPaleVioletRed)
		})
		return
	}

	workspaceBase := workspaceBaseInfo{
		ID:               workspace.ID,
		Name:             workspace.Name,
		Description:      workspace.Description,
		TerraformVersion: workspace.TerraformVersion,
		Resources:        workspace.ResourceCount,
		Updated:          workspace.UpdatedAt.Local().Format(time.RFC3339),
		Locked:           workspace.Locked,
	}

	workspaceLastRun := workspaceLastRun{
		By:               workspace.CurrentRun.CreatedBy.Username,
		When:             workspace.CurrentRun.CreatedAt.Local().Format(time.RFC3339),
		Status:           string(workspace.CurrentRun.Status),
		ResourcesAdded:   workspace.CurrentRun.Plan.ResourceAdditions,
		ResourcesUpdated: workspace.CurrentRun.Plan.ResourceChanges,
		ResourcesDeleted: workspace.CurrentRun.Plan.ResourceDestructions,
	}

	yamlBaseData, _ := yaml.Marshal(workspaceBase)
	yamlLastRunData, _ := yaml.Marshal(workspaceLastRun)

	t1 := tview.NewTextView()
	t2 := tview.NewTextView()
	t3 := tview.NewList()

	w.app.QueueUpdateDraw(func() {
		t1.SetBorder(true)
		t1.SetBorderPadding(0, 1, 1, 1)
		t1.SetTitle(" details ")
		t1.SetText(colorizeYAML(string(yamlBaseData)))
		t1.SetDynamicColors(true)

		t2.SetBorder(true)
		t2.SetBorderPadding(0, 1, 1, 1)
		t2.SetTitle(" last run ")
		t2.SetText(colorizeYAML(string(yamlLastRunData)))
		t2.SetDynamicColors(true)

		t3.SetBorder(true)
		t3.SetBorderPadding(0, 1, 1, 1)
		t3.SetMainTextStyle(tcell.StyleDefault.Bold(true))
		t3.SetSecondaryTextStyle(tcell.StyleDefault.Dim(true))
		t3.SetHighlightFullLine(true)
		t3.SetWrapAround(true)

		flex := tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(t1, 0, 1, false).
				AddItem(t2, 0, 1, false), 0, 1, false).
			AddItem(t3, 0, 2, false)
		flex.SetBorder(false)

		w.Flex = flex
	})
	go w.loadVariables(workspace.ID, t3)

	w.app.SetFocus(w)
	w.SetInputCapture(w.keyboard)
}

func (w *Workspace) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	// nolint:exhaustive
	switch evt.Key() {
	case tcell.KeyEnter, tcell.KeyCtrlE:
		return nil
	}

	return evt
}

func (w *Workspace) loadVariables(workspaceID string, list *tview.List) {
	w.app.QueueUpdateDraw(func() {
		list.SetTitle(" loading variables... ").
			SetTitleColor(tcell.ColorPaleVioletRed)
	})

	vars, err := w.tfeClient.ListWorkspaceVariables(workspaceID)
	if err != nil {
		w.app.message.ShowError("could not fetch variables")
		w.app.QueueUpdateDraw(func() {
			list.SetTitle(fmt.Sprintf(" 😵 error loading variables: %s ", err.Error())).
				SetTitleColor(tcell.ColorPaleVioletRed)
		})
		return
	}

	w.app.QueueUpdateDraw(func() {
		list.SetTitle(fmt.Sprintf(" variables (%d) ", len(vars.Items))).
			SetTitleColor(tview.Styles.PrimaryTextColor)
	})

	w.app.QueueUpdateDraw(func() {
		shortcut := int('a')
		for _, v := range vars.Items {
			value := v.Value
			if v.Sensitive {
				value = "******"
			}
			mainText := fmt.Sprintf("%s = %s", v.Key, value)
			secondaryText := fmt.Sprintf("%s", v.Category)
			if v.Sensitive {
				secondaryText += ", sensitive"
			}

			list.AddItem(mainText, secondaryText, rune(shortcut), nil)

			if shortcut == 'z' {
				shortcut = 'A' - 1
			} else if shortcut == 'Z' {
				shortcut = '0' - 1
			} else if shortcut == '9' {
				shortcut = 0
			}

			if shortcut != 0 {
				shortcut = shortcut + 1
			}
		}
	})

	w.app.SetFocus(list)
}
