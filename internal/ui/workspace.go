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
	WorkingDirectory string `yaml:"Working Directory"`
	ExecutionMode    string `yaml:"Execution Mode"`
	AutoApply        bool   `yaml:"Auto Apply"`
}

type workspaceLastRun struct {
	By               string `yaml:"By"`
	When             string `yaml:"When"`
	Status           string `yaml:"Status"`
	ResourcesAdded   int    `yaml:"Resources Added"`
	ResourcesUpdated int    `yaml:"Resources Updated"`
	ResourcesDeleted int    `yaml:"Resources Deleted"`
}

type workspaceMetrics struct {
	ApplyDurationAverage string `yaml:"Average Apply Duration"`
	PlanDurationAverage  string `yaml:"Average Plan Duration"`
	RunFailures          int    `yaml:"Total Failed Runs"`
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
					ShortCut: "?",
					Name:     "help",
				},
				{
					ShortCut: ":",
					Name:     "execute commands",
				},
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
					ShortCut: "s",
					Name:     "show state",
				},
				{
					ShortCut: "v",
					Name:     "manage variables",
				},
				{
					ShortCut: "t",
					Name:     "manage tags",
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
		WorkingDirectory: workspace.WorkingDirectory,
		ExecutionMode:    workspace.ExecutionMode,
		AutoApply:        workspace.AutoApply,
	}

	workspaceLastRun := workspaceLastRun{
		By:               workspace.CurrentRun.CreatedBy.Username,
		When:             workspace.CurrentRun.CreatedAt.Local().Format(time.RFC3339),
		Status:           string(workspace.CurrentRun.Status),
		ResourcesAdded:   workspace.CurrentRun.Plan.ResourceAdditions,
		ResourcesUpdated: workspace.CurrentRun.Plan.ResourceChanges,
		ResourcesDeleted: workspace.CurrentRun.Plan.ResourceDestructions,
	}

	workspaceMetrics := workspaceMetrics{
		ApplyDurationAverage: workspace.ApplyDurationAverage.String(),
		PlanDurationAverage:  workspace.PlanDurationAverage.String(),
		RunFailures:          workspace.RunFailures,
	}

	yamlBaseData, _ := yaml.Marshal(workspaceBase)
	yamlLastRunData, _ := yaml.Marshal(workspaceLastRun)
	yamlMetrics, _ := yaml.Marshal(workspaceMetrics)

	details := tview.NewTextView()
	metrics := tview.NewTextView()
	tags := tview.NewList()
	lastRun := tview.NewTextView()
	variables := tview.NewList()

	w.app.QueueUpdateDraw(func() {
		details.SetBorder(true)
		details.SetBorderPadding(0, 1, 1, 1)
		details.SetTitle(" details ")
		details.SetText(colorizeYAML(string(yamlBaseData)))
		details.SetDynamicColors(true)

		metrics.SetBorder(true)
		metrics.SetBorderPadding(0, 1, 1, 1)
		metrics.SetTitle(fmt.Sprintf(" metrics (last %d runs) ", workspace.RunsCount))
		metrics.SetText(colorizeYAML(string(yamlMetrics)))
		metrics.SetDynamicColors(true)

		tags.SetBorder(true)
		tags.SetBorderPadding(0, 1, 1, 1)
		tags.SetTitle(" tags ")
		tags.SetMainTextStyle(tcell.StyleDefault.Bold(true))
		tags.SetSecondaryTextStyle(tcell.StyleDefault.Dim(true))
		tags.SetWrapAround(true)
		tags.SetSelectedStyle(tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(tview.Styles.PrimitiveBackgroundColor))
		tags.ShowSecondaryText(false)
		for _, t := range workspace.TagNames {
			tags.AddItem(t, "", 0, nil)
		}

		lastRun.SetBorder(true)
		lastRun.SetBorderPadding(0, 1, 1, 1)
		lastRun.SetTitle(" last run ")
		lastRun.SetText(colorizeYAML(string(yamlLastRunData)))
		lastRun.SetDynamicColors(true)

		variables.SetBorder(true)
		variables.SetBorderPadding(0, 1, 1, 1)
		variables.SetMainTextStyle(tcell.StyleDefault.Bold(true))
		variables.SetSecondaryTextStyle(tcell.StyleDefault.Dim(true))
		variables.SetHighlightFullLine(true)
		variables.SetWrapAround(true)

		flex := tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(details, 0, 3, false).
				AddItem(lastRun, 0, 1, false).
				AddItem(metrics, 0, 1, false).
				AddItem(tags, 0, 1, false), 0, 1, false)

		if w.app.config.WorkspaceShowVariables {
			flex.AddItem(variables, 0, 2, false)
		}

		flex.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			details.ScrollToBeginning()
			lastRun.ScrollToBeginning()
			return x, y, width, height
		})

		w.Flex = flex
	})
	go w.loadVariables(workspace.ID, variables, true)

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

func (w *Workspace) loadVariables(workspaceID string, list *tview.List, showShortcuts bool) {
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
		shortcut := int('0')
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

			if !showShortcuts {
				shortcut = 0
			}
			list.AddItem(mainText, secondaryText, rune(shortcut), nil)

			if shortcut == '9' {
				shortcut = 'A' - 1
			} else if shortcut == 'Z' {
				shortcut = '.'
			}

			if shortcut != '.' {
				shortcut = shortcut + 1
			}
		}
	})

	w.app.SetFocus(list)
}
